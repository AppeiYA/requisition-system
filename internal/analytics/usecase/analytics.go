package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	analyticdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
	analyticsout "github.com/AppeiYA/requisition-system/internal/analytics/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
)

const cacheTTL = 60 * time.Second

type AnalyticsUsecase struct {
	repo   analyticsout.Repository
	cache  cache.Cache
	source string
}

func NewAnalyticsUsecase(repo analyticsout.Repository, cache cache.Cache, source string) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo, cache: cache, source: source}
}

func (u *AnalyticsUsecase) ExecuteOverview(ctx context.Context, filter analyticdomain.Filter) (*analyticdomain.Overview, error) {
	filter, err := NormalizeFilter(filter)
	if err != nil {
		return nil, err
	}
	key := u.analyticsCacheKey("overview", filter)
	if u.cache != nil {
		var result analyticdomain.Overview
		if found, err := u.cache.Get(ctx, key, &result); err == nil && found {
			setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
			return &result, nil
		}
	}
	result, err := u.repo.GetOverview(ctx, filter)
	if err == nil {
		setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
	}
	if err == nil && u.cache != nil {
		_ = u.cache.Set(ctx, key, result, cacheTTL)
	}
	return result, err
}

func (u *AnalyticsUsecase) ExecuteRequisitions(ctx context.Context, filter analyticdomain.Filter) (*analyticdomain.RequisitionAnalytics, error) {
	filter, err := NormalizeFilter(filter)
	if err != nil {
		return nil, err
	}
	key := u.analyticsCacheKey("requisitions", filter)
	if u.cache != nil {
		var result analyticdomain.RequisitionAnalytics
		if found, err := u.cache.Get(ctx, key, &result); err == nil && found {
			setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
			return &result, nil
		}
	}
	result, err := u.repo.GetRequisitionAnalytics(ctx, filter)
	if err == nil {
		setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
	}
	if err == nil && u.cache != nil {
		_ = u.cache.Set(ctx, key, result, cacheTTL)
	}
	return result, err
}

func (u *AnalyticsUsecase) ExecuteUsers(ctx context.Context, filter analyticdomain.Filter) (*analyticdomain.UserAnalytics, error) {
	filter, err := NormalizeFilter(filter)
	if err != nil {
		return nil, err
	}
	key := u.analyticsCacheKey("users", filter)
	if u.cache != nil {
		var result analyticdomain.UserAnalytics
		if found, err := u.cache.Get(ctx, key, &result); err == nil && found {
			setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
			return &result, nil
		}
	}
	result, err := u.repo.GetUserAnalytics(ctx, filter)
	if err == nil {
		setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
	}
	if err == nil && u.cache != nil {
		_ = u.cache.Set(ctx, key, result, cacheTTL)
	}
	return result, err
}

func (u *AnalyticsUsecase) ExecuteCatalogue(ctx context.Context, filter analyticdomain.Filter) (*analyticdomain.CatalogueAnalytics, error) {
	filter, err := NormalizeFilter(filter)
	if err != nil {
		return nil, err
	}
	key := u.analyticsCacheKey("catalogue", filter)
	if u.cache != nil {
		var result analyticdomain.CatalogueAnalytics
		if found, err := u.cache.Get(ctx, key, &result); err == nil && found {
			setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
			return &result, nil
		}
	}
	result, err := u.repo.GetCatalogueAnalytics(ctx, filter)
	if err == nil {
		setAnalyticsMetadata(result.GeneratedAt, u.source, &result.DataAsOf, &result.Source)
	}
	if err == nil && u.cache != nil {
		_ = u.cache.Set(ctx, key, result, cacheTTL)
	}
	return result, err
}

func (u *AnalyticsUsecase) analyticsCacheKey(endpoint string, filter analyticdomain.Filter) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%d",
		filter.From.UTC().Format(time.RFC3339),
		filter.To.UTC().Format(time.RFC3339),
		filter.DepartmentID,
		filter.Interval,
		endpoint,
		filter.Limit,
	)
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("analytics:v1:%s:%s:%x", u.source, endpoint, sum)
}

func setAnalyticsMetadata(generatedAt time.Time, source string, dataAsOf *time.Time, resultSource *string) {
	*resultSource = source
	if source == "direct" {
		*dataAsOf = generatedAt
		return
	}
	if (*dataAsOf).IsZero() {
		*dataAsOf = generatedAt
	}
}
