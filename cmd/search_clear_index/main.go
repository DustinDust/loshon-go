package main

import (
	"log"
	"loshon-api/internals/config"
	"loshon-api/internals/search"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config %v", err)
	}
	searchClient, err := search.NewSearchClient(config.AlgoliaAppID, config.AlgoliaAPIKey)
	if err != nil {
		log.Fatalf("failed to create search client %v", err)
	}
	searchClient.Clear(config.SearchIndex)
}
