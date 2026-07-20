package service

import (
	"context"
	"errors"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rezect/url-shortener/internal/repository"
)

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrInvalidAlias = errors.New("invalid alias")
	ErrAliasExists  = errors.New("alias already taken")
	ErrNotFound     = errors.New("link not found")
)

type LinkService struct {
	repo repository.Repository
}

func NewLinkService(db repository.Repository) *LinkService {
	return &LinkService{
		repo: db,
	}
}

func (ls *LinkService) CreateLink(originUrl string, customAlias string) (string, time.Time, error) {
	if err := validateURL(originUrl); err != nil {
		return "", time.Time{}, ErrInvalidURL
	}
	if customAlias != "" {
		if len(customAlias) > 20 || len(customAlias) < 6 || !isAliasValid(customAlias) {
			return "", time.Time{}, ErrInvalidAlias
		}

		isExists, err := ls.repo.Exists(context.Background(), customAlias)
		if err != nil {
			return "", time.Time{}, err
		}
		if isExists {
			return "", time.Time{}, ErrAliasExists
		}
	} else {
		for {
			customAlias = generateAlias()
			isExists, err := ls.repo.Exists(context.Background(), customAlias)
			if err != nil {
				return "", time.Time{}, err
			}
			if !isExists {
				break
			}
		}
	}

	createdAt, err := ls.repo.Create(context.Background(), originUrl, customAlias, nil, nil)
	if err != nil {
		return "", time.Time{}, nil
	}

	return customAlias, createdAt, nil
}

func (ls *LinkService) DeleteLink(targetAlias string) error {
	isExists, err := ls.repo.Exists(context.Background(), targetAlias)
	if err != nil {
		return err
	}
	if !isExists {
		return ErrNotFound
	}

	err = ls.repo.Delete(context.Background(), targetAlias)
	if err != nil {
		return err
	}

	return nil
}

func (ls *LinkService) Redirect(targetAlias string) (string, error) {
	isExists, err := ls.repo.Exists(context.Background(), targetAlias)
	if err != nil {
		return "", err
	}
	if !isExists {
		return "", ErrNotFound
	}

	link, err := ls.repo.Get(context.Background(), targetAlias)
	if err != nil {
		return "", err
	}

	return link.OriginalUrl, nil
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
