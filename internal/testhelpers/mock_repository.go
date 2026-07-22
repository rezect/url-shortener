package testhelpers

import (
	"context"
	"time"

	"github.com/rezect/url-shortener/internal/models"
)

type MockLinkRepo struct{}

func (r *MockLinkRepo) Create(ctx context.Context, originalUrl string, shortCode string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error) {
	return time.Now(), nil
}

func (r *MockLinkRepo) Get(ctx context.Context, alias string) (*models.ShortLink, error) {
	createdAt := time.Now()
	return &models.ShortLink{
		Id:          1,
		ShortCode:   "rezect",
		OriginalUrl: "http://github.com/rezect",
		CreatedAt:   &createdAt,
		ExpiresAt:   nil,
	}, nil
}

func (r *MockLinkRepo) Exists(ctx context.Context, shortCode string) (bool, error) {
	return shortCode == "exists", nil
}

func (r *MockLinkRepo) Delete(ctx context.Context, shortCode string) error {
	return nil
}

type MockClickRepo struct{}

func (r *MockClickRepo) Create(ctx context.Context, shortCode string, ip string, userAgent, referrer *string) error {
	return nil
}

func (r *MockClickRepo) GetTotalClicks(ctx context.Context, shortCode string) (int64, error) {
	return 0, nil
}

func (r *MockClickRepo) GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error) {
	report := map[time.Time]int{
		time.Now().AddDate(0, -1, 0): 100,
		time.Now().AddDate(0, -2, 0): 200,
		time.Now().AddDate(0, -3, 0): 300,
		time.Now().AddDate(0, -4, 0): 400,
	}

	return &report, nil
}
