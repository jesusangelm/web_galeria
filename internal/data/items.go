package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	filestorage "jesusmarin.dev/galeria/internal/file_storage"
)

type Item struct {
	ID             int64
	Name           string
	Description    string
	CreatedAt      time.Time
	CategoryID     int64
	CategoryName   string
	ImageURL       string
	ItemAttachment ItemAttachment
}

type ItemModel struct {
	DB        *pgxpool.Pool
	S3Manager filestorage.S3
}

// Return a single item based on the ID given
func (m *ItemModel) Get(id int64) (*Item, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT
			items.id, items.name, items.description, items.created_at,
			items.category_id, categories.name AS category_name,
			COALESCE(item_attachments.key, '') as key,
			COALESCE(item_attachments.filename, '') as filename
		FROM items
		INNER JOIN categories ON categories.id = items.category_id
		LEFT JOIN item_attachments ON items.id = item_attachments.item_id
		WHERE items.id = $1
	`

	var item Item

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.CreatedAt,
		&item.CategoryID,
		&item.CategoryName,
		&item.ItemAttachment.Key,
		&item.ItemAttachment.Filename,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	url := m.S3Manager.ProxyImageUrl(item.ItemAttachment.Key)
	item.ImageURL = url

	return &item, nil
}
