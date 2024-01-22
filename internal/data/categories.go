package data

import (
	"context"
	"errors"
	"fmt"
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
func (m *CategoryModel) Get(id int64, filters Filters) (*Category, Metadata, error) {
	if id < 1 {
		return nil, Metadata{}, ErrRecordNotFound
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
			return nil, Metadata{}, ErrRecordNotFound
		default:
			return nil, Metadata{}, err
		}
	}

	// query to get the items in a given category
	query = fmt.Sprintf(`
		SELECT count(*) OVER(), items.id, items.name, items.description, items.created_at,
				COALESCE(item_attachments.key, '') as key,
				COALESCE(item_attachments.filename, '') as filename
		FROM items
		LEFT JOIN item_attachments ON items.id = item_attachments.item_id
		WHERE items.category_id = $1
		ORDER BY %s %s, id ASC
		LIMIT $2
		OFFSET $3
	`, filters.sortColumn(), filters.sortDirection())

	args := []any{id, filters.limit(), filters.offset()}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var items []*Item
	totalRecords := 0

	for rows.Next() {
		var item Item
		err := rows.Scan(
			&totalRecords,
			&item.ID,
			&item.Name,
			&item.Description,
			&item.CreatedAt,
			&item.ItemAttachment.Key,
			&item.ItemAttachment.Filename,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		url := m.S3Manager.ProxyImageUrl(item.ItemAttachment.Key)
		item.ImageURL = url

		items = append(items, &item)
	}
	category.Items = items

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return &category, metadata, nil
}

// Return a slice of categories.
func (m *CategoryModel) List(name string, filters Filters) ([]*Category, Metadata, error) {
	query := fmt.Sprintf(`
    SELECT
      count(*) OVER(), categories.id, categories.name, categories.description,
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
  	WHERE (to_tsvector('simple', categories.name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		GROUP BY categories.id
		ORDER BY %s %s, id ASC
		LIMIT $2
		OFFSET $3
  `, filters.sortColumn(), filters.sortDirection())

	// 3 seconds timeout for quering the DB
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, filters.limit(), filters.offset()}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var categories []*Category

	for rows.Next() {
		var category Category
		err := rows.Scan(
			&totalRecords,
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.ItemsCount,
			&category.ImageKey,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		category.ImageURL = m.S3Manager.ProxyImageUrl(category.ImageKey)

		categories = append(categories, &category)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return categories, metadata, nil
}
