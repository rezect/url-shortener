package testhelpers

import (
	"context"
	"errors"
	"time"

	"github.com/rezect/url-shortener/internal/models"
	"github.com/rezect/url-shortener/internal/repository"
)

type MockLinkRepo struct{}

func (r *MockLinkRepo) Create(ctx context.Context, originalUrl string, alias string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error) {
	return time.Now(), nil
}

func (r *MockLinkRepo) Get(ctx context.Context, alias string) (*models.ShortLink, bool, error) {
	createdAt := time.Now()
	switch alias {
	case "existsRepo", "existsCache":
		return &models.ShortLink{
			Id:          1,
			ShortCode:   "rezect",
			OriginalUrl: "http://github.com/rezect",
			CreatedAt:   &createdAt,
			ExpiresAt:   nil,
		}, true, nil
	case "internalError":
		return &models.ShortLink{}, false, errors.New("internal error")
	default:
		return &models.ShortLink{}, false, nil
	}
}

func (r *MockLinkRepo) Exists(ctx context.Context, alias string) (bool, error) {
	return (alias == "existsRepo" || alias == "existsCache"), nil
}

func (r *MockLinkRepo) Delete(ctx context.Context, alias string) error {
	switch alias {
	case "existsRepo", "existsCache":
		return nil
	case "internalError":
		return errors.New("internal error")
	default:
		return repository.ErrNotFound
	}
}

type MockClickRepo struct{}

func (r *MockClickRepo) Create(ctx context.Context, alias string, ip string, userAgent, referrer *string) error {
	return nil
}

func (r *MockClickRepo) GetTotalClicks(ctx context.Context, alias string) (int64, error) {
	return 0, nil
}

func (r *MockClickRepo) GetDailyClicks(ctx context.Context, alias string) (*map[time.Time]int, error) {
	report := map[time.Time]int{
		time.Now().AddDate(0, -1, 0): 100,
		time.Now().AddDate(0, -2, 0): 200,
		time.Now().AddDate(0, -3, 0): 300,
		time.Now().AddDate(0, -4, 0): 400,
	}

	return &report, nil
}

func (r *MockLinkRepo) Stop() {}

func (r *MockClickRepo) Stop() {}
