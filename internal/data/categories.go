package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	filestorage "jesusmarin.dev/galeria/internal/file_storage"
)

// struct to represent the Category model
type Category struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
	Items       []*Item
	ItemsCount  int64
	ImageKey    string
	ImageURL    string
}

type CategoryModel struct {
	DB        *pgxpool.Pool
	S3Manager filestorage.S3
}

// Return a single category based on the ID given
func (m *CategoryModel) Get(id int64) (*Category, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT
			categories.id, categories.name, categories.description,
			categories.created_at, COUNT(items.id) AS items_count
		FROM categories
		LEFT JOIN items ON categories.id = items.category_id
		WHERE categories.id = $1
		GROUP BY categories.id
	`
	var category Category

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.ItemsCount,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// query to get the items in a given category
	query = `
		SELECT items.id, items.name, items.description, items.created_at,
				COALESCE(item_attachments.key, '') as key,
				COALESCE(item_attachments.filename, '') as filename
		FROM items
		LEFT JOIN item_attachments ON items.id = item_attachments.item_id
		WHERE items.category_id = $1
		ORDER BY items.created_at DESC
	`
	rows, err := m.DB.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CreatedAt,
			&item.ItemAttachment.Key,
			&item.ItemAttachment.Filename,
		)
		if err != nil {
			return nil, err
		}

		url := m.S3Manager.ProxyImageUrl(item.ItemAttachment.Key)
		item.ImageURL = url

		items = append(items, &item)
	}
	category.Items = items

	return &category, nil
}

// Return a slice of categories.
func (m *CategoryModel) List() ([]*Category, error) {
	// TODO: Add pagination support.
	query := `
    SELECT
      categories.id, categories.name, categories.description,
      categories.created_at, COUNT(items.id) AS items_count,
      COALESCE(
        (SELECT item_attachments.key
          FROM items
					LEFT JOIN item_attachments ON items.id = item_attachments.item_id
					WHERE items.category_id = categories.id
					limit 1
        ), ''
      ) as image_key
		FROM categories
		LEFT JOIN items ON categories.id = items.category_id
		LEFT JOIN item_attachments ON items.id = item_attachments.item_id
		GROUP BY categories.id
		ORDER BY categories.created_at DESC
  `
	// 3 seconds timeout for quering the DB
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var category Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.ItemsCount,
			&category.ImageKey,
		)
		if err != nil {
			return nil, err
		}

		category.ImageURL = m.S3Manager.ProxyImageUrl(category.ImageKey)

		categories = append(categories, &category)
	}

	return categories, nil
}
