package models

import "time"

type ShortLink struct {
	Id          int64     `db:"id"`
	ShortCode   string    `db:"short_code"`
	OriginalUrl string    `db:"original_url"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}
