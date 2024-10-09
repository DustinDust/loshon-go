package app

import (
	"errors"
	"loshon-api/internals/data"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (a App) healthCheck(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"ping": "pong",
	})
}

func (app App) GetDocuments(c echo.Context) error {
	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid context")
	}
	var parentDocument interface{}
	if parentId := c.Get("parent_id"); parentId != "" {
		parentDocument = parentId
	} else {
		parentDocument = nil
	}

	documents := []data.Document{}

	results := app.db.
		Where("user_id", user.ID).
		Where("parent_document", parentDocument).
		Where("is_archived", false).
		Find(&documents)

	if results.Error != nil {
		switch {
		case errors.Is(results.Error, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, results.Error.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, results.Error.Error())
		}
	}
	return c.JSON(http.StatusOK, Response[[]data.Document]{
		Data: documents,
	})
}
