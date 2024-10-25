package data

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Document struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();index" json:"id"`
	Title            string         `json:"title"`
	UserID           string         `gorm:"index" json:"userId"`
	IsArchived       bool           `gorm:"default=false" json:"isArchived"`
	IsPublished      bool           `gorm:"default=false" json:"isPublished"`
	ParentDocumentID *string        `gorm:"index,type:uuid" json:"parentDocumentId"`
	ChildDocuments   []Document     `gorm:"foreignKey:ParentDocumentID" json:"-"`
	Content          *string        `json:"content"`
	CoverImage       *string        `json:"coverImage"`
	Icon             *string        `json:"icon"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deletedAt"`
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

func (doc *Document) ArchiveRecursively(db *gorm.DB) error {
	statement := `
	WITH RECURSIVE d AS (
  	SELECT documents.id
   		FROM documents
   		WHERE documents.id = ?
 		UNION ALL
  	SELECT child.id
  		FROM d JOIN documents child ON child.parent_document_id = d.id
		)
	UPDATE documents b set is_archived = true
 		FROM d
 		WHERE d.id = b.id
	`

	if err := db.Exec(statement, doc.ID).Error; err != nil {
		return err
	}

	// not really required, but It would be cleaner to reload the state of archived object
	if err := db.Preload("ChildDocuments").Find(doc).Error; err != nil {
		slog.Warn("error realoading object", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
	return nil
}

func (doc *Document) DeleteRecursively(db *gorm.DB) error {
	statement := `
	WITH RECURSIVE d AS (
  	SELECT documents.id
   		FROM documents
   		WHERE documents.id = ?
 		UNION ALL
  	SELECT child.id
  		FROM d JOIN documents child ON child.parent_document_id = d.id
		)
    UPDATE documents b set deleted_at = NOW()::TIMESTAMP
 		FROM d
 		WHERE d.id = b.id
	`
	if err := db.Exec(statement, doc.ID).Error; err != nil {
		return err
	}
	if err := db.Preload("ChildDocuments").Find(doc).Error; err != nil {
		slog.Warn("error realoading object", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
	return nil
}

func (doc *Document) RestoreRecursively(db *gorm.DB) error {
	doc.IsArchived = false
	if doc.ParentDocumentID != nil {
		parentDocument := Document{}
		if err := db.First(&parentDocument, "id = ?", doc.ParentDocumentID).Error; err != nil {
			return err
		}
		if err := parentDocument.RestoreRecursively(db); err != nil {
			return err
		}
	}
	return db.Save(&doc).Error
}

func (doc *Document) SetParentDocument(parentDocumentID Optional[string]) {
	if parentDocumentID.Defined {
		doc.ParentDocumentID = parentDocumentID.Value
	}
}

func (doc *Document) SetTitle(title Optional[string]) {
	if title.Defined {
		doc.Title = *title.Value
	}
}

func (doc *Document) SetContent(content Optional[string]) {
	if content.Defined {
		doc.Content = content.Value
	}
}

func (doc *Document) SetCoverImage(coverImage Optional[string]) {
	if coverImage.Defined {
		doc.CoverImage = coverImage.Value
	}
}

func (doc *Document) SetIcon(icon Optional[string]) {
	if icon.Defined {
		doc.Icon = icon.Value
	}
}

func (doc *Document) SetIsArchived(isArchived Optional[bool]) {
	if isArchived.Defined {
		doc.IsArchived = *isArchived.Value
	}
}

func (doc *Document) SetIsPublished(isPublished Optional[bool]) {
	if isPublished.Defined {
		doc.IsPublished = *isPublished.Value
	}
}
