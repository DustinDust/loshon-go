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

func (app App) GetDocuments(c echo.Context) error {
	var user (*clerk.User)
	var parentDocument interface{}
	var documents []data.Document

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}
	if parentId := c.QueryParam("parentDocument"); parentId != "" {
		parentDocument = parentId
	} else {
		parentDocument = nil
	}

	documents, err := app.documentRepo.Get(
		map[string]interface{}{
			"user_id":            user.ID,
			"parent_document_id": parentDocument,
			"is_archived":        false,
		},
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.JSON(http.StatusOK, Response[[]data.Document]{
		Data:  documents,
		Total: int(len(documents)),
	})
}

func (app App) GetDocumentByID(c echo.Context) error {
	var user (*clerk.User)
	var document *data.Document

	documentID := c.Param("documentID")
	document, err := app.documentRepo.First("id = ?", documentID)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, err)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}
	if document.IsPublished && !document.IsArchived {
		return c.JSON(http.StatusOK, Response[data.Document]{
			Data: *document,
		})
	}

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	if document.UserID != user.ID {
		return echo.ErrForbidden
	}

	return c.JSON(http.StatusOK, Response[data.Document]{
		Data: *document,
	})
}

func (app App) CreateDocument(c echo.Context) error {
	var user *clerk.User
	createData := CreateDocumentRequest{}
	v := validator.NewValidator()

	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(&createData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}

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
	err := app.documentRepo.Save(&document)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	app.sclient.SaveObject(app.config.SearchIndex, document.ToSearchObject())
	return c.JSON(http.StatusOK, Response[data.Document]{
		Data: document,
	})
}

func (app App) UpdateDocument(c echo.Context) error {
	var user *clerk.User
	var document *data.Document

	updateData := UpdateDocumentRequest{
		ID: c.Param("documentID"),
	}
	v := validator.NewValidator()
	user, ok := c.Get("user").(*clerk.User)
	if !ok {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(&updateData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request object")
	}
	if err := v.ValidateStruct(updateData); err != nil {
		if verr, ok := err.(*validator.StructValidationErrors); ok {
			return verr.TranslateToHttpError()
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
	}

	document, err := app.documentRepo.First("id = ?", updateData.ID)
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

	// patch attributes
	document.SetTitle(updateData.Title)
	document.SetContent(updateData.Content)
	document.SetMdContent(updateData.MdContent)
	document.SetParentDocument(updateData.ParentDocumentID)
	document.SetIcon(updateData.Icon)
	document.SetCoverImage(updateData.CoverImage)
	document.SetIsPublished(updateData.IsPublished)
	document.SetIsArchived(updateData.IsArchived)

	if err := app.documentRepo.Save(document); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	app.sclient.SaveObject(app.config.SearchIndex, document.ToSearchObject())

	return c.JSON(http.StatusOK, Response[data.Document]{
		Data: *document,
	})
}
