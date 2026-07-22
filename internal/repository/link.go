package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezect/url-shortener/internal/models"
)

var (
	ErrNoTransaction = errors.New("transaction is not set")
)

type LinkRepository struct {
	url  string
	pool *pgxpool.Pool
	conn Querier
}

func NewLinkRepository(ctx context.Context, url string, pool *pgxpool.Pool) *LinkRepository {
	return &LinkRepository{
		url:  url,
		pool: pool,
		conn: pool,
	}
}

func (db *LinkRepository) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *LinkRepository) WithTx(tx pgx.Tx) *LinkRepository {
	dbCopy := *db
	dbCopy.conn = tx
	return &dbCopy
}

func (db *LinkRepository) Stop() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *LinkRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	var isAlreadyExists int
	err := db.conn.QueryRow(ctx, "SELECT COUNT(*) FROM short_links WHERE short_code=$1", shortCode).Scan(&isAlreadyExists)
	if err != nil {
		return false, fmt.Errorf("Error while get info from database: %v", err)
	}
	if isAlreadyExists != 0 {
		return true, nil
	}

	return false, nil
}

func (db *LinkRepository) Get(ctx context.Context, alias string) (*models.ShortLink, error) {
	var link models.ShortLink
	err := db.conn.QueryRow(ctx, "SELECT id, short_code, original_url, created_at, expires_at FROM short_links WHERE short_code=$1", alias).Scan(
		&link.Id,
		&link.ShortCode,
		&link.OriginalUrl,
		&link.CreatedAt,
		&link.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (db *LinkRepository) Create(ctx context.Context, originalUrl string, shortCode string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error) {
	var insertedCreatedAt time.Time
	var err error
	if createdAt != nil {
		err = db.conn.QueryRow(ctx, "INSERT INTO short_links (short_code, original_url, created_at, expires_at) VALUES	($1, $2, $3, $4) RETURNING created_at", shortCode, originalUrl, createdAt, expiresAt).Scan(&insertedCreatedAt)
	} else {
		err = db.conn.QueryRow(ctx, "INSERT INTO short_links (short_code, original_url, expires_at) VALUES	($1, $2, $3) RETURNING created_at", shortCode, originalUrl, expiresAt).Scan(&insertedCreatedAt)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("Error while inserting values in database: %v", err.Error())
	}

	return insertedCreatedAt, nil
}

func (db *LinkRepository) Delete(ctx context.Context, shortCode string) error {
	_, err := db.conn.Exec(ctx, "DELETE FROM short_links WHERE short_code = $1", shortCode)
	if err != nil {
		return fmt.Errorf("Error while deleting code \"%v\" from database: %v", shortCode, err)
	}

	return nil
}
