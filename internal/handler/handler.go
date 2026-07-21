package handler

import (
	"net/http"
	"time"
)

type HandlerService struct {
	LinkService Service
	BaseURL     string
}

type Service interface {
	CreateLink(originUrl string, customAlias string) (string, time.Time, error)

	DeleteLink(targetAlias string) error
}

func NewHandlerService(ls Service, baseUrl string) *HandlerService {
	return &HandlerService{
		LinkService: ls,
		BaseURL: baseUrl,
	}
}

func (h *HandlerService) Stop() {}

func (h *HandlerService) GetHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", h.HandlerPost_CreateLink)

	return mux
}
