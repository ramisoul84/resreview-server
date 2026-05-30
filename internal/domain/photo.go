package domain

import "time"

type Photo struct {
	PhotoID   string    `db:"photo_id"`
	VersionID string    `db:"version_id"`
	URL       string    `db:"url"`
	Key       string    `db:"key"`
	CreatedAt time.Time `db:"created_at"`
}
