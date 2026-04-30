package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for JSON API responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Error   any    `json:"error"`
}

// RespondSuccess writes a success envelope with the given HTTP status and data payload.
func RespondSuccess(c *gin.Context, status int, message string, data any) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

// RespondBadRequest writes a 400 error envelope.
func RespondBadRequest(c *gin.Context, message string, err string) {
	RespondError(c, http.StatusBadRequest, message, err)
}

// RespondError writes an error envelope with the given HTTP status.
func RespondError(c *gin.Context, status int, message string, err any) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   err,
	})
}

