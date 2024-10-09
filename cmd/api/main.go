package main

import (
	"log"
	"loshon-api/internals/app"
)

func main() {
	app := app.NewApp()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
