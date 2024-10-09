package main

import "loshon-api/internals/app"

func main() {
	app := app.NewApp()
	app.RunMigrate()
}
