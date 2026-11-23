package usecase

import (
	"time"

	analyticdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidInterval = customerrors.NewBadRequestError("interval must be day, week or month")
	ErrInvalidRange    = customerrors.NewBadRequestError("from must be before to and the range must not exceed 366 days")
	ErrInvalidLimit    = customerrors.NewBadRequestError("limit must be between 1 and 50")
)

func NormalizeFilter(filter analyticdomain.Filter) (analyticdomain.Filter, error) {
	now := time.Now().UTC()
	if filter.To.IsZero() {
		filter.To = now
	}
	if filter.From.IsZero() {
		filter.From = filter.To.AddDate(0, 0, -30)
	}
	filter.From = filter.From.UTC()
	filter.To = filter.To.UTC()
	if filter.From.After(filter.To) || filter.From.Equal(filter.To) || filter.To.Sub(filter.From) > 366*24*time.Hour {
		return filter, ErrInvalidRange
	}
	if filter.Interval == "" {
		filter.Interval = "day"
	}
	switch filter.Interval {
	case "day", "week", "month":
	default:
		return filter, ErrInvalidInterval
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		return filter, ErrInvalidLimit
	}
	return filter, nil
}
