package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
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

// ErrorFrom maps an application error onto the shared response format.
// *customerrors.CustomError carries the status code; anything else is reported
// as an internal server error without leaking infrastructure detail.
func ErrorFrom(c *gin.Context, err error) {
	var customErr *customerrors.CustomError
	if errors.As(err, &customErr) {
		Error(c, customErr.Message, customErr.Code, nil)
		return
	}

	Error(c, "internal server error", http.StatusInternalServerError, nil)
}
