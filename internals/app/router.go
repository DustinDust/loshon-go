package app

import (
	"errors"
	"loshon-api/internals/data"
	"loshon-api/internals/validator"
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
	if parentId := c.QueryParam("parentDocument"); parentId != "" {
		parentDocument = parentId
	} else {
		parentDocument = nil
	}

	documents := []data.Document{}

	result := app.db.
		Where("user_id", user.ID).
		Where("parent_document_id", parentDocument).
		Where("is_archived", false).
		Find(&documents)

	if result.Error != nil {
		switch {
		case errors.Is(result.Error, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, result.Error.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, result.Error.Error())
		}
	}
	return c.JSON(http.StatusOK, Response[[]data.Document]{
		Data:  documents,
		Total: int(result.RowsAffected),
	})
}

func (app App) CreateDocument(c echo.Context) error {
	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid context object")
	}
	createData := CreateDocumentRequest{}
	if err := c.Bind(&createData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}
	v := validator.NewValidator()
	if err := v.ValidateStruct(createData); err != nil {
		if verr, ok := err.(*validator.StructValidationErrors); ok {
			return verr.TranslateToHttpError()
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	document := data.Document{
		Title:            createData.Title,
		UserID:           user.ID,
		IsArchived:       createData.IsArchived,
		IsPublished:      createData.IsPublished,
		ParentDocumentID: createData.ParentDocumentID,
		Content:          createData.Content,
		CoverImage:       createData.CoverImage,
		Icon:             createData.Icon,
	}
	result := app.db.Create(&document)
	if result.Error != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, result.Error)
	}
	return c.JSON(http.StatusOK, Response[data.Document]{
		Data: document,
	})
}

func (app App) ArchiveDocument(c echo.Context) error {
	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid context")
	}

	documentID := c.Param("documentID")
	var document data.Document
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
	if err := document.ArchiveAll(app.db); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, echo.Map{})
}
