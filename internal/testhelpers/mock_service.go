package testhelpers

import (
	"context"
	"time"

	"github.com/rezect/url-shortener/internal/service"
)

type MockService struct{}

func (ls *MockService) CreateLink(ctx context.Context, originUrl string, customAlias string) (string, time.Time, error) {
	switch customAlias {
	case "invalid url":
		return "", time.Time{}, service.ErrInvalidURL
	case "invalid alias":
		return "", time.Time{}, service.ErrInvalidAlias
	case "exists":
		return "", time.Time{}, service.ErrAliasExists
	case "":
		return "exists", time.Now(), nil
	default:
		return customAlias, time.Now(), nil
	}
}

func (ls *MockService) DeleteLink(ctx context.Context, targetAlias string) error {
	if targetAlias == "exists" {
		return nil
	} else {
		return service.ErrNotFound
	}
}

func (ls *MockService) Redirect(ctx context.Context, targetAlias string) (string, error) {
	if targetAlias == "exists" {
		return "original url", nil
	} else {
		return "", service.ErrNotFound
	}
}

func (ls *MockService) CreateClick(ctx context.Context, shortCode string, ip string, userAgent, referer *string) error {
	return nil
}

func (ls *MockService) GetTotalClicks(ctx context.Context, shortCode string) (string, int64, time.Time, error) {
	return "", 0, time.Time{}, nil
}

func (ls *MockService) GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error) {
	return nil, nil
}

func (ls *MockService) Stop() {}
