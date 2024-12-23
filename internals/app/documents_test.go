package app

import (
	"encoding/json"
	"loshon-api/internals/data"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		if assert.Error(t, app.GetDocuments(ectx)) {
			err, ok := app.GetDocuments(ectx).(*echo.HTTPError)
			assert.True(t, ok)
			assert.Equal(t, http.StatusInternalServerError, err.Code)
		}
	})
}
