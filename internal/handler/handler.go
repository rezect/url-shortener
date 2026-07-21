package handler

import (
	"net/http"
	"time"
)

type Handler struct {
	LinkService Service
	BaseURL     string
}

type Service interface {
	CreateLink(originUrl string, customAlias string) (string, time.Time, error)

	DeleteLink(targetAlias string) error

	Redirect(alias string) (string, error)
}

func NewHandler(ls Service, baseUrl string) *Handler {
	return &Handler{
		LinkService: ls,
		BaseURL: baseUrl,
	}
}

func (h *Handler) Stop() {}

func (h *Handler) GetMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", h.HandlerPost_CreateLink)
	mux.HandleFunc("GET /s/{alias}", h.HandlerGet_Redirect)

	return mux
}
