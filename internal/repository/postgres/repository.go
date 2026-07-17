package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	URL string
	Pool *pgxpool.Pool
	Ctx context.Context
}

func NewDatabase(url string, ctx context.Context) (*Database, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("Error while connecting to database: %v", err)
	}

	db := &Database{
		URL: url,
		Pool: pool,
		Ctx: ctx,
	}
	
	return db, nil
}

func (db *Database) isLinkExists(shortCode string) (bool, error) {
	var isAlreadyExists int
	err := db.Pool.QueryRow(db.Ctx, "SELECT COUNT(*) FROM short_links WHERE short_code=$1", shortCode).Scan(&isAlreadyExists)
	if err != nil {
		return false, fmt.Errorf("Error while get info from database: %v", err)
	}
	if isAlreadyExists != 0 {
		return true, nil
	}

	return false, nil
}

func generateShortCode(length int) string {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range length {
		randIdx := rand.Intn(len(alphabet))
		randChar := alphabet[randIdx]
		b[i] = randChar
	}

	return string(b)
}

func (db *Database) Create(originalUrl string, createdAt time.Time, expiresAt *time.Time, shortCode string) (string, error) {
	if shortCode != "" {
		isNewCodeExists, err := db.isLinkExists(shortCode)
		if err != nil {
			return "", err
		}
		if isNewCodeExists {
			return "", fmt.Errorf("Link with code %v is already exists", shortCode)
		}
	} else {
		for {
			shortCode = generateShortCode(8)
			isNewCodeExists, err := db.isLinkExists(shortCode)
			if err != nil {
				return "", err
			}
			if !isNewCodeExists {
				break
			}
		}
	}

	_, err := db.Pool.Exec(db.Ctx, "INSERT INTO short_links (short_code, original_url, created_at, expires_at) VALUES	($1, $2, $3, $4)", shortCode, originalUrl, createdAt, expiresAt)
	if err != nil {
		return "", fmt.Errorf("Error while inserting values: %v", err.Error())
	}

	return shortCode, nil
}

func (db *Database) Delete(shortCode string) error {
	isCodeExists, err := db.isLinkExists(shortCode)
	if err != nil {
		return err
	}
	if !isCodeExists {
		return fmt.Errorf("Code \"%v\" doesn't exists", shortCode)
	}

	_, err = db.Pool.Exec(db.Ctx, "DELETE FROM short_links WHERE short_code = $1", shortCode)
	if err != nil {
		return fmt.Errorf("Error while deleting code \"%v\" from database: %v", shortCode, err)
	}

	return nil
}

func (db *Database) Stop() {
	db.Ctx.Done()
	db.Pool.Close()
}
