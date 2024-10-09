package data

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	*gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();index" json:"id"`
	Title          string    `json:"title"`
	UserID         string    `gorm:"index" json:"uesrId"`
	IsArchived     bool      `gorm:"default=false" json:"isArchived"`
	ParentDocument *string   `gorm:"index" json:"parentDocument"`
	Content        *string   `json:"content"`
	CoverImage     *string   `json:"coverImage"`
	Icon           *string   `json:"icon"`
	IsPublished    bool      `gorm:"default=false" json:"isPublished"`
}
