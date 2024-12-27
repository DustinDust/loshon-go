package app

import (
	"encoding/json"
	"loshon-api/internals/config"
	"loshon-api/internals/data"
	"loshon-api/internals/search"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// https://dwarvesf.hashnode.dev/unit-testing-best-practices-in-golang#heading-table-driven-testing

func TestGetDocuments(t *testing.T) {
	os.Setenv("ENV", "test")
	e := echo.New()
	app := &App{
		engine: e,
	}

	t.Run("successful", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("Get", mock.Anything).Return(
			[]data.Document{
				{ID: uuid.New(), Title: "Test Document"},
				{ID: uuid.New(), Title: "Test Document 2"},
			},
			nil,
		)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.NoError(t, app.GetDocuments(ectx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var respJSON Response[[]data.Document]
			if err := json.Unmarshal(rec.Body.Bytes(), &respJSON); err != nil {
				t.Errorf("Error unmarshalling response: %v", err)
			}
			assert.Equal(t, 2, respJSON.Total)
			assert.Equal(t, "Test Document", respJSON.Data[0].Title)
			assert.Equal(t, "Test Document 2", respJSON.Data[1].Title)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/documents", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)

		if assert.Error(t, app.GetDocuments(ectx)) {
			assert.ErrorIs(t, app.GetDocuments(ectx), echo.ErrUnauthorized)
		}
	})

	t.Run("error", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("Get", mock.Anything).Return([]data.Document{}, assert.AnError)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		err := app.GetDocuments(ectx)

		if assert.Error(t, err) {
			assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
		}
	})
}

func TestGetDocumentByID(t *testing.T) {
	os.Setenv("ENV", "test")
	e := echo.New()
	app := &App{
		engine: e,
	}

	t.Run("successful", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("First", mock.Anything).Return(&data.Document{ID: uuid.New(), Title: "test one", UserID: "1"}, nil)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents/1", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.NoError(t, app.GetDocumentByID(ectx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var respJSON Response[data.Document]
			if err := json.Unmarshal(rec.Body.Bytes(), &respJSON); err != nil {
				t.Errorf("error unmarshalling response: %v", err)
			}
			assert.Equal(t, "test one", respJSON.Data.Title)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("First", mock.Anything).Return(&data.Document{}, gorm.ErrRecordNotFound)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents/909090", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		err := app.GetDocumentByID(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusNotFound, err.(*echo.HTTPError).Code)
		}
	})

	t.Run("public post", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("First", mock.Anything).Return(&data.Document{ID: uuid.New(), Title: "public", IsPublished: true}, nil)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents/1", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.NoError(t, app.GetDocumentByID(ectx)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			var respJSON Response[data.Document]
			if err := json.Unmarshal(rec.Body.Bytes(), &respJSON); err != nil {
				t.Errorf("error unmarshalling response: %v", err)
			}
			assert.Equal(t, "public", respJSON.Data.Title)
		}
	})

	t.Run("different user", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("First", mock.Anything).Return(&data.Document{ID: uuid.New(), UserID: "1"}, nil)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents/1", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "2"})

		err := app.GetDocumentByID(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusForbidden, err.(*echo.HTTPError).Code)
		}
	})
	t.Run("no user", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("First", mock.Anything).Return(&data.Document{ID: uuid.New(), UserID: "1"}, nil)
		app.documentRepo = mockRepo
		req := httptest.NewRequest(http.MethodGet, "/documents/1", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)

		err := app.GetDocumentByID(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		}
	})

}

func TestCreateDocument(t *testing.T) {
	os.Setenv("ENV", "test")
	e := echo.New()
	app := &App{
		engine: e,
	}

	t.Run("unauthorized", func(t *testing.T) {
		body := `{"title": "example_title","content": "lorem ipsum"}`
		req := httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(body))
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)

		err := app.CreateDocument(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		}
	})

	t.Run("bad request", func(t *testing.T) {
		body := `{"title": 1,"content": "lorem ipsum"}`

		req := httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(body))
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.Error(t, app.CreateDocument(ectx)) {
			err, isEchoError := app.CreateDocument(ectx).(*echo.HTTPError)
			assert.True(t, isEchoError)
			assert.Equal(t, http.StatusBadRequest, err.Code)
		}
	})

	t.Run("unexpected db error", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("Save", mock.AnythingOfType("*data.Document")).Return(assert.AnError)
		app.documentRepo = mockRepo

		body := `{"title":"example_title","content":"lorem ipsum"}`
		req := httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		err := app.CreateDocument(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
		}
	})

	t.Run("sucess", func(t *testing.T) {
		mockRepo := new(data.MockDocumentRepository)
		mockRepo.On("Save", mock.AnythingOfType("*data.Document")).Return(nil)
		mockSearch := new(search.MockSearchClient)
		mockSearch.On("SaveObject", mock.Anything, mock.Anything).Return(nil)
		app.documentRepo = mockRepo
		app.sclient = mockSearch
		app.config = &config.AppConfig{SearchIndex: "test_index"}

		body := `{"title":"example_title","content":"lorem ipsum"}`
		req := httptest.NewRequest(http.MethodPost, "/documents", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.NoError(t, app.CreateDocument(ectx)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			mockRepo.AssertCalled(t, "Save", mock.AnythingOfType("*data.Document"))
			mockSearch.AssertCalled(t, "SaveObject", "test_index", mock.Anything)
		}
	})
}

func TestUpdateDocument(t *testing.T) {
	e := echo.New()
	app := &App{
		engine: e,
	}
	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/documents/1", nil)
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)

		err := app.UpdateDocument(ectx)
		if assert.Error(t, err) {
			assert.Equal(t, http.StatusUnauthorized, err.(*echo.HTTPError).Code)
		}
	})

	t.Run("bad request", func(t *testing.T) {
		body := `{"title": 1,"content": "lorem ipsum"}`

		req := httptest.NewRequest(http.MethodPatch, "/documents/1", strings.NewReader(body))
		rec := httptest.NewRecorder()
		ectx := e.NewContext(req, rec)
		ectx.Set("user", &clerk.User{ID: "1"})

		if assert.Error(t, app.UpdateDocument(ectx)) {
			err, isEchoError := app.UpdateDocument(ectx).(*echo.HTTPError)
			assert.True(t, isEchoError)
			assert.Equal(t, http.StatusBadRequest, err.Code)
		}
	})
}
