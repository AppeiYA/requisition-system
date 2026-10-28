package pagination

import (
	"net/http"
	"strconv"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 100
	MaxLimit     = 100
)

type Query struct {
	Limit  int
	Offset int
}

func NewQuery(limit, offset int) Query {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return Query{Limit: limit, Offset: offset}
}

func Parse(c *gin.Context) (Query, bool) {
	q := Query{}
	limit := DefaultLimit
	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > MaxLimit {
			response.Error(c, "invalid limit: must be an integer between 1 and "+strconv.Itoa(MaxLimit), http.StatusBadRequest, nil)
			return Query{}, false
		}
		limit = value
	}
	q.Limit = limit

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			response.Error(c, "invalid offset: must be a non-negative integer", http.StatusBadRequest, nil)
			return Query{}, false
		}
		offset = value
	}
	q.Offset = offset
	return q, true
}

func Clamp(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
