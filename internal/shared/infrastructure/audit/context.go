package audit

import "context"

// RequestMeta is the transport-level metadata attached to a request by the
// admin HTTP middleware. The use cases never see HTTP, so the recorder reads
// this from the context instead.
type RequestMeta struct {
	RequestID string
	IPAddress string
	UserAgent string
}

type metaKey struct{}

// WithRequestMeta attaches request metadata to a context.
func WithRequestMeta(ctx context.Context, meta RequestMeta) context.Context {
	return context.WithValue(ctx, metaKey{}, meta)
}

// RequestMetaFromContext returns the request metadata, if the request went
// through the admin middleware.
func RequestMetaFromContext(ctx context.Context) (RequestMeta, bool) {
	meta, ok := ctx.Value(metaKey{}).(RequestMeta)
	return meta, ok
}
