package apigen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"strings"
)

type methodAPI struct {
	URL        string `json:"url,omitempty"`
	HTTPMethod string `json:"method,omitempty"`
	Auth       bool   `json:"auth,omitempty"`
}

type argType struct {
	name      string
	isPointer bool
}

type serviceMethod struct {
	name   string
	recv   argType
	params argType
	result argType
	*methodAPI
	pos token.Pos
}

type serviceMethodCollection struct {
	items       map[string][]*serviceMethod
	methodCount int
}

func (h *serviceMethodCollection) add(m *serviceMethod) {
	if h.items == nil {
		h.items = map[string][]*serviceMethod{}
	}
	if m != nil {
		h.items[m.recv.name] = append(h.items[m.recv.name], m)
		h.methodCount++
	}
}

type kind = reflect.Kind

const (
	Int     = reflect.Int
	String  = reflect.String
	Float64 = reflect.Float64
	Float32 = reflect.Float32
)

type paramStructField struct {
	name string
	kind kind
	*validator
	pos token.Pos
}

func (p paramStructField) apiParamName() string {
	if p.validator == nil || p.validator.paramName == "" {
		return strings.ToLower(p.name)
	}
	return p.validator.paramName
}

type paramStructFieldCollection struct {
	items      map[string][]*paramStructField
	fieldCount int
}

// adds a field to a param struct. If p is nil and struct not exists,
// then a empty named struct is added to the collection.
func (ps *paramStructFieldCollection) add(structName string, p *paramStructField) {
	if ps.items == nil {
		ps.items = map[string][]*paramStructField{}
	}
	if p != nil {
		ps.items[structName] = append(ps.items[structName], p)
		ps.fieldCount++
	} else {
		ps.items[structName] = nil
	}
}

func (ps *paramStructFieldCollection) contains(structName string) bool {
	_, ok := ps.items[structName]
	return ok
}

type ParseError struct {
	Pos token.Pos
	Err error
}

func (e *ParseError) Error() string {
	return e.Err.Error()
}

type GenConfig struct {
	packageName string
	servs       serviceMethodCollection
	params      paramStructFieldCollection
}

func ParseFiles(files map[string]*ast.File) (cfg GenConfig, err error) {
	const op = "ParseFiles"

	// XXX
	defer func() {
		if p := recover(); p != nil {
			if p, ok := p.(*ParseError); ok {
				err = p
			}
			panic(p)
		}
	}()

	for _, f := range files {
		if cfg.packageName == "" {
			cfg.packageName = f.Name.Name
		} else if f.Name.Name != cfg.packageName {
			return cfg, &ParseError{
				Err: fmt.Errorf("different package name %s, want %s", f.Name.Name, cfg.packageName),
				Pos: f.Name.Pos(),
			}
		}
	}

	for _, f := range files {
		err := findServiceMethods(f, &cfg.servs)
		if err != nil {
			return cfg, err
		}
	}

	log.Printf("%s: FOUND %d/%d service/methods", op, len(cfg.servs.items), cfg.servs.methodCount)

	for _, methods := range cfg.servs.items {
		for _, m := range methods {
			cfg.params.add(m.params.name, nil)
		}
	}

	for _, f := range files {
		if err := findParamStructFields(f, &cfg.params); err != nil {
			return cfg, err
		}
	}

	for t, p := range cfg.params.items {
		if p == nil {
			return cfg, fmt.Errorf("%s: NOT FOUND %s param struct", op, t)
		}
	}

	log.Printf("%s: FOUND %d/%d param struct/fields", op, len(cfg.params.items), cfg.params.fieldCount)

	return cfg, nil
}

func findServiceMethods(f *ast.File, servs *serviceMethodCollection) error {
	const op = "findServiceMethods"

	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcName := funcDecl.Name.Name

		api, err := getMethodApi(funcDecl)
		if err != nil {
			return err
		}
		if api == nil {
			log.Printf("%s: SKIP %s func doesnt have apigen:api mark", op, funcName)
			continue
		}

		recv := funcDecl.Recv
		if recv == nil {
			return &ParseError{
				Err: fmt.Errorf("%s: method must have receiver", funcName),
				Pos: funcDecl.Pos(),
			}
		}

		params := funcDecl.Type.Params
		if params == nil || len(params.List) != 2 { // (ctx, params)
			return &ParseError{
				Err: fmt.Errorf("%s: method must have two parameters (ctx, params)", funcName),
				Pos: params.Pos(),
			}
		}

		results := funcDecl.Type.Results
		if results == nil || len(results.List) != 2 { // (result, err)
			return &ParseError{
				Err: fmt.Errorf("%s: method must have two results (result, err)", funcName),
				Pos: results.Pos(),
			}
		}

		m := serviceMethod{
			name:      funcName,
			recv:      getArgType(funcDecl.Recv.List[0].Type),
			params:    getArgType(params.List[1].Type),
			result:    getArgType(results.List[0].Type),
			methodAPI: api,
			pos:       funcDecl.Pos(),
		}

		log.Printf("%s: FOUND %s.%s method", op, m.recv.name, m.name)
		servs.add(&m)
	}

	return nil
}

// returns nil if not marked with comment `// apigen:api`
func getMethodApi(funcDecl *ast.FuncDecl) (*methodAPI, error) {
	if funcDecl.Doc == nil {
		return nil, nil
	}

	for _, comment := range funcDecl.Doc.List {
		if strings.HasPrefix(comment.Text, "// apigen:api") {
			api := methodAPI{HTTPMethod: anyHTTPMethod}
			err := json.Unmarshal([]byte(strings.TrimPrefix(comment.Text, "// apigen:api")), &api)
			if err != nil {
				return nil, &ParseError{
					Err: fmt.Errorf("apigen:api: %w", err),
					Pos: comment.Pos(),
				}
			}
			api.HTTPMethod = strings.ToUpper(api.HTTPMethod)
			return &api, nil
		}
	}

	return nil, nil
}

func getArgType(t ast.Expr) argType {
	switch t := t.(type) {
	case *ast.StarExpr:
		return argType{name: t.X.(*ast.Ident).Name, isPointer: true}
	case *ast.Ident:
		return argType{name: t.Name}
	}
	panic(&ParseError{
		Err: fmt.Errorf("type must be *ats.Ident or *ast.StarExpr, got %T", t),
		Pos: t.Pos(),
	})
}

func findParamStructFields(f *ast.File, params *paramStructFieldCollection) error {
	const op = "findParamStructFields"

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
			if !params.contains(typeName) {
				log.Printf("%s: SKIP %s is unknown type", op, typeName)
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return &ParseError{
					Err: fmt.Errorf("%s: params must be struct", typeName),
					Pos: typeSpec.Type.Pos(),
				}
			}

			log.Printf("%s: FOUND %s struct", op, typeName)

			for _, field := range structType.Fields.List {
				validator, err := getApiValidator(field)
				if err != nil {
					return &ParseError{Err: err, Pos: field.Pos()}
				}
				if validator == nil {
					continue
				}

				fieldName := field.Names[0].Name
				fieldType := field.Type.(*ast.Ident).Name

				var fieldKind kind
				switch fieldType {
				case "string":
					fieldKind = String
				case "int":
					fieldKind = Int
				case "float64":
					fieldKind = Float64
				default:
					return &ParseError{
						Err: fmt.Errorf("%s: field type must be int, string or float64", fieldName),
						Pos: field.Type.Pos(),
					}
				}

				params.add(typeName, &paramStructField{
					name:      fieldName,
					kind:      fieldKind,
					validator: validator,
					pos:       field.Pos(),
				})
			}
		}
	}

	return nil
}

func getApiValidator(field *ast.Field) (*validator, error) {
	if field.Tag == nil {
		return &validator{}, nil
	}
	tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
	tagVal, ok := tag.Lookup("apivalidator")
	if !ok || tagVal == "" {
		return &validator{}, nil
	}
	if tagVal == "-" {
		return nil, nil
	}
	return parseValidator(tagVal)
}
