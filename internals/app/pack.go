package app

import "loshon-api/internals/data"

type Response[T any] struct {
	Data  T   `json:"data"`
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}

type CreateDocumentRequest struct {
	Title            string  `json:"title" validate:"required,min=2"`
	IsArchived       bool    `json:"isArchived"`
	IsPublished      bool    `json:"isPublished"`
	ParentDocumentID *string `json:"parentDocumentId"`
	Content          *string `json:"content" `
	CoverImage       *string `json:"coverImage"`
	Icon             *string `json:"icon"`
}

type UpdateDocumentRequest struct {
	ID               string                `json:"id" validate:"required,uuid"`
	Title            data.Optional[string] `json:"title"`
	IsArchived       data.Optional[bool]   `json:"isArchived"`
	IsPublished      data.Optional[bool]   `json:"isPublished"`
	ParentDocumentID data.Optional[string] `json:"parentDocumentId"`
	Content          data.Optional[string] `json:"content" `
	CoverImage       data.Optional[string] `json:"coverImage"`
	Icon             data.Optional[string] `json:"icon"`
}
