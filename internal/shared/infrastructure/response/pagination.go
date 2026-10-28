package response

import (
	"github.com/gin-gonic/gin"
)

type PaginationMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Limit    int `json:"limit,omitempty"`
	Offset   int `json:"offset,omitempty"`
}

func NewPaginationMeta(limit, offset int) PaginationMeta {
	page := 1
	if limit > 0 {
		page = offset/limit + 1
	}
	if page < 1 {
		page = 1
	}
	return PaginationMeta{
		Page:     page,
		PageSize: limit,
		Limit:    limit,
		Offset:   offset,
	}
}

func SuccessWithPagination(c *gin.Context, message string, status int, data any, meta PaginationMeta) {
	if data == nil {
		data = []any{}
	}
	c.JSON(status, gin.H{
		"success":    true,
		"status":     "success",
		"message":    message,
		"data":       data,
		"pagination": meta,
	})
}
