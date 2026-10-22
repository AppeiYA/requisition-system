// Package version exposes the identity of the build.
//
// Version is overridden at build time with:
//
//	go build -ldflags "-X github.com/AppeiYA/requisition-system/internal/version.Version=v1.2.3"
//
// It is reported as the service.version resource attribute, so a backend can
// tell which build produced a trace or a metric.
package version

// Version is the version of the build. It falls back to "dev" for a build
// that was not stamped.
var Version = "dev"
