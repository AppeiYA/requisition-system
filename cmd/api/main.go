package main

import app "github.com/AppeiYA/requisition-system/internal"

func main() {
	err := app.NewApp().Run()
	if err != nil {
		panic(err)
	}
}
