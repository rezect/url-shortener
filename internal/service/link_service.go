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
}

type ClickRepository interface {
	Create(ctx context.Context, shortCode string, ip string, userAgent, referrer *string) error

	GetTotalClicks(ctx context.Context, shortCode string) (int64, error)

	GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error)
}

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrInvalidAlias = errors.New("invalid alias")
	ErrAliasExists  = errors.New("alias already taken")
	ErrNotFound     = errors.New("link not found")
)

type LinkService struct {
	linkRepo  LinkRepository
	clickRepo ClickRepository
}

func NewLinkService(linkRepo LinkRepository, clickRepo ClickRepository) *LinkService {
	return &LinkService{
		linkRepo:  linkRepo,
		clickRepo: clickRepo,
	}
}

func (ls *LinkService) CreateLink(ctx context.Context, originUrl string, customAlias string) (string, time.Time, error) {
	if err := validateURL(originUrl); err != nil {
		return "", time.Time{}, ErrInvalidURL
	}
	if customAlias != "" {
		if len(customAlias) > 20 || len(customAlias) < 6 || !isAliasValid(customAlias) {
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
		return "", time.Time{}, nil
	}

	return customAlias, createdAt, nil
}

func (ls *LinkService) DeleteLink(ctx context.Context, targetAlias string) error {
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

func (ls *LinkService) Redirect(ctx context.Context, targetAlias string) (string, error) {
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

func (h *LinkService) CreateClick(ctx context.Context, shortCode string, ip string, userAgent, referrer *string) error {
	if !isAliasValid(shortCode) {
		return ErrInvalidAlias
	}
	// TODO: проверка валидности ip

	// TODO: перенести обработку userAgent, referrer сюда из слоя репозитория
	err := h.clickRepo.Create(ctx, shortCode, ip, userAgent, referrer)
	if err != nil {
		return err
	}

	return nil
}

func (h *LinkService) GetTotalClicks(ctx context.Context, shortCode string) (int64, error) {
	if !isAliasValid(shortCode) {
		return 0, ErrInvalidAlias
	}
	totalClicks, err := h.clickRepo.GetTotalClicks(ctx, shortCode)
	if err != nil {
		return 0, err
	}

	return totalClicks, nil
}

func (h *LinkService) GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error) {
	if !isAliasValid(shortCode) {
		return nil, ErrInvalidAlias
	}
	totalClicks, err := h.clickRepo.GetDailyClicks(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return totalClicks, nil
}

func (h *LinkService) Stop() {}

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
