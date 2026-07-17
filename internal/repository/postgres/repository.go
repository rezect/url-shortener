package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func isLinkExists(ctx context.Context, dbpool *pgxpool.Pool, shortCode string) (bool, error) {
	var isAlreadyExists int
	err := dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM short_links WHERE short_code=$1", shortCode).Scan(&isAlreadyExists)
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

func Create(ctx context.Context, dbpool *pgxpool.Pool, originalUrl string, createdAt time.Time, expiresAt *time.Time, shortCode string) (string, error) {
	if shortCode != "" {
		isNewCodeExists, err := isLinkExists(ctx, dbpool, shortCode)
		if err != nil {
			return "", err
		}
		if isNewCodeExists {
			return "", fmt.Errorf("Link with code %v is already exists", shortCode)
		}
	} else {
		for {
			shortCode = generateShortCode(8)
			isNewCodeExists, err := isLinkExists(ctx, dbpool, shortCode)
			if err != nil {
				return "", err
			}
			if !isNewCodeExists {
				break
			}
		}
	}

	_, err := dbpool.Exec(ctx, "INSERT INTO short_links (short_code, original_url, created_at, expires_at) VALUES	($1, $2, $3, $4)", shortCode, originalUrl, createdAt, expiresAt)
	if err != nil {
		return "", fmt.Errorf("Error while inserting values: %v", err.Error())
	}

	return shortCode, nil
}

func Get(ctx context.Context, dbpool *pgxpool.Pool, shortCode string) (bool, error) {
	return isLinkExists(ctx, dbpool, shortCode)
}

func Delete(ctx context.Context, dbpool *pgxpool.Pool, shortCode string) error {
	isCodeExists, err := isLinkExists(ctx, dbpool, shortCode)
	if err != nil {
		return err
	}
	if !isCodeExists {
		return fmt.Errorf("Code \"%v\" doesn't exists", shortCode)
	}

	_, err = dbpool.Exec(ctx, "DELETE FROM short_links WHERE short_code = $1", shortCode)
	if err != nil {
		return fmt.Errorf("Error while deleting code \"%v\" from database: %v", shortCode, err)
	}

	return nil
}
