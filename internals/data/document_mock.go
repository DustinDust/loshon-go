package data

import "github.com/stretchr/testify/mock"

type MockDocumentRepository struct {
	mock.Mock
}

func (repo *MockDocumentRepository) Get(query interface{}, _ ...any) ([]Document, error) {
	args := repo.Called(query)
	err := args.Get(1)
	if err != nil {
		return args.Get(0).([]Document), args.Error(1)
	} else {
		return args.Get(0).([]Document), nil
	}
}

func (repo *MockDocumentRepository) First(query interface{}, _ ...any) (*Document, error) {
	args := repo.Called(query)
	return args.Get(0).(*Document), args.Error(1)
}

func (repo *MockDocumentRepository) Save(doc *Document) error {
	args := repo.Called(doc)
	return args.Error(0)
}

func (repo *MockDocumentRepository) Delete(doc *Document) error {
	args := repo.Called(doc)
	return args.Error(0)
}

func (repo *MockDocumentRepository) Archive(doc *Document) error {
	args := repo.Called(doc)
	return args.Error(0)
}

func (repo *MockDocumentRepository) Restore(doc *Document) error {
	args := repo.Called(doc)
	return args.Error(0)
}
