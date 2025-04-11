package search

import (
	"log/slog"
	"strings"

	"github.com/algolia/algoliasearch-client-go/v4/algolia/search"
)

type SearchClientInterface interface {
	Reindex(indexName string, data []map[string]any) error
	SaveObject(indexName string, data map[string]any) error
	Clear(indexName string) error
}

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

func (search SearchClient) Reindex(indexName string, data []map[string]any) error {
	resps, err := search.client.SaveObjects(indexName, data)
	if err != nil {
		return err
	}
	for _, resp := range resps {
		slog.Info("object indexed", slog.Attr{Key: "objectID", Value: slog.StringValue(strings.Join(resp.ObjectIDs, ", "))})
	}
	return nil
}

func (sclient SearchClient) SaveObject(indexName string, data map[string]any) error {
	resp, err := sclient.client.SaveObject(
		sclient.client.NewApiSaveObjectRequest(indexName, data),
	)
	if err != nil {
		return err
	}
	slog.Info("search object saved", slog.Attr{
		Key:   "resp",
		Value: slog.StringValue(resp.String()),
	})
	return nil
}

func (sclient SearchClient) Clear(indexName string) error {
	indexExist, err := sclient.client.IndexExists(indexName)
	if err != nil {
		return err
	}
	if !indexExist {
		slog.Info("index does not exist", slog.Attr{Key: "indexName", Value: slog.StringValue(indexName)})
		return nil
	}
	_, err = sclient.client.ClearObjects(sclient.client.NewApiClearObjectsRequest(indexName))

	return err
}
