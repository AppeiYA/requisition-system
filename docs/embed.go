package docs

import _ "embed"

// OpenAPISpecYAML is the OpenAPI 3 specification of the API, served to the
// Scalar API reference. Regenerate it after changing any swag annotation:
//
//	swag init -g main.go -d ./cmd/api,./internal -o docs
//	go run ./cmd/specgen
//
//go:embed openapi.yaml
var OpenAPISpecYAML []byte
