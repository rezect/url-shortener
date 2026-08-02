package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/rezect/url-shortener/internal/middleware"
	"github.com/rezect/url-shortener/internal/models"
)

type Handler struct {
	Service Service
	Queue   Queue
	BaseURL string
}

type Queue interface {
	StartWorkers(n int)

	Push(click models.Click)

	Stop()
}

type Service interface {
	CreateLink(ctx context.Context, originUrl string, customAlias string) (string, time.Time, error)

	DeleteLink(ctx context.Context, targetAlias string) error

	Redirect(ctx context.Context, targetAlias string) (string, error)

	GetTotalClicks(ctx context.Context, shortCode string) (string, int64, time.Time, error)

	GetDailyClicks(ctx context.Context, shortCode string) (*map[string]int, error)

	Stop()
}

func NewHandler(s Service, queue Queue, baseUrl string) *Handler {
	return &Handler{
		Service: s,
		Queue:   queue,
		BaseURL: baseUrl,
	}
}

func (h *Handler) Stop() {
	h.Queue.Stop()
	h.Service.Stop()
}

func (h *Handler) GetMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", h.HandlerPost_CreateLink)
	mux.HandleFunc("GET /s/{alias}", h.HandlerGet_Redirect)
	mux.HandleFunc("GET /api/v1/stats/{short_code}", h.HandlerGet_LinkStatistic)
	mux.HandleFunc("GET /health", h.HandlerGet_Health)

	logger := middleware.Logger(mux)

	return logger
}
