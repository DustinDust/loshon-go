package app

import (
	"errors"
	"loshon-api/internals/data"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (app App) ArchiveDocument(c echo.Context) error {
	var user *clerk.User
	var document *data.Document
	var archivedDocuments []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid context")
	}

	documentID := c.Param("documentID")
	document, err := app.documentRepo.First("id = ?", documentID)
	if err != nil {
		switch {
		case (errors.Is(err, gorm.ErrRecordNotFound)):
			return echo.NewHTTPError(http.StatusNotFound, err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	if user.ID != document.UserID {
		return echo.ErrForbidden
	}
	if err := app.documentRepo.Archive(document); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all archived documents
	archivedDocuments, err = app.documentRepo.Get(map[string]any{"user_id": user.ID, "is_archived": true})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, ad := range archivedDocuments {
		reindexObjects = append(reindexObjects, ad.ToSearchObject())
	}
	if err := app.sclient.Reindex("documents", reindexObjects); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{})
}

func (app App) RestoreArchivedDocument(c echo.Context) error {
	var user *clerk.User
	var document *data.Document
	var documents []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	documentID := c.Param("documentID")
	document, err := app.documentRepo.First(map[string]any{"id": documentID})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	if document.UserID != user.ID {
		return echo.ErrForbidden
	}
	if err := app.documentRepo.Restore(document); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all unarchived documents
	documents, err = app.documentRepo.Get(map[string]any{"user_id": user.ID, "is_archived": false})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, ad := range documents {
		reindexObjects = append(reindexObjects, ad.ToSearchObject())
	}
	if err := app.sclient.Reindex("documents", reindexObjects); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{})
}

func (app App) DeleteArchivedDocument(c echo.Context) error {
	var user *clerk.User
	var document *data.Document
	var deletedDocuments []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	documentID := c.Param("documentID")
	document, err := app.documentRepo.First(map[string]any{"id": documentID})
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err)

		}
	}
	if user.ID != document.UserID {
		return echo.ErrForbidden
	}
	if err := app.documentRepo.Delete(document); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all archived documents
	deletedDocuments, err = app.documentRepo.Get("user_id = ? AND deleted IS NOT NULL", user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	for _, ad := range deletedDocuments {
		reindexObjects = append(reindexObjects, ad.ToSearchObject())
	}
	if err := app.sclient.Reindex("documents", reindexObjects); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, echo.Map{})
}

func (app App) GetArchivedDocuments(c echo.Context) error {
	var user *clerk.User
	var documents []data.Document

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	documents, err := app.documentRepo.Get("user_id=? AND is_archive=true", user.ID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	return c.JSON(http.StatusOK, Response[[]data.Document]{
		Data:  documents,
		Total: int(len(documents)),
	})
}
