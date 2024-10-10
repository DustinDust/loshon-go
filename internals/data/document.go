package data

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();index" json:"id"`
	Title          string         `json:"title"`
	UserID         string         `gorm:"index" json:"userId"`
	IsArchived     bool           `gorm:"default=false" json:"isArchived"`
	IsPublished    bool           `gorm:"default=false" json:"isPublished"`
	ParentDocument *string        `gorm:"index" json:"parentDocument"`
	Content        *string        `json:"content"`
	CoverImage     *string        `json:"coverImage"`
	Icon           *string        `json:"icon"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (doc *Document) MarshalJSON() ([]byte, error) {
	type Alias Document
	var deletedAt *string
	if doc.DeletedAt.Valid {
		utcDeletedAt := doc.DeletedAt.Time.UTC().Format(time.RFC3339)
		deletedAt = &utcDeletedAt
	}

	return json.Marshal(&struct {
		*Alias
		CreatedAt string  `json:"createdAt"`
		UpdatedAt string  `json:"updatedAt"`
		DeletedAt *string `json:"deletedAt"`
	}{
		Alias:     (*Alias)(doc),
		CreatedAt: doc.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: doc.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt: deletedAt,
	})
}
