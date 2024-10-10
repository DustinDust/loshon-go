package app

type Response[T any] struct {
	Data  T   `json:"data"`
	Page  int `json:"page,omitempty"`
	Total int `json:"total,omitempty"`
}

type CreateDocumentRequest struct {
	Title          string  `json:"title" validate:"required,min=2"`
	IsArchived     bool    `json:"isArchived"`
	IsPublished    bool    `json:"isPublished"`
	ParentDocument *string `json:"parentDocument"`
	Content        *string `json:"content" `
	CoverImage     *string `json:"coverImage"`
	Icon           *string `json:"icon"`
}
