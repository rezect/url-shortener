package handler

import (
	"context"
	"net/http"
	"time"

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

	CreateClick(ctx context.Context, shortCode string, ip string, userAgent, referer *string) error

	GetTotalClicks(ctx context.Context, shortCode string) (string, int64, time.Time, error)

	GetDailyClicks(ctx context.Context, shortCode string) (*map[time.Time]int, error)

	Stop()
}

func NewHandler(ls Service, queue Queue, baseUrl string) *Handler {
	return &Handler{
		Service: ls,
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
	mux.HandleFunc("POST /api/v1/clicks", h.HandlerPost_CreateClick)
	mux.HandleFunc("GET /s/{alias}", h.HandlerGet_Redirect)
	mux.HandleFunc("GET /api/v1/stats/{short_code}", h.HandlerGet_LinkStatistic)

	return mux
}
