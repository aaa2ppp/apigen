// !!! Do not change this code !!!
// The code is generated automatically by apigen tool
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func writeApiError(w http.ResponseWriter, ae ApiError) {
	const op = "writeApiError"
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(ae.HTTPStatus)
	if _, err := fmt.Fprintf(w, "{\"error\":%q}", ae.Err.Error()); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/users":
		switch /*r.Method*/ {
		case strings.EqualFold(r.Method, "DELETE"):
			if key := r.Header.Get("X-Auth"); key != "100500" { // XXX
				writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})
				return
			}
			h.wrapperDeleteUser(w, r)
		case strings.EqualFold(r.Method, "GET"):
			if key := r.Header.Get("X-Auth"); key != "100500" { // XXX
				writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})
				return
			}
			h.wrapperGetUser(w, r)
		case strings.EqualFold(r.Method, "POST"):
			h.wrapperCreateUser(w, r)
		case strings.EqualFold(r.Method, "PUT"):
			if key := r.Header.Get("X-Auth"); key != "100500" { // XXX
				writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})
				return
			}
			h.wrapperUpdateUser(w, r)
		default:
			writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New("bad method")})
			return
		}
	default:
		writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New("unknown method")})
	}
}

func (h *Service) wrapperCreateUser(w http.ResponseWriter, r *http.Request) {
	const op = "Service.wrapperCreateUser"
	var params CreateUser
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
		Response NewUser `json:"response"`
		Error    string  `json:"error"`
	}{
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *Service) wrapperGetUser(w http.ResponseWriter, r *http.Request) {
	const op = "Service.wrapperGetUser"
	var params GetUser
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.GetUser(ctx, params)
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
		Response User   `json:"response"`
		Error    string `json:"error"`
	}{
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *Service) wrapperUpdateUser(w http.ResponseWriter, r *http.Request) {
	const op = "Service.wrapperUpdateUser"
	var params UpdateUser
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.UpdateUser(ctx, params)
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
		Response None   `json:"response"`
		Error    string `json:"error"`
	}{
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *Service) wrapperDeleteUser(w http.ResponseWriter, r *http.Request) {
	const op = "Service.wrapperDeleteUser"
	var params DeleteUser
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.DeleteUser(ctx, params)
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
		Response None   `json:"response"`
		Error    string `json:"error"`
	}{
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (p *CreateUser) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			Name    *string `json:"name"`
			Skill   float64 `json:"skill"`
			Latency float64 `json:"latency"`
		}{
			Skill:   0,
			Latency: 1,
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.Name == nil {
			return errors.New("name must be not empty")
		}
		p.Name = *req.Name
		p.Skill = req.Skill
		p.Latency = req.Latency
	} else {
		// get from form or query
		{
			s := r.FormValue("name")
			if s == "" {
				return errors.New("name must be not empty")
			}
			p.Name = s
		}
		{
			s := r.FormValue("skill")
			if s == "" {
				p.Skill = 0
			} else {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return errors.New("skill must be float64")
				}
				p.Skill = v
			}
		}
		{
			s := r.FormValue("latency")
			if s == "" {
				p.Latency = 1
			} else {
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return errors.New("latency must be float64")
				}
				p.Latency = v
			}
		}
	}
	return nil
}

func (p *CreateUser) validate() error {
	if !(p.Skill >= 0) {
		return errors.New("skill must be >= 0")
	}
	if !(p.Latency > 0) {
		return errors.New("latency must be > 0")
	}
	return nil
}

func (p *DeleteUser) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			ID *int `json:"id"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.ID == nil {
			return errors.New("id must be not empty")
		}
		p.ID = *req.ID
	} else {
		// get from form or query
		{
			s := r.FormValue("id")
			if s == "" {
				return errors.New("id must be not empty")
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("id must be int")
			}
			p.ID = v
		}
	}
	return nil
}

func (p *DeleteUser) validate() error {
	if !(p.ID > 0) {
		return errors.New("id must be > 0")
	}
	return nil
}

func (p *GetUser) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			ID *int `json:"id"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.ID == nil {
			return errors.New("id must be not empty")
		}
		p.ID = *req.ID
	} else {
		// get from form or query
		{
			s := r.FormValue("id")
			if s == "" {
				return errors.New("id must be not empty")
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("id must be int")
			}
			p.ID = v
		}
	}
	return nil
}

func (p *GetUser) validate() error {
	if !(p.ID > 0) {
		return errors.New("id must be > 0")
	}
	return nil
}

func (p *UpdateUser) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			ID      *int     `json:"id"`
			Name    *string  `json:"name"`
			Skill   *float64 `json:"skill"`
			Latency *float64 `json:"latency"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.ID == nil {
			return errors.New("id must be not empty")
		}
		if req.Name == nil {
			return errors.New("name must be not empty")
		}
		if req.Skill == nil {
			return errors.New("skill must be not empty")
		}
		if req.Latency == nil {
			return errors.New("latency must be not empty")
		}
		p.ID = *req.ID
		p.Name = *req.Name
		p.Skill = *req.Skill
		p.Latency = *req.Latency
	} else {
		// get from form or query
		{
			s := r.FormValue("id")
			if s == "" {
				return errors.New("id must be not empty")
			}
			v, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("id must be int")
			}
			p.ID = v
		}
		{
			s := r.FormValue("name")
			if s == "" {
				return errors.New("name must be not empty")
			}
			p.Name = s
		}
		{
			s := r.FormValue("skill")
			if s == "" {
				return errors.New("skill must be not empty")
			}
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return errors.New("skill must be float64")
			}
			p.Skill = v
		}
		{
			s := r.FormValue("latency")
			if s == "" {
				return errors.New("latency must be not empty")
			}
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return errors.New("latency must be float64")
			}
			p.Latency = v
		}
	}
	return nil
}

func (p *UpdateUser) validate() error {
	if !(p.ID > 0) {
		return errors.New("id must be > 0")
	}
	if !(p.Skill >= 0) {
		return errors.New("skill must be >= 0")
	}
	if !(p.Latency > 0) {
		return errors.New("latency must be > 0")
	}
	return nil
}
