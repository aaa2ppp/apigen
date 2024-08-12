package model

import "net/http"

type ApiError struct {
	HTTPStatus int
	Err        error
}

func (e ApiError) Error() string {
	if e.Err == nil {
		return http.StatusText(e.HTTPStatus)
	} else {
		return e.Err.Error()
	}
}
