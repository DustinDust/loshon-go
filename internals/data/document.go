package data

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TYPEDEF Documents
type Document struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();index" json:"id"`
	Title            string         `json:"title"`
	UserID           string         `gorm:"index" json:"userId"`
	IsArchived       bool           `gorm:"default=false" json:"isArchived"`
	IsPublished      bool           `gorm:"default=false" json:"isPublished"`
	ParentDocumentID *string        `gorm:"index,type:uuid" json:"parentDocumentId"`
	ChildDocuments   []Document     `gorm:"foreignKey:ParentDocumentID" json:"-"`
	Content          *string        `json:"content"`
	MdContent        *string        `json:"mdContent"` // for full text search only
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
func (doc *Document) SetMdContent(mdContent Optional[string]) {
	if mdContent.Defined {
		doc.MdContent = mdContent.Value
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

func (doc Document) ToSearchObject() map[string]any {
	return map[string]any{
		"objectID":   doc.ID.String(),
		"userId":     doc.UserID,
		"title":      doc.Title,
		"content":    doc.MdContent,
		"isArchived": doc.IsArchived,
		"createdAt":  doc.CreatedAt,
		"updatedAt":  doc.UpdatedAt,
		"deletedAt":  doc.DeletedAt,
	}
}

// DOCUMENT MODEL AND IMPLEMENTATION
type DocumentRepositoryInterface interface {
	Save(*Document) error
	Delete(*Document) error
	Archive(*Document) error
	Restore(*Document) error
	Get(interface{}, ...any) ([]Document, error)
	First(interface{}, ...any) (*Document, error)
}

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return DocumentRepository{
		db: db,
	}
}

func (repo DocumentRepository) Save(doc *Document) error {
	if err := repo.db.Save(doc).Error; err != nil {
		return err
	}
	return nil
}

func (repo DocumentRepository) Get(query interface{}, args ...any) ([]Document, error) {
	documents := make([]Document, 0)
	if err := repo.db.Where(query, args).Find(&documents).Order("created_at asc").Error; err != nil {
		return documents, err
	}
	return documents, nil
}

func (repo DocumentRepository) First(query interface{}, args ...any) (*Document, error) {
	var document Document
	if err := repo.db.First(document, query, args).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

func (repo DocumentRepository) Archive(doc *Document) error {
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

	if err := repo.db.Exec(statement, doc.ID).Error; err != nil {
		return err
	}

	// not really required, but It would be cleaner to reload the state of archived object
	if err := repo.db.Preload("ChildDocuments").Find(doc).Error; err != nil {
		slog.Warn("error realoading object", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
	return nil
}

func (repo DocumentRepository) Delete(doc *Document) error {
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
	if err := repo.db.Exec(statement, doc.ID).Error; err != nil {
		return err
	}
	if err := repo.db.Preload("ChildDocuments").Find(doc).Error; err != nil {
		slog.Warn("error realoading object", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}
	return nil
}

func (repo DocumentRepository) Restore(doc *Document) error {
	doc.IsArchived = false
	if doc.ParentDocumentID != nil {
		parentDocument := Document{}
		if err := repo.db.First(&parentDocument, "id = ?", doc.ParentDocumentID).Error; err != nil {
			return err
		}
		if err := repo.Restore(&parentDocument); err != nil {
			return err
		}
	}
	return repo.db.Save(doc).Error
}
