package models

import "time"

type Click struct {
	Id        int64     `db:"id"`
	ShortCode string    `db:"short_code"`
	Ip        string    `db:"ip"`
	UserAgent string    `db:"user_agent"`
	Referer   string    `db:"referer"`
	ClickedAt time.Time `db:"clicked_at"`
}
