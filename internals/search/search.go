package search

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
	"github.com/labstack/gommon/log"
)

type SearchClient struct {
	client *search.APIClient
}

func NewSearchClient(appID, apiKey string) (*SearchClient, error) {
	client, err := search.NewClient(appID, apiKey)
	if err != nil {
		return nil, err
	}
	return &SearchClient{
		client: client,
	}, nil
}

func (search SearchClient) Reindex(indexName string, data []map[string]any) {
	resps, err := search.client.SaveObjects(indexName, data)
	if err != nil {
		slog.Warn("failed to save objects", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return
	}
	for _, resp := range resps {
		slog.Info("object indexed", slog.Attr{Key: "objectID", Value: slog.StringValue(strings.Join(resp.ObjectIDs, ", "))})
	}
}

func StructsToMaps[T any](obj []T) ([]map[string]any, error) {
	result := make([]map[string]any, 0)

	jbyte, err := json.Marshal(obj)
	if err != nil {
		log.Warnf("failed to convert struct to map %v", err)
		return result, err
	}
	if err := json.Unmarshal(jbyte, &result); err != nil {
		log.Warnf("failed to convert struct to map %v", err)
		return result, err
	}
	return result, nil
}
