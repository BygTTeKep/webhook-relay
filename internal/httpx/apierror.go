package httpx

import "net/http"

type APIError struct {
	Status  int
	Message string
	Err     error
}

func (e APIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func BadRequest(msg string, err error) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Message: msg,
		Err:     err,
	}
}

func NotFound(msg string, err error) *APIError {
	return &APIError{
		Status:  http.StatusNotFound,
		Message: msg,
		Err:     err,
	}
}
