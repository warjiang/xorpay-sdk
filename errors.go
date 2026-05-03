package xorpay

import (
	"errors"
	"fmt"
)

type APIError struct {
	Endpoint   string
	Status     string
	Message    string
	HTTPStatus int
	RawBody    string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return fmt.Sprintf("xorpay api error endpoint=%s status=%s message=%s", e.Endpoint, e.Status, e.Message)
	}
	return fmt.Sprintf("xorpay api error endpoint=%s status=%s", e.Endpoint, e.Status)
}

func IsStatus(err error, status string) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Status == status
}
