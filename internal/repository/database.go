package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoTransaction = errors.New("transaction is not set")
)

type Database struct {
	url  string
	pool *pgxpool.Pool
	conn Querier
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func (db *Database) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}

func (db *Database) WithTx(tx pgx.Tx) *Database {
	dbCopy := *db
	dbCopy.conn = tx
	return &dbCopy
}

func (db *Database) Stop() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func NewDatabase(url string, ctx context.Context) (*Database, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to database: %v", err)
	}

	db := &Database{
		url:  url,
		pool: pool,
		conn: pool,
	}

	return db, nil
}

func (db *Database) Exists(ctx context.Context, shortCode string) (bool, error) {
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

func (db *Database) Create(ctx context.Context, originalUrl string, shortCode string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error) {
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

func (db *Database) Delete(ctx context.Context, shortCode string) error {
	_, err := db.conn.Exec(ctx, "DELETE FROM short_links WHERE short_code = $1", shortCode)
	if err != nil {
		return fmt.Errorf("Error while deleting code \"%v\" from database: %v", shortCode, err)
	}

	return nil
}
