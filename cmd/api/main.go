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
	launchApp := app.NewApp()
	cfg, err := launchApp.Config()
	if err != nil {
		panic(err)
	}

	if cfg.App.Binaries == 1 {
		errCh := make(chan error, 2)

		go func() {
			errCh <- launchApp.Run()
		}()
		go func() {
			errCh <- app.NewWorker().Run()
		}()

		for range 2 {
			if err := <-errCh; err != nil {
				panic(err)
			}
		}
		return
	}

	if err := launchApp.Run(); err != nil {
		panic(err)
	}
}
