package data

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ItemAttachment struct {
	ID          int64
	Key         string
	Filename    string
	ContentType string
	ByteSize    int64
	CreatedAt   time.Time
	ItemID      int64
}

type ItemAttachmentModel struct {
	DB *pgxpool.Pool
}

func (m *ItemAttachmentModel) Insert(itemAttachment *ItemAttachment) error {
	query := `
		INSERT INTO item_attachments (filename, content_type, byte_size, item_id)
		VALUES($1, $2, $3, $4)
		RETURNING id, created_at
	`

	args := []interface{}{
		itemAttachment.Filename,
		itemAttachment.ContentType,
		itemAttachment.ByteSize,
		itemAttachment.ItemID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRow(ctx, query, args...).Scan(
		&itemAttachment.ID,
		&itemAttachment.CreatedAt,
	)
}
