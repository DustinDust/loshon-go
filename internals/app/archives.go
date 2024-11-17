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
	var document data.Document
	var archivedDocuments []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid context")
	}

	documentID := c.Param("documentID")
	if err := app.db.First(&document, "id = ?", documentID).Error; err != nil {
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
	if err := document.ArchiveRecursively(app.db); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all archived documents
	if err := app.db.Where("user_id=?", user.ID).Where("is_archived=?", true).Find(&archivedDocuments).Error; err != nil {
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
	var document data.Document
	var documents []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	documentID := c.Param("documentID")
	if err := app.db.First(&document, "id = ?", documentID).Error; err != nil {
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
	if err := document.RestoreRecursively(app.db); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all archived documents
	if err := app.db.Where("user_id=?", user.ID).Where("is_archived=?", false).Find(&documents).Error; err != nil {
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
	var document data.Document
	var deletedDocuments []data.Document
	var reindexObjects []map[string]any

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	documentID := c.Param("documentID")
	if err := app.db.First(&document, "id = ?", documentID).Error; err != nil {
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
	if err := document.DeleteRecursively(app.db); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// reindex all archived documents
	if err := app.db.Where("user_id=?", user.ID).Where("deleted_at IS NOT NULL").Find(&deletedDocuments).Error; err != nil {
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

	result := app.db.Where("user_id=?", user.ID).Where("is_archived=?", true).Find(&documents)
	if err := result.Error; err != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, result.Error)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
		}
	}
	return c.JSON(http.StatusOK, Response[[]data.Document]{
		Data:  documents,
		Total: int(result.RowsAffected),
	})
}
