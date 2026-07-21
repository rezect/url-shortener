package testhelpers

import (
	"time"

	"github.com/rezect/url-shortener/internal/service"
)

type MockLinkService struct{}

func (ls *MockLinkService) CreateLink(originUrl string, customAlias string) (string, time.Time, error) {
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

func (ls *MockLinkService) DeleteLink(targetAlias string) error {
	if targetAlias == "exists" {
		return nil
	} else {
		return service.ErrNotFound
	}
}

func (ls *MockLinkService) Redirect(targetAlias string) (string, error) {
	if targetAlias == "exists" {
		return "original url", nil
	} else {
		return "", service.ErrNotFound
	}
}
