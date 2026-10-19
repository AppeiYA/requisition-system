// Requisition System API.
//
// The API reference is rendered by Scalar at /api/v1/docs from the OpenAPI 3
// specification embedded in the docs package. The specification is generated
// from the swag annotations; after changing any annotation, regenerate it
// with:
//
//	swag init -g main.go -d ./cmd/api,./internal -o docs
//	go run ./cmd/specgen
//
// @title Requisition System API
// @version 1.0
// @description Requisition System REST API.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the access token.
package main

import app "github.com/AppeiYA/requisition-system/internal"

func main() {
	err := app.NewApp().Run()
	if err != nil {
		panic(err)
	}
}
