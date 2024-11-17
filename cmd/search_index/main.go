package main

import (
	"log"
	"loshon-api/internals/config"
	"loshon-api/internals/data"
	"loshon-api/internals/search"
)

/*
	USAGE: INITIALIZE SEARCH INDEX WITH ALL CURRENT DOCUMENTS IN DB
*/

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config %v", err)
	}

	searchClient, err := search.NewSearchClient(config.AngoliaAppID, config.AngoliaAPIKey)
	if err != nil {
		log.Fatalf("failed to create search client %v", err)
	}

	gormdb, err := data.OpenDB(config.PostgresUrl)
	if err != nil {
		log.Fatalf("failed to open db %v", err)
	}

	documents := []data.Document{}
	if err := gormdb.Find(&documents).Error; err != nil {
		log.Fatalf("failed to fetch documents %v", err)
	}

	documentSearchObjects := []map[string]any{}
	for _, doc := range documents {
		documentSearchObjects = append(documentSearchObjects, doc.ToSearchObject())
	}

	searchClient.Reindex("documents", documentSearchObjects)
}
