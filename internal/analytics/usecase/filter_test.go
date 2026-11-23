package usecase

import (
	"testing"
	"time"

	analyticdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

func TestNormalizeFilter(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		filter, err := NormalizeFilter(analyticdomain.Filter{})
		if err != nil {
			t.Fatal(err)
		}
		if filter.Interval != "day" {
			t.Fatalf("interval = %q", filter.Interval)
		}
		if filter.Limit != 10 {
			t.Fatalf("limit = %d", filter.Limit)
		}
		if filter.To.Sub(filter.From) < 29*24*time.Hour || filter.To.Sub(filter.From) > 31*24*time.Hour {
			t.Fatalf("unexpected default range: %v", filter.To.Sub(filter.From))
		}
	})

	t.Run("rejects long range", func(t *testing.T) {
		to := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
		_, err := NormalizeFilter(analyticdomain.Filter{
			From: to.AddDate(-2, 0, 0), To: to,
		})
		if err != ErrInvalidRange {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects invalid interval", func(t *testing.T) {
		to := time.Now().UTC()
		_, err := NormalizeFilter(analyticdomain.Filter{
			From: to.Add(-24 * time.Hour), To: to, Interval: "hour",
		})
		if err != ErrInvalidInterval {
			t.Fatalf("err = %v", err)
		}
	})
}
