package response

import (
	"github.com/gin-gonic/gin"
)

type PaginationMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Limit    int `json:"limit,omitempty"`
	Offset   int `json:"offset,omitempty"`
	// Total is a pointer so that list endpoints which do not count rows keep
	// omitting the field, while admin lists can report an empty result as
	// total: 0 instead of hiding it.
	Total *int `json:"total,omitempty"`
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

// NewPaginationMetaWithTotal builds pagination metadata that includes the
// total number of matching rows.
func NewPaginationMetaWithTotal(limit, offset, total int) PaginationMeta {
	return NewPaginationMeta(limit, offset).WithTotal(total)
}

// WithTotal returns a copy of the metadata carrying the total row count.
func (m PaginationMeta) WithTotal(total int) PaginationMeta {
	m.Total = &total
	return m
}

func SuccessWithPagination(c *gin.Context, message string, status int, data any, meta PaginationMeta) {
	writePagination(c, message, status, data, meta)
}

// SuccessWithPaginationTotal is the explicit form for endpoints that count the
// full result set. It exists so a caller cannot forget to set Total.
func SuccessWithPaginationTotal(c *gin.Context, message string, status int, data any, meta PaginationMeta) {
	writePagination(c, message, status, data, meta)
}

func writePagination(c *gin.Context, message string, status int, data any, meta PaginationMeta) {
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
