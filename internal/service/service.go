package service

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rezect/url-shortener/internal/models"
)

type LinkRepository interface {
	Exists(ctx context.Context, shortCode string) (bool, error)

	Get(ctx context.Context, alias string) (*models.ShortLink, error)

	Create(ctx context.Context, originalUrl string, shortCode string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error)

	Delete(ctx context.Context, shortCode string) error

	Stop()
}

type ClickRepository interface {
	Create(ctx context.Context, shortCode string, ip string, userAgent, referrer *string) error

	GetTotalClicks(ctx context.Context, shortCode string) (int64, error)

	GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error)

	Stop()
}

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrInvalidAlias = errors.New("invalid alias")
	ErrAliasExists  = errors.New("alias already taken")
	ErrNotFound     = errors.New("link not found")
)

type Service struct {
	linkRepo  LinkRepository
	clickRepo ClickRepository
}

func NewService(linkRepo LinkRepository, clickRepo ClickRepository) *Service {
	return &Service{
		linkRepo:  linkRepo,
		clickRepo: clickRepo,
	}
}

func (ls *Service) CreateLink(ctx context.Context, originUrl string, customAlias string) (string, time.Time, error) {
	if err := validateURL(originUrl); err != nil {
		return "", time.Time{}, ErrInvalidURL
	}
	if customAlias != "" {
		if !isAliasValid(customAlias) {
			return "", time.Time{}, ErrInvalidAlias
		}

		isExists, err := ls.linkRepo.Exists(ctx, customAlias)
		if err != nil {
			return "", time.Time{}, err
		}
		if isExists {
			return "", time.Time{}, ErrAliasExists
		}
	} else {
		for {
			customAlias = generateAlias()
			isExists, err := ls.linkRepo.Exists(ctx, customAlias)
			if err != nil {
				return "", time.Time{}, err
			}
			if !isExists {
				break
			}
		}
	}

	createdAt, err := ls.linkRepo.Create(ctx, originUrl, customAlias, nil, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	return customAlias, createdAt, nil
}

func (ls *Service) DeleteLink(ctx context.Context, targetAlias string) error {
	isExists, err := ls.linkRepo.Exists(ctx, targetAlias)
	if err != nil {
		return err
	}
	if !isExists {
		return ErrNotFound
	}

	err = ls.linkRepo.Delete(ctx, targetAlias)
	if err != nil {
		return err
	}

	return nil
}

func (ls *Service) Redirect(ctx context.Context, targetAlias string) (string, error) {
	isExists, err := ls.linkRepo.Exists(ctx, targetAlias)
	if err != nil {
		return "", err
	}
	if !isExists {
		return "", ErrNotFound
	}

	link, err := ls.linkRepo.Get(ctx, targetAlias)
	if err != nil {
		return "", err
	}

	return link.OriginalUrl, nil
}

func (h *Service) CreateClick(ctx context.Context, shortCode string, ip string, userAgent, referrer *string) error {
	if !isAliasValid(shortCode) {
		return ErrInvalidAlias
	}
	// TODO: проверка валидности ip

	exists, err := h.linkRepo.Exists(ctx, shortCode)
	if err != nil {
		return err
	} else if !exists {
		return ErrNotFound
	}

	// TODO: перенести обработку userAgent, referrer сюда из слоя репозитория
	err = h.clickRepo.Create(ctx, shortCode, ip, userAgent, referrer)
	if err != nil {
		return err
	}

	return nil
}

func (ls *Service) GetTotalClicks(ctx context.Context, shortCode string) (string, int64, time.Time, error) {
	if !isAliasValid(shortCode) {
		return "", 0, time.Time{}, ErrInvalidAlias
	}

	isExists, err := ls.linkRepo.Exists(ctx, shortCode)
	if err != nil {
		return "", 0, time.Time{}, err
	}
	if !isExists {
		return "", 0, time.Time{}, ErrNotFound
	}

	linkData, err := ls.linkRepo.Get(ctx, shortCode)
	if err != nil {
		return "", 0, time.Time{}, err
	}
	if linkData.CreatedAt == nil {
		return "", 0, time.Time{}, errors.New("Link Data is invalid: linkData.CreatedAt is nil")
	}

	totalClicks, err := ls.clickRepo.GetTotalClicks(ctx, shortCode)
	if err != nil {
		return "", 0, time.Time{}, err
	}

	return linkData.OriginalUrl, totalClicks, *linkData.CreatedAt, nil
}

func (ls *Service) GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error) {
	if !isAliasValid(shortCode) {
		return nil, ErrInvalidAlias
	}
	isExists, err := ls.linkRepo.Exists(ctx, shortCode)
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, ErrNotFound
	}

	totalClicks, err := ls.clickRepo.GetDailyClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return totalClicks, nil
}

func (s *Service) Stop() {
	s.linkRepo.Stop()
	s.clickRepo.Stop()
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return errors.New("URL cannot be empty")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid URL format")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only HTTP and HTTPS schemes are allowed")
	}
	if u.Host == "" {
		return errors.New("missing hostname")
	}
	if len(rawURL) > 2048 {
		return errors.New("URL too long")
	}
	hostname := u.Hostname()
	if !strings.Contains(hostname, ".") {
		return errors.New("hostname must contain a domain (e.g., example.com)")
	}
	return nil
}

func isAliasValid(code string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(code)
}

func generateAlias() string {
	length := 8
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range length {
		randIdx := rand.Intn(len(alphabet))
		randChar := alphabet[randIdx]
		b[i] = randChar
	}

	return string(b)
}
