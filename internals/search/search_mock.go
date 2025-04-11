package search

import "github.com/stretchr/testify/mock"

type MockSearchClient struct {
	mock.Mock
}

func (client *MockSearchClient) Reindex(indexName string, data []map[string]interface{}) error {
	args := client.Called(indexName, data)
	return args.Error(0)
}

func (client *MockSearchClient) SaveObject(indexName string, data map[string]interface{}) error {
	args := client.Called(indexName, data)
	return args.Error(0)
}

func (client *MockSearchClient) Clear(objectID string) error {
	args := client.Called(objectID)
	return args.Error(0)
}
