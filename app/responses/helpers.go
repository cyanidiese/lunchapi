package responses

import (
	"net/http"
)

func SuccessfulResponse(message string) GeneralResponse{
	if len(message) == 0 {
		message = "Your request was completed successfully"
	}
	return GeneralResponse{
		Name:    "Success",
		Message: message,
		Status:  http.StatusOK,
	}
}
