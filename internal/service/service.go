package service

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rezect/url-shortener/internal/cache"
	"github.com/rezect/url-shortener/internal/models"
)

type Cache interface {
	Get(key string) (string, error)

	Set(key string, value any, ttl time.Duration) error

	Delete(key string) error

	Clear()

	Stop()
}

type LinkRepository interface {
	Exists(ctx context.Context, alias string) (bool, error)

	Get(ctx context.Context, alias string) (*models.ShortLink, bool, error)

	Create(ctx context.Context, originalUrl string, alias string, createdAt *time.Time, expiresAt *time.Time) (time.Time, error)

	Delete(ctx context.Context, alias string) error

	Stop()
}

type ClickRepository interface {
	Create(ctx context.Context, alias string, ip string, userAgent, referrer *string) error

	GetTotalClicks(ctx context.Context, alias string) (int64, error)

	GetDailyClicks(ctx context.Context, alias string) (*map[time.Time]int, error)

	Stop()
}

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrInvalidAlias = errors.New("invalid alias")
	ErrAliasExists  = errors.New("alias already taken")
	ErrNotFound     = errors.New("link not found")

	TTL = 60 * time.Second
)

type Service struct {
	linkRepo  LinkRepository
	clickRepo ClickRepository
	cache     Cache
}

func NewService(linkRepo LinkRepository, clickRepo ClickRepository, cache Cache) *Service {
	return &Service{
		linkRepo:  linkRepo,
		clickRepo: clickRepo,
		cache:     cache,
	}
}

func (svc *Service) CreateLink(ctx context.Context, originUrl string, customAlias string) (string, time.Time, error) {
	if err := validateURL(originUrl); err != nil {
		return "", time.Time{}, ErrInvalidURL
	}
	if customAlias != "" {
		if !isAliasValid(customAlias) {
			return "", time.Time{}, ErrInvalidAlias
		}

		_, err := svc.cache.Get(customAlias)
		if errors.Is(err, cache.CacheHit) {
			return "", time.Time{}, ErrAliasExists
		} else if errors.Is(err, cache.CacheMiss) {
			isExists, err := svc.linkRepo.Exists(ctx, customAlias)
			if err != nil {
				return "", time.Time{}, err
			}
			if isExists {
				svc.cache.Set(customAlias, originUrl, TTL)
				return "", time.Time{}, ErrAliasExists
			}
		}
	} else {
		for {
			customAlias = generateAlias()

			_, err := svc.cache.Get(customAlias)
			if errors.Is(err, cache.CacheHit) {
				continue
			} else if errors.Is(err, cache.NotExists) {
				break
			}

			isExists, err := svc.linkRepo.Exists(ctx, customAlias)
			if err != nil {
				return "", time.Time{}, err
			}
			if !isExists {
				break
			} else {
				svc.cache.Set(customAlias, originUrl, TTL)
			}
		}
	}

	createdAt, err := svc.linkRepo.Create(ctx, originUrl, customAlias, nil, nil)
	if err != nil {
		return "", time.Time{}, err
	}
	svc.cache.Set(customAlias, originUrl, TTL)

	return customAlias, createdAt, nil
}

func (svc *Service) DeleteLink(ctx context.Context, targetAlias string) error {
	_, err := svc.cache.Get(targetAlias)
	if errors.Is(err, cache.CacheHit) {
		err = svc.linkRepo.Delete(ctx, targetAlias)
		if err != nil {
			return err
		}
		svc.cache.Delete(targetAlias)
		return nil
	} else if errors.Is(err, cache.CacheMiss) {
		isExists, err := svc.linkRepo.Exists(ctx, targetAlias)
		if err != nil {
			return err
		}
		if !isExists {
			return ErrNotFound
		} else {
			err = svc.linkRepo.Delete(ctx, targetAlias)
			if err != nil {
				return err
			}
			svc.cache.Delete(targetAlias)
			return nil
		}
	} else if errors.Is(err, cache.NotExists) {
		return ErrNotFound
	}

	return nil
}

func (svc *Service) Redirect(ctx context.Context, targetAlias string) (string, error) {
	originUrl, err := svc.cache.Get(targetAlias)
	if errors.Is(err, cache.CacheHit) {
		return originUrl, nil
	} else if errors.Is(err, cache.NotExists) {
		return "", ErrNotFound
	}

	link, isExists, err := svc.linkRepo.Get(ctx, targetAlias)
	if err != nil {
		return "", err
	}
	if !isExists {
		svc.cache.Set(targetAlias, nil, TTL)
		return "", ErrNotFound
	}
	if link == nil {
		return "", errors.New("trying to get nil value. link is nil!")
	}
	svc.cache.Set(targetAlias, link.OriginalUrl, TTL)

	return link.OriginalUrl, nil
}

func (svc *Service) CreateClick(ctx context.Context, alias string, ip string, userAgent, referrer *string) error {
	if !isAliasValid(alias) {
		return ErrInvalidAlias
	}
	// TODO: проверка валидности ip

	_, isExists, err := svc.linkRepo.Get(ctx, alias)
	if err != nil {
		return err
	} else if !isExists {
		return ErrNotFound
	}

	// TODO: перенести обработку userAgent, referrer сюда из слоя репозитория
	err = svc.clickRepo.Create(ctx, alias, ip, userAgent, referrer)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) GetTotalClicks(ctx context.Context, alias string) (string, int64, time.Time, error) {
	if !isAliasValid(alias) {
		return "", 0, time.Time{}, ErrInvalidAlias
	}

	linkData, isExists, err := svc.linkRepo.Get(ctx, alias)
	if err != nil {
		return "", 0, time.Time{}, err
	}
	if !isExists {
		return "", 0, time.Time{}, ErrNotFound
	}
	if linkData == nil {
		return "", 0, time.Time{}, errors.New("trying to get nil value. link is nil!")
	}
	if linkData.CreatedAt == nil {
		return "", 0, time.Time{}, errors.New("Link Data is invalid: linkData.CreatedAt is nil")
	}

	totalClicks, err := svc.clickRepo.GetTotalClicks(ctx, alias)
	if err != nil {
		return "", 0, time.Time{}, err
	}

	return linkData.OriginalUrl, totalClicks, *linkData.CreatedAt, nil
}

func (svc *Service) GetDailyClicks(ctx context.Context, alias string) (*map[time.Time]int, error) {
	if !isAliasValid(alias) {
		return nil, ErrInvalidAlias
	}
	isExists, err := svc.linkRepo.Exists(ctx, alias)
	if err != nil {
		return nil, err
	}
	if !isExists {
		return nil, ErrNotFound
	}

	totalClicks, err := svc.clickRepo.GetDailyClicks(ctx, alias)
	if err != nil {
		return nil, err
	}

	return totalClicks, nil
}

func (svc *Service) Stop() {
	svc.linkRepo.Stop()
	svc.clickRepo.Stop()
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
