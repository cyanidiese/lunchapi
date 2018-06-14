package errors

import (
	"net/http"
)

func ErrorNotFound(message string) RequestError{
	if len(message) == 0 {
		message = "Unable to find entity based on your request."
	}
	return RequestError{
		Name:    "Not Found",
		Message: message,
		Status:  http.StatusNotFound,
	}
}

func ErrorBadRequest(message string, additionalData map[int64]int64) RequestError{
	if len(message) == 0 {
		message = "Your request was made without required parameters."
	}
	return RequestError{
		Name:    "Bad Request",
		Message: message,
		Status:  http.StatusBadRequest,
		Data:  additionalData,
	}
}

func ErrorForbidden(message string) RequestError{
	if len(message) == 0 {
		message = "You have no permission to do this"
	}
	return RequestError{
		Name:    "Forbidden",
		Message: message,
		Status:  http.StatusForbidden,
	}
}

func ErrorUnauthorized(message string) RequestError{
	if len(message) == 0 {
		message = "Your request was made with invalid credentials."
	}
	return RequestError{
		Name:    "Unauthorized",
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}
