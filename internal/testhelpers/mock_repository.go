package testhelpers

import (
	"context"
	"time"

	"github.com/rezect/url-shortener/internal/models"
)

type MockRepository struct{}

func (r *MockRepository) Create(ctx context.Context, originalUrl string, shortCode string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error) {
	return time.Now(), nil
}

func (r *MockRepository) Get(ctx context.Context, alias string) (*models.ShortLink, error) {
	createdAt := time.Now()
	return &models.ShortLink{
		Id:          1,
		ShortCode:   "rezect",
		OriginalUrl: "http://github.com/rezect",
		CreatedAt:   &createdAt,
		ExpiresAt:   nil,
	}, nil
}

func (r *MockRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	return shortCode == "exists", nil
}

func (r *MockRepository) Delete(ctx context.Context, shortCode string) error {
	return nil
}