// !!! Do not change this code !!!
// The code is generated automatically by apigen tool
package main

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

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		switch /*r.Method*/ {
		case strings.EqualFold(r.Method, "POST"):
			if key := r.Header.Get("X-Auth"); key != "100500" { // XXX
				writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})
				return
			}
			h.wrapperCreate(w, r)
		default:
			writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New("bad method")})
			return
		}
	case "/user/profile":
		switch /*r.Method*/ {
		default:
			h.wrapperProfile(w, r)
		}
	default:
		writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New("unknown method")})
	}
}

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	const op = "MyApi.wrapperProfile"
	var params ProfileParams
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.Profile(ctx, params)
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
		Response *User  `json:"response"`
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

func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	const op = "MyApi.wrapperCreate"
	var params CreateParams
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.Create(ctx, params)
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
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		switch /*r.Method*/ {
		case strings.EqualFold(r.Method, "POST"):
			if key := r.Header.Get("X-Auth"); key != "100500" { // XXX
				writeApiError(w, ApiError{HTTPStatus: http.StatusForbidden, Err: errors.New("unauthorized")})
				return
			}
			h.wrapperCreate(w, r)
		default:
			writeApiError(w, ApiError{HTTPStatus: http.StatusNotAcceptable, Err: errors.New("bad method")})
			return
		}
	default:
		writeApiError(w, ApiError{HTTPStatus: http.StatusNotFound, Err: errors.New("unknown method")})
	}
}

func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	const op = "OtherApi.wrapperCreate"
	var params OtherCreateParams
	if err := params.getFromRequest(r); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	if err := params.validate(); err != nil {
		writeApiError(w, ApiError{HTTPStatus: http.StatusBadRequest, Err: err})
		return
	}
	ctx := r.Context()
	res, err := h.Create(ctx, params)
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
		Response *OtherUser `json:"response"`
		Error    string     `json:"error"`
	}{
		Response: res,
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Printf("%s: can't write response body: %v", op, err)
	}
}

func (p *CreateParams) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			Login  *string `json:"login"`
			Name   string  `json:"full_name"`
			Status string  `json:"status"`
			Age    int     `json:"age"`
		}{
			Status: "user",
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.Login == nil {
			return errors.New("login must be not empty")
		}
		p.Login = *req.Login
		p.Name = req.Name
		p.Status = req.Status
		p.Age = req.Age
	} else {
		// get from form or query
		{
			s := r.FormValue("login")
			if s == "" {
				return errors.New("login must be not empty")
			}
			p.Login = s
		}
		{
			s := r.FormValue("full_name")
			p.Name = s
		}
		{
			s := r.FormValue("status")
			if s == "" {
				p.Status = "user"
			} else {
				p.Status = s
			}
		}
		{
			s := r.FormValue("age")
			v, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("age must be int")
			}
			p.Age = v
		}
	}
	return nil
}

func (p *CreateParams) validate() error {
	if !(len(p.Login) >= 10) {
		return errors.New("login len must be >= 10")
	}
	valid := false
	valid = valid || p.Status == "user"
	valid = valid || p.Status == "moderator"
	valid = valid || p.Status == "admin"
	if !valid {
		return errors.New("status must be one of [user, moderator, admin]")
	}
	if !(p.Age >= 0) {
		return errors.New("age must be >= 0")
	}
	if !(p.Age <= 128) {
		return errors.New("age must be <= 128")
	}
	return nil
}

func (p *OtherCreateParams) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			Username *string `json:"username"`
			Name     string  `json:"account_name"`
			Class    string  `json:"class"`
			Level    int     `json:"level"`
		}{
			Class: "warrior",
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.Username == nil {
			return errors.New("username must be not empty")
		}
		p.Username = *req.Username
		p.Name = req.Name
		p.Class = req.Class
		p.Level = req.Level
	} else {
		// get from form or query
		{
			s := r.FormValue("username")
			if s == "" {
				return errors.New("username must be not empty")
			}
			p.Username = s
		}
		{
			s := r.FormValue("account_name")
			p.Name = s
		}
		{
			s := r.FormValue("class")
			if s == "" {
				p.Class = "warrior"
			} else {
				p.Class = s
			}
		}
		{
			s := r.FormValue("level")
			v, err := strconv.Atoi(s)
			if err != nil {
				return errors.New("level must be int")
			}
			p.Level = v
		}
	}
	return nil
}

func (p *OtherCreateParams) validate() error {
	if !(len(p.Username) >= 3) {
		return errors.New("username len must be >= 3")
	}
	valid := false
	valid = valid || p.Class == "warrior"
	valid = valid || p.Class == "sorcerer"
	valid = valid || p.Class == "rouge"
	if !valid {
		return errors.New("class must be one of [warrior, sorcerer, rouge]")
	}
	if !(p.Level >= 1) {
		return errors.New("level must be >= 1")
	}
	if !(p.Level <= 50) {
		return errors.New("level must be <= 50")
	}
	return nil
}

func (p *ProfileParams) getFromRequest(r *http.Request) error {
	if r.Header.Get("content-type") == "application/json" {
		// get from json body
		defer io.Copy(io.Discard, r.Body)
		req := struct {
			Login *string `json:"login"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return /*bad json*/ err
		}
		if req.Login == nil {
			return errors.New("login must be not empty")
		}
		p.Login = *req.Login
	} else {
		// get from form or query
		{
			s := r.FormValue("login")
			if s == "" {
				return errors.New("login must be not empty")
			}
			p.Login = s
		}
	}
	return nil
}

func (p *ProfileParams) validate() error {
	return nil
}
