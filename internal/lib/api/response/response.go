package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Respounse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func OK() Respounse {
	return Respounse{
		Status: StatusOK,
	}
}

func Error(msg string) Respounse {
	return Respounse{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Respounse {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid url", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}
	return Respounse{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
