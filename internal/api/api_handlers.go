// !!! Do not change this code !!!
// The code is generated automatically by apigen from ./internal/api/api.go
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

func writeApiError(w http.ResponseWriter, ae ApiError) {
	const op = "writeApiError"
	var resp bytes.Buffer
	resp.WriteByte('{')
	resp.WriteString(`"error":`)
	resp.WriteString(strconv.Quote(ae.Error()))
	resp.WriteByte('}')
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(ae.HTTPStatus)
	if _, err := resp.WriteTo(w); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/users":
		if r.Method != "POST" {
			writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New("bad method")})
			return
		}
		h.wrapperCreateUser(w, r)
	default:
		writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New("unknown method")})
	}
}

func (h *Api) wrapperCreateUser(w http.ResponseWriter, r *http.Request) {
	const op = "Api.wrapperCreateUser"
	var params CreateUserParams
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.CreateUser(ctx, params)
	if err != nil {
		switch err := err.(type) {
		case *ApiError:
			writeApiError(w, *err)
		case ApiError:
			writeApiError(w, err)
		default:
			writeApiError(w, ApiError{HTTPStatus: http.StatusInternalServerError, Err: err})
		}
		return
	}
	resp := struct {
		Response *NewUser `json:"response"`
		Error    string   `json:"error"`
	}{
		Response: &res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (p *CreateUserParams) getFromRequest(r *http.Request) error {
	{
		s := r.FormValue("name")
		p.Name = s
	}
	{
		s := r.FormValue("skill")
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("skill must be float64")
		}
		p.Skill = v
	}
	{
		s := r.FormValue("latency")
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("latency must be float64")
		}
		p.Latency = v
	}
	return nil
}

func (p *CreateUserParams) validate() error {
	{
		if p.Name == "" {
			return errors.New("name must be not empty")
		}
	}
	{
		if p.Skill == 0 {
			return errors.New("skill must be not empty")
		}
		if p.Skill < 0 {
			return errors.New("skill must be >= 0")
		}
	}
	{
		if p.Latency == 0 {
			return errors.New("latency must be not empty")
		}
		if p.Latency < 0 {
			return errors.New("latency must be >= 0")
		}
	}
	return nil
}
