package apigen

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
)

const (
	anyHTTPMethod = "*"
	authKey       = "100500"
	q             = "`"
)

var imports = []string{"encoding/json", "errors", "io", "fmt", "log", "net/http", "strconv", "strings"}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func GenCode(w io.Writer, cfg GenConfig) error {
	const op = "GenCode"

	p := newPrinter(w)
	p.printf(`// !!! Do not change this code !!!`)
	p.printf(`// The code is generated automatically by apigen tool`)
	p.printf(`package %s`, cfg.packageName)
	p.printf(`import ("%s")`, strings.Join(imports, "\";\""))

	if err := genWriteApiError(p); err != nil {
		return err
	}

	order := sortedKeys(cfg.servs.items)
	log.Printf("%s: generate methods for services: %v", op, strings.Join(order, ", "))

	for _, servName := range order {
		methods := cfg.servs.items[servName]

		if err := genServeHTTP(p, servName, methods); err != nil {
			return err
		}
		for _, method := range methods {
			if err := genMethodWrapper(p, method); err != nil {
				return err
			}
		}
	}

	order = sortedKeys(cfg.params.items)
	log.Printf("%s: generate methods for param struct: %v", op, strings.Join(order, ", "))

	for _, structName := range order {
		params := cfg.params.items[structName]

		if err := genGetFromRequest(p, structName, params); err != nil {
			return err
		}
		if err := genValidate(p, structName, params); err != nil {
			return err
		}
	}

	return nil
}

func genWriteApiError(p *printer) error {
	p.printf(``)
	p.printf(`func writeApiError(w http.ResponseWriter, ae ApiError) {`)
	p.printf(`const op = "writeApiError"`)

	p.printf(`w.Header().Add("content-type", "application/json")`)
	p.printf(`w.WriteHeader(ae.HTTPStatus)`)

	p.printf(`if _, err := fmt.Fprintf(w, "{\"error\":%%q}", ae.Err.Error()); err != nil {`)
	p.printf(`	log.Printf("%%s: can't write response body: %%v", op, err)`)
	p.printf(`}`)

	p.printf(`}`)
	return p.err
}

func genServeHTTP(p *printer, recvType string, methods []*serviceMethod) error {

	// func (h *SomeStructName ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 	switch r.URL.Path {
	// 	case "...":
	// 		h.wrapperDoSomeJob(w, r)
	// 	default:
	// 		// 404
	// 	}
	// }

	byPath := make(map[string]map[string]*serviceMethod)
	for _, m := range methods {
		byMethod, ok := byPath[m.URL]
		if !ok {
			byMethod = map[string]*serviceMethod{}
			byPath[m.URL] = byMethod
		}
		if _, ok := byMethod[m.HTTPMethod]; ok {
			return &ParseError{
				Err: errors.New("dublicate HTTP method"),
				Pos: m.pos,
			}
		}
		byMethod[m.HTTPMethod] = m
	}

	p.printf(``)
	p.printf(`func (h *%s) ServeHTTP(w http.ResponseWriter, r *http.Request) {`, recvType)
	p.printf(`switch r.URL.Path {`)

	for _, path := range sortedKeys(byPath) {
		byMethod := byPath[path]
		p.printf(`case "%s":`, path)

		p.printf(`switch /*r.Method*/ {`)
		for _, httpMethod := range sortedKeys(byPath[path]) {
			if httpMethod == anyHTTPMethod {
				continue
			}
			method := byMethod[httpMethod]
			p.printf(`case strings.EqualFold(r.Method, "%s"):`, httpMethod)
			if method.Auth {
				if err := genAuth(p, method); err != nil {
					return err
				}
			}
			p.printf(`h.wrapper%s(w, r)`, method.name)
		}
		p.printf(`default:`)
		if method, ok := byMethod[anyHTTPMethod]; ok {
			if method.Auth {
				if err := genAuth(p, method); err != nil {
					return err
				}
			}
			p.printf(`h.wrapper%s(w, r)`, method.name)
		} else {
			p.printf(`writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New("bad method")})`)
			p.printf(`return`)
		}
		p.printf(`}`)
	}

	p.printf(`default:`)
	p.printf(`	writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New("unknown method")})`)
	p.printf(`}`)
	p.printf(`}`)

	return p.err
}

// XXX now generates dummy code
func genAuth(p *printer, _ *serviceMethod) error {
	p.printf(`if key := r.Header.Get("X-Auth"); key != "%s" { // XXX`, authKey)
	p.printf(`	writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})`)
	p.printf(`	return`)
	p.printf(`}`)
	return nil
}

func genMethodWrapper(p *printer, m *serviceMethod) error {

	// func (h *SomeStructName ) wrapperDoSomeJob() {
	// 	// заполнение структуры params
	// 	// валидирование параметров
	// 	res, err := h.DoSomeJob(ctx, params)
	// 	// прочие обработки
	// }

	p.printf(``)
	if m.recv.isPointer {
		p.printf(`func (h *%s) wrapper%s(w http.ResponseWriter, r *http.Request) {`, m.recv.name, m.name)
	} else {
		p.printf(`func (h %s) wrapper%s(w http.ResponseWriter, r *http.Request) {`, m.recv.name, m.name)
	}
	p.printf(`const op = "%s.wrapper%s"`, m.recv.name, m.name)
	p.printf(`var params %s`, m.params.name)

	p.printf(`if err := params.getFromRequest(r); err != nil {`)
	p.printf(`	writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})`)
	p.printf(`	return`)
	p.printf(`}`)

	p.printf(`if err := params.validate(); err != nil {`)
	p.printf(`	writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})`)
	p.printf(`	return`)
	p.printf(`}`)

	p.printf(`ctx := r.Context()`)
	if m.params.isPointer {
		p.printf(`res, err := h.%s(ctx, &params)`, m.name)
	} else {
		p.printf(`res, err := h.%s(ctx, params)`, m.name)
	}
	p.printf(`if err != nil {`)
	p.printf(`	switch err := err.(type) {`)
	p.printf(`	case *ApiError:`)
	p.printf(`		writeApiError(w, *err)`)
	p.printf(`	case ApiError:`)
	p.printf(`		writeApiError(w, err)`)
	p.printf(`	default:`)
	p.printf(`		writeApiError(w, ApiError{HTTPStatus: http.StatusInternalServerError, Err: err})`)
	p.printf(`	}`)
	p.printf(`	return`)
	p.printf(`}`)

	p.printf(`resp := struct {`)
	if m.result.isPointer {
		p.printf(`	Response *%s `+q+`json:"response"`+q, m.result.name)
	} else {
		p.printf(`	Response  %s `+q+`json:"response"`+q, m.result.name)
	}
	p.printf(`	Error string ` + q + `json:"error"` + q)
	p.printf(`}{`)
	p.printf(`	Response: res,`)
	p.printf(`}`)

	p.printf(`w.Header().Add("content-type", "application/json")`)
	p.printf(`w.WriteHeader(http.StatusOK)`)

	p.printf(`if err := json.NewEncoder(w).Encode(&resp); err != nil {`)
	p.printf(`	log.Printf("%%s: can't write response body: %%v", op, err)`)
	p.printf(`}`)

	p.printf(`}`)

	return p.err
}

func genGetFromRequest(p *printer, structName string, fields []*paramStructField) error {
	p.printf(``)
	p.printf(`func (p *%s) getFromRequest(r *http.Request) error {`, structName)
	p.printf(`if r.Header.Get("content-type") == "application/json" {`)
	genGetFromJsonBody(p, structName, fields)
	p.printf(`} else {`)
	genGetFromFormOrQuery(p, structName, fields)
	p.printf(`}`)
	p.printf(`return nil`)
	p.printf(`}`)

	return p.err
}

func genGetFromJsonBody(p *printer, structName string, fields []*paramStructField) error {
	const op = "genGetFromJsonBody"

	p.printf(`// get from json body`)
	p.printf(`defer io.Copy(io.Discard, r.Body)`)
	p.printf(`req := struct{`)
	for _, field := range fields {
		if field.rules&requiredRule != 0 && field.rules&defaultRule == 0 {
			p.printf(`%s *%s `+q+`json:"%s"`+q, field.name, field.kind, field.apiParamName())
		} else {
			p.printf(`%s %s `+q+`json:"%s"`+q, field.name, field.kind, field.apiParamName())
		}
	}
	p.printf(`}{`)
	for _, field := range fields {
		if field.rules&defaultRule != 0 {
			if field.kind == String {
				p.printf(`%s: %q,`, field.name, field.defaultVal)
			} else {
				p.printf(`%s: %s,`, field.name, field.defaultVal)
			}
		}
	}
	p.printf(`}`)

	p.printf(`if err := json.NewDecoder(r.Body).Decode(&req); err != nil { return /*bad json*/ err }`)

	for _, field := range fields {
		if field.rules&requiredRule != 0 && field.rules&defaultRule == 0 {
			p.printf(`if req.%s == nil { return errors.New("%s must be not empty") }`, field.name, field.apiParamName())
		}
	}

	for _, field := range fields {
		if field.rules&requiredRule != 0 && field.rules&defaultRule == 0 {
			p.printf(`p.%s = *req.%s`, field.name, field.name)
		} else {
			p.printf(`p.%s = req.%s`, field.name, field.name)
		}
	}

	return nil
}

func genGetFromFormOrQuery(p *printer, structName string, fields []*paramStructField) error {
	const op = "genGetFromFormOrQuery"

	p.printf(`// get from form or query`)
	for _, field := range fields {
		p.printf(`{`)
		p.printf(`s := r.FormValue(%q)`, field.apiParamName())

		if field.rules&requiredRule != 0 && field.rules&defaultRule == 0 {
			p.printf(`if s == "" { return errors.New("%s must be not empty") }`, field.apiParamName())
		}

		if field.rules&defaultRule != 0 {
			p.printf(`if s == "" {`)
			if field.kind == String {
				p.printf(`p.%s = %q`, field.name, field.defaultVal)
			} else {
				p.printf(`p.%s = %s`, field.name, field.defaultVal)
			}
			p.printf(`} else {`)
		}

		switch field.kind {
		case String:
			p.printf(`p.%s = s`, field.name)
		case Int:
			p.printf(`v, err := strconv.Atoi(s)`)
			p.printf(`if err != nil { return errors.New("%s must be int") }`, field.apiParamName())
			p.printf(`p.%s = v`, field.name)
		case Float32:
			p.printf(`v, err := strconv.ParseFloat(s, 32)`)
			p.printf(`if err != nil { return errors.New("%s must be float32") }`, field.apiParamName())
			p.printf(`p.%s = v`, field.name)
		case Float64:
			p.printf(`v, err := strconv.ParseFloat(s, 64)`)
			p.printf(`if err != nil { return errors.New("%s must be float64") }`, field.apiParamName())
			p.printf(`p.%s = v`, field.name)
		default:
			return &ParseError{
				Err: fmt.Errorf(`%s: %s.%s: invalid param type: %v`, op, structName, field.name, field.kind),
				Pos: field.pos,
			}
		}

		if field.rules&defaultRule != 0 {
			p.printf(`}`)
		}

		p.printf(`}`)
	}
	return nil
}

func genValidate(p *printer, structName string, fields []*paramStructField) error {
	const op = `genValidate`

	p.printf(``)
	p.printf(`func (p *%s) validate() error {`, structName)

	for _, field := range fields {
		if field.rules == 0 {
			continue
		}

		// moved to GetFrom*
		// if field.rules&requiredRule != 0 {...}

		// moved to GetFrom*
		// if field.rules&defaultRule != 0 {...}

		if field.rules&enumRule != 0 {
			p.printf(`valid := false`)
			switch field.kind {
			case String:
				for _, s := range field.enum {
					p.printf(`valid = valid || p.%s == %q`, field.name, s)
				}
				p.printf(`if !valid { return errors.New("%s must be one of [%s]") }`, field.apiParamName(), strings.Join(field.enum, `, `))
			case Int, Float32, Float64:
				for _, s := range field.enum {
					p.printf(`valid |= p.%s == %s`, field.name, s)
				}
				p.printf(`if !valid { return errors.New("%s must be one of [%s]") }`, field.apiParamName(), strings.Join(field.enum, `, `))
			default:
				return &ParseError{
					Err: fmt.Errorf(`%s: %s.%s: enum rule not applicable for %v type`, op, structName, field.name, field.kind),
					Pos: field.pos,
				}
			}
		}

		if field.rules&minRule != 0 {
			switch field.kind {
			case String:
				p.printf(`if !(len(p.%s) >= %s) { return errors.New("%s len must be >= %s") }`, field.name, field.min, field.apiParamName(), field.min)
			case Int, Float32, Float64:
				p.printf(`if !(p.%s >= %s) { return errors.New("%s must be >= %s") }`, field.name, field.min, field.apiParamName(), field.min)
			default:
				return &ParseError{
					Err: fmt.Errorf(`%s: %s.%s: min rule not applicable for %v type`, op, structName, field.name, field.kind),
					Pos: field.pos,
				}
			}
		}

		if field.rules&maxRule != 0 {
			switch field.kind {
			case String:
				p.printf(`if !(len(p.%s) <= %s) { return errors.New("%s len must be <= %s") }`, field.name, field.min, field.apiParamName(), field.min)
			case Int, Float32, Float64:
				p.printf(`if !(p.%s <= %s) { return errors.New("%s must be <= %s") }`, field.name, field.max, field.apiParamName(), field.max)
			default:
				return &ParseError{
					Err: fmt.Errorf(`%s: %s.%s: max rule not applicable for %v type`, op, structName, field.name, field.kind),
					Pos: field.pos,
				}
			}
		}

		if field.rules&greaterRule != 0 {
			switch field.kind {
			case String:
				p.printf(`if !(len(p.%s) > %s) { return errors.New("%s len must be > %s") }`, field.name, field.greater, field.apiParamName(), field.greater)
			case Int, Float32, Float64:
				p.printf(`if !(p.%s > %s) { return errors.New("%s must be > %s") }`, field.name, field.greater, field.apiParamName(), field.greater)
			default:
				return &ParseError{
					Err: fmt.Errorf(`%s: %s.%s: greate rule not applicable for %v type`, op, structName, field.name, field.kind),
					Pos: field.pos,
				}
			}
		}

		if field.rules&lessRule != 0 {
			switch field.kind {
			case String:
				p.printf(`if !(len(p.%s) < %s) { return errors.New("%s len must be < %s") }`, field.name, field.less, field.apiParamName(), field.less)
			case Int, Float32, Float64:
				p.printf(`if !(p.%s < %s) { return errors.New("%s must be < %s") }`, field.name, field.less, field.apiParamName(), field.less)
			default:
				return &ParseError{
					Err: fmt.Errorf(`%s: %s.%s: greate rule not applicable for %v type`, op, structName, field.name, field.kind),
					Pos: field.pos,
				}
			}
		}
	}

	p.printf(`return nil`)
	p.printf(`}`)

	return p.err
}
