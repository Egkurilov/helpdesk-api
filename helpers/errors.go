package helpers

import "github.com/gin-gonic/gin"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAPIError(code, message string) APIError {
	return APIError{Code: code, Message: message}
}

func SendError(c *gin.Context, status int, err APIError) {
	c.JSON(status, gin.H{"error": err})
}
