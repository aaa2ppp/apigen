package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const anyHTTPMethod = ""
const authKey = "100500"

var imports = []string{"bytes", "encoding/json", "errors", "log", "net/http", "strconv"}

type api struct {
	Url    string `json:"url,omitempty"`
	Auth   bool   `json:"auth,omitempty"`
	Method string `json:"method,omitempty"`
}

type method struct {
	name string
	*api
	recvType         string
	paramsType       string
	resultType       string
	recvIsStarExpr   bool
	paramsIsStarExpr bool
	resultIsStarExpr bool
}

type kind = reflect.Kind

const (
	Int     = reflect.Int
	String  = reflect.String
	Float64 = reflect.Float64
	Float32 = reflect.Float32
)

type ruleSet int

const (
	requiredRule ruleSet = 1 << iota
	defaultRule
	enumRule
	minRule
	maxRule
	greaterRule
	lessRule
)

type apiValidator struct {
	rules      ruleSet
	paramName  string
	defaultVal string
	enum       []string
	min        string
	max        string
	greater    string
	less       string
}

func parseApiValidator(s string) (*apiValidator, error) {
	var v apiValidator
	var err error

	for _, entry := range strings.Split(s, ",") {
		switch {
		case strings.HasPrefix(entry, "paramname="):
			v.paramName = strings.TrimPrefix(entry, "paramname=")
		case entry == "required":
			v.rules |= requiredRule
		case strings.HasPrefix(entry, "default="):
			v.rules |= defaultRule
			v.defaultVal = strings.TrimPrefix(entry, "default=")
		case strings.HasPrefix(entry, "enum="):
			v.rules |= enumRule
			v.enum = strings.Split(strings.TrimPrefix(entry, "enum="), "|")
		case strings.HasPrefix(entry, "min="):
			v.rules |= minRule
			v.min = strings.TrimPrefix(entry, "min=")
		case strings.HasPrefix(entry, "max="):
			v.rules |= maxRule
			v.max = strings.TrimPrefix(entry, "max=")
		case strings.HasPrefix(entry, ">="):
			v.rules |= minRule
			v.min = strings.TrimPrefix(entry, ">=")
		case strings.HasPrefix(entry, "<="):
			v.rules |= maxRule
			v.max = strings.TrimPrefix(entry, "<=")
		case strings.HasPrefix(entry, ">"):
			v.rules |= greaterRule
			v.greater = strings.TrimPrefix(entry, ">")
		case strings.HasPrefix(entry, "<"):
			v.rules |= lessRule
			v.less = strings.TrimPrefix(entry, "<")
		default:
			err = fmt.Errorf("%s: unknown rule", entry)
		}

		if err != nil {
			return nil, err
		}
	}

	return &v, nil
}

type param struct {
	name      string
	paramType kind
	*apiValidator
}

func (p param) apiParamName() string {
	if name := p.apiValidator.paramName; name != "" {
		return name
	}
	return strings.ToLower(p.name)
}

type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(f string, a ...interface{}) error {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, f, a...)
	}
	if p.err == nil && (len(f) == 0 || f[len(f)-1] != '\n') {
		_, p.err = p.w.Write([]byte{'\n'})
	}
	return p.err
}

type parserError struct {
	pos token.Pos
	err error
}

func (e parserError) Error() string {
	return e.err.Error()
}

func main() {
	const op = "main"

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <input_file> [<output_file>]", filepath.Base(os.Args[0]))
	}

	inputFile := os.Args[1]
	outputFile := ""

	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	handlerSet, err := findHandlers(f)
	if err != nil {
		if err, ok := err.(*parserError); ok {
			log.Fatalf("%v: %v", fset.Position(err.pos), err)
		}
		log.Fatal(err)
	}

	log.Printf("%s: FOUND %d handlers", op, len(handlerSet))

	paramsSet := map[string][]*param{}
	for _, methods := range handlerSet {
		for _, m := range methods {
			paramsSet[m.paramsType] = nil
		}
	}

	if err := findParamsStructs(f, paramsSet); err != nil {
		if err, ok := err.(*parserError); ok {
			log.Fatalf("%v: %v", fset.Position(err.pos), err)
		}
		log.Fatal(err)
	}

	{
		count := 0
		for t, p := range paramsSet {
			if p == nil {
				log.Printf("%s: NOT FOUND %s params struct", op, t)
				continue
			}
			count++
		}
		log.Printf("%s: FOUND %d/%d params structs", op, count, len(paramsSet))
	}

	buf := &bytes.Buffer{}
	p := &printer{w: buf}

	genOpts := genCodeOpts{
		inputFile:   inputFile,
		packageName: f.Name.Name,
		handlerSet:  handlerSet,
		paramsSet:   paramsSet,
	}
	if err := genCode(p, genOpts); err != nil {
		log.Fatal(err)
	}

	if outputFile == "" {
		if _, err := buf.WriteTo(os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := os.WriteFile(outputFile, buf.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}
}

func findHandlers(f *ast.File) (map[string][]*method, error) {
	const op = "findHandlers"

	handlers := map[string][]*method{}

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcName := funcDecl.Name.Name

		api, err := getMethodApi(funcDecl)
		if err != nil {
			return nil, err
		}
		if api == nil {
			log.Printf("%s: SKIP func %#v doesnt have apigen:api mark", op, funcName)
			continue
		}

		recv := funcDecl.Recv
		if recv == nil {
			return nil, &parserError{
				pos: funcDecl.Pos(),
				err: fmt.Errorf("%s: method must have receiver", funcName),
			}
		}

		params := funcDecl.Type.Params
		if params == nil || len(params.List) != 2 { // (ctx, params)
			return nil, &parserError{
				pos: funcDecl.Pos(),
				err: fmt.Errorf("%s: method must have two parameters", funcName),
			}
		}

		results := funcDecl.Type.Results
		if results == nil || len(results.List) != 2 { // (result, err)
			return nil, &parserError{
				pos: funcDecl.Pos(),
				err: fmt.Errorf("%s: method must have two results", funcName),
			}
		}

		m := method{name: funcName, api: api}

		m.recvType, m.recvIsStarExpr = parseType(funcDecl.Recv.List[0].Type)
		m.paramsType, m.paramsIsStarExpr = parseType(params.List[1].Type)
		m.resultType, m.resultIsStarExpr = parseType(results.List[0].Type)

		log.Printf("%s: FOUND %s.%s", op, m.recvType, m.name)

		handlers[m.recvType] = append(handlers[m.recvType], &m)
	}

	return handlers, nil
}

// returns nil if not marked with comment `// apigen:api`
func getMethodApi(funcDecl *ast.FuncDecl) (*api, error) {
	if funcDecl.Doc == nil {
		return nil, nil
	}

	for _, comment := range funcDecl.Doc.List {
		if strings.HasPrefix(comment.Text, "// apigen:api") {
			api := api{Method: anyHTTPMethod}
			err := json.Unmarshal([]byte(strings.TrimPrefix(comment.Text, "// apigen:api")), &api)
			if err != nil {
				return nil, &parserError{
					pos: comment.Pos(),
					err: fmt.Errorf("apigen:api: %w", err),
				}
			}
			return &api, nil
		}
	}

	return nil, nil
}

func parseType(t ast.Expr) (string, bool) {
	switch t := t.(type) {
	case *ast.StarExpr:
		return t.X.(*ast.Ident).Name, true
	case *ast.Ident:
		return t.Name, false
	}
	panic("type must be *ats.Ident or *ast.StarExpr")
}

func findParamsStructs(f *ast.File, params map[string][]*param) error {
	const op = "findParamsStructs"

	for _, decl := range f.Decls {
		g, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range g.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeName := typeSpec.Name.Name
			if _, ok := params[typeName]; !ok {
				log.Printf("%s: SKIP %s is unknown type", op, typeName)
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return &parserError{
					pos: structType.Pos(),
					err: fmt.Errorf("%s: method parameters must be struct", typeName),
				}
			}

			log.Printf("%s: FOUND %s", op, typeName)

			for _, field := range structType.Fields.List {
				validator, err := getApiValidator(field)
				if err != nil {
					return &parserError{
						pos: field.Pos(),
						err: err,
					}
				}

				fieldName := field.Names[0].Name
				fieldType := field.Type.(*ast.Ident).Name

				var paramType kind
				switch fieldType {
				case "string":
					paramType = String
				case "int":
					paramType = Int
				case "float64":
					paramType = Float64
				default:
					return &parserError{
						pos: typeSpec.Pos(),
						err: fmt.Errorf("%s: field type must be int, string or float64", fieldName),
					}
				}

				params[typeName] = append(params[typeName], &param{
					name:         fieldName,
					paramType:    paramType,
					apiValidator: validator,
				})
			}
		}
	}

	return nil
}

// returns null if there is no apvalidator tag
func getApiValidator(field *ast.Field) (*apiValidator, error) {
	if field.Tag == nil {
		return nil, nil
	}

	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
	tagVal := tag.Get("apivalidator")

	if tagVal == "-" {
		return nil, nil
	}

	return parseApiValidator(tagVal)
}

type genCodeOpts struct {
	inputFile   string
	packageName string
	handlerSet  map[string][]*method
	paramsSet   map[string][]*param
	// TODO resultSet map[string][]*result
}

func genCode(p *printer, o genCodeOpts) error {
	p.printf("// !!! Do not change this code !!!")
	p.printf("// The code is generated automatically by apigen from %s", o.inputFile)
	p.printf("package %s", o.packageName)
	p.printf("")
	p.printf("import (")
	p.printf("\"%s\"", strings.Join(imports, "\"\n\""))
	p.printf(")")

	if err := genWriteApiError(p); err != nil {
		return err
	}

	var order []string
	for k := range o.handlerSet {
		order = append(order, k)
	}
	sort.Strings(order)

	for _, recvType := range order {
		methods := o.handlerSet[recvType]

		if err := genServeHTTPMethod(p, recvType, methods); err != nil {
			return err
		}
		for _, method := range methods {
			if err := genMethodWrapper(p, method); err != nil {
				return err
			}
		}
	}

	order = order[:0]
	for k := range o.paramsSet {
		order = append(order, k)
	}
	sort.Strings(order)

	for _, paramsType := range order {
		params := o.paramsSet[paramsType]

		if err := genGetFromRequest(p, paramsType, params); err != nil {
			return err
		}
		if err := genValidate(p, paramsType, params); err != nil {
			return err
		}
	}

	return nil
}

func genGetFromRequest(p *printer, paramsType string, params []*param) error {
	const op = "genGetFromRequest"

	p.printf("")
	p.printf("func (p *%s) getFromRequest(r *http.Request) error {", paramsType)

	for _, param := range params {
		p.printf("{")
		p.printf("s := r.FormValue(%q)", param.apiParamName())

		switch param.paramType {
		case String:
			p.printf("p.%s = s", param.name)
		case Int:
			p.printf("v, err := strconv.Atoi(s)")
			p.printf("if err != nil { return errors.New(\"%s must be int\") }", param.apiParamName())
			p.printf("p.%s = v", param.name)
		case Float32:
			p.printf("v, err := strconv.ParseFloat(s, 32)")
			p.printf("if err != nil { return errors.New(\"%s must be float32\") }", param.apiParamName())
			p.printf("p.%s = float32(v)", param.name)
		case Float64:
			p.printf("v, err := strconv.ParseFloat(s, 64)")
			p.printf("if err != nil { return errors.New(\"%s must be float64\") }", param.apiParamName())
			p.printf("p.%s = v", param.name)
		default:
			return fmt.Errorf("%s: %s.%s: invalid param type: %v", op, paramsType, param.paramName, param.paramType)
		}

		p.printf("}")
	}

	p.printf("return nil")
	p.printf("}")

	return p.err
}

func genValidate(p *printer, paramsType string, params []*param) error {
	const op = "genValidate"

	p.printf("")
	p.printf("func (p *%s) validate() error {", paramsType)

	for _, param := range params {
		if param.apiValidator == nil || param.rules == 0 {
			continue
		}

		p.printf("{")

		if param.rules&requiredRule != 0 {
			switch param.paramType {
			case String:
				p.printf("if p.%s == \"\" { return errors.New(\"%s must be not empty\") }", param.name, param.apiParamName())
			case Int, Float32, Float64:
				p.printf("if p.%s == 0 { return errors.New(\"%s must be not empty\") }", param.name, param.apiParamName())
			default:
				return fmt.Errorf("%s: %s.%s: required rule not applicable for %v type", op, paramsType, param.name, param.paramType)
			}
		}

		if param.rules&defaultRule != 0 {
			switch param.paramType {
			case String:
				p.printf("if p.%s == \"\" { p.%s = %q }", param.name, param.name, param.defaultVal)
			case Int, Float32, Float64:
				p.printf("if p.%s == 0 { p.%s = %s }", param.name, param.name, param.defaultVal)
			default:
				return fmt.Errorf("%s: %s.%s: default rule not applicable for %v type", op, paramsType, param.name, param.paramType)
			}
		}

		if param.rules&enumRule != 0 {
			p.printf("valid := false")
			switch param.paramType {
			case String:
				for _, s := range param.enum {
					p.printf("valid = valid || p.%s == %q", param.name, s)
				}
				p.printf("if !valid { return errors.New(\"%s must be one of [%s]\") }", param.apiParamName(), strings.Join(param.enum, ", "))
			case Int, Float32, Float64:
				for _, s := range param.enum {
					p.printf("valid |= p.%s == %s", param.name, s)
				}
				p.printf("if !valid { return errors.New(\"%s must be one of [%s]\") }", param.apiParamName(), strings.Join(param.enum, ", "))
			default:
				return fmt.Errorf("%s: %s.%s: enum rule not applicable for %v type", op, paramsType, param.name, param.paramType)
			}
		}

		if param.rules&minRule != 0 {
			switch param.paramType {
			case String:
				p.printf("if len(p.%s) < %d { return errors.New(\"%s len must be >= %d\") }", param.name, param.min, param.apiParamName(), param.min)
			case Int, Float32, Float64:
				p.printf("if p.%s < %d { return errors.New(\"%s must be >= %d\") }", param.name, param.min, param.apiParamName(), param.min)
			default:
				return fmt.Errorf("%s: %s.%s: min rule not applicable for %v type", op, paramsType, param.name, param.paramType)
			}
		}

		if param.rules&maxRule != 0 {
			switch param.paramType {
			case Int, Float32, Float64:
				p.printf("if p.%s > %d { return errors.New(\"%s must be <= %d\") }", param.name, param.max, param.apiParamName(), param.max)
			default:
				return fmt.Errorf("%s: %s.%s: max rule not applicable for %v type", op, paramsType, param.name, param.paramType)
			}
		}

		p.printf("}")
	}

	p.printf("return nil")
	p.printf("}")

	return p.err
}

func genServeHTTPMethod(p *printer, recvType string, methods []*method) error {

	// func (h *SomeStructName ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 	switch r.URL.Path {
	// 	case "...":
	// 		h.wrapperDoSomeJob(w, r)
	// 	default:
	// 		// 404
	// 	}
	// }

	p.printf("")
	p.printf("func (h *%s) ServeHTTP(w http.ResponseWriter, r *http.Request) {", recvType)
	p.printf("switch r.URL.Path {")

	for _, method := range methods {
		p.printf("case \"%s\":", method.Url)

		if method.Method != anyHTTPMethod {
			p.printf("if r.Method != \"%s\" {", method.Method)
			p.printf("	writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New(\"bad method\")})")
			p.printf("	return")
			p.printf("}")
		}

		if method.Auth {
			p.printf("if key := r.Header.Get(\"X-Auth\"); key != \"%s\" {", authKey)
			p.printf("	writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New(\"unauthorized\")})")
			p.printf("	return")
			p.printf("}")
		}

		p.printf("	h.wrapper%s(w, r)", method.name)
	}

	p.printf("default:")
	p.printf("	writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New(\"unknown method\")})")
	p.printf("}")
	p.printf("}")

	return p.err
}

func genMethodWrapper(p *printer, method *method) error {

	// func (h *SomeStructName ) wrapperDoSomeJob() {
	// 	// заполнение структуры params
	// 	// валидирование параметров
	// 	res, err := h.DoSomeJob(ctx, params)
	// 	// прочие обработки
	// }

	p.printf("")
	if method.recvIsStarExpr {
		p.printf("func (h *%s) wrapper%s(w http.ResponseWriter, r *http.Request) {", method.recvType, method.name)
	} else {
		p.printf("func (h %s) wrapper%s(w http.ResponseWriter, r *http.Request) {", method.recvType, method.name)
	}
	p.printf("const op = \"%s.wrapper%s\"", method.recvType, method.name)
	p.printf("var params %s", method.paramsType)

	p.printf("if err := params.getFromRequest(r); err != nil {")
	p.printf("	writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})")
	p.printf("	return")
	p.printf("}")

	p.printf("if err := params.validate(); err != nil {")
	p.printf("	writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})")
	p.printf("	return")
	p.printf("}")

	p.printf("ctx := r.Context()") // TODO: different context?
	p.printf("res, err := h.%s(ctx, params)", method.name)
	p.printf("if err != nil {")
	p.printf("	switch err := err.(type) {")
	p.printf("	case *ApiError:")
	p.printf("		writeApiError(w, *err)")
	p.printf("	case ApiError:")
	p.printf("		writeApiError(w, err)")
	p.printf("	default:")
	p.printf("		writeApiError(w, ApiError{HTTPStatus: http.StatusInternalServerError, Err: err})")
	p.printf("	}")
	p.printf("	return")
	p.printf("}")

	p.printf("resp := struct {")
	p.printf("	Response *%s    `json:\"response\"`", method.resultType)
	p.printf("	Error    string `json:\"error\"`")
	p.printf("}{")
	if method.resultIsStarExpr {
		p.printf("	Response: res,")
	} else {
		p.printf("	Response: &res,")
	}
	p.printf("}")

	p.printf("w.Header().Add(\"content-type\", \"application/json\")")
	p.printf("w.WriteHeader(http.StatusOK)")

	// TODO: codegen for json marshaling
	p.printf("if err := json.NewEncoder(w).Encode(&resp); err != nil {")
	p.printf("	log.Printf(\"%%s: can't write response body: %%v\", op, err)")
	p.printf("}")

	p.printf("}")

	return p.err
}

func genWriteApiError(p *printer) error {
	p.printf("")
	p.printf("func writeApiError(w http.ResponseWriter, ae ApiError) {")
	p.printf("const op = \"writeApiError\"")

	p.printf("var resp bytes.Buffer")
	p.printf("resp.WriteByte('{')")
	p.printf("resp.WriteString(`\"error\":`)")
	p.printf("resp.WriteString(strconv.Quote(ae.Error()))")
	p.printf("resp.WriteByte('}')")

	p.printf("w.Header().Add(\"content-type\", \"application/json\")")
	p.printf("w.WriteHeader(ae.HTTPStatus)")

	p.printf("if _, err := resp.WriteTo(w); err != nil {")
	p.printf("	log.Printf(\"%%s: can't write response body: %%v\", op, err)")
	p.printf("}")

	p.printf("}")
	return p.err
}
