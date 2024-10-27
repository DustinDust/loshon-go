package main

import (
	"log"
	"loshon-api/internals/config"
	"loshon-api/internals/data"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load app config %v", err)
	}
	db, err := data.OpenDB(config.PostgresUrl)
	if err != nil {
		log.Fatalf("Failed to open connection to DB %v", err)
	}
	db.AutoMigrate(data.Document{})
}
