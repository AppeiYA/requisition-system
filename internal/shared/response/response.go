package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func isSuccess(code int) bool {
	return code >= http.StatusOK && code < http.StatusMultipleChoices
}

func Success(c *gin.Context, message string, code int, data any) {
	c.JSON(code, Response{
		Success: isSuccess(code),
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, message string, code int, data any) {
	c.AbortWithStatusJSON(code, Response{
		Success: isSuccess(code),
		Message: message,
		Data:    data,
	})
}
