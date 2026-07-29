package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rezect/url-shortener/internal/response"
	"github.com/rezect/url-shortener/internal/service"
)

type postLinkData struct {
	Url         string `json:"url"`
	CustomAlias string `json:"custom_alias"`
}

func (h *Handler) HandlerPost_CreateLink(w http.ResponseWriter, r *http.Request) {
	var linkData postLinkData
	err := json.NewDecoder(r.Body).Decode(&linkData)
	if err != nil || linkData.Url == "" {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error while decoding input body: %v", err)})
		return
	}
	defer r.Body.Close()

	alias, createdAt, err := h.Service.CreateLink(context.Background(), linkData.Url, linkData.CustomAlias)
	if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrInvalidAlias) {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	} else if errors.Is(err, service.ErrAliasExists) {
		response.WriteJSON(w, http.StatusConflict, map[string]string{
			"error": err.Error(),
		})
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{
		"short_url":    fmt.Sprintf("%v/s/%v", h.BaseURL, alias),
		"original_url": linkData.Url,
		"created_at":   createdAt.Format(time.RFC3339),
	})
}

type postClickData struct {
	ShortCode string `json:"short_code"`
}

func (h *Handler) HandlerPost_CreateClick(w http.ResponseWriter, r *http.Request) {
	var clickData postClickData
	err := json.NewDecoder(r.Body).Decode(&clickData)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	defer r.Body.Close()

	fmt.Print(r.RemoteAddr)
	
	err = h.Service.CreateClick(context.Background(), clickData.ShortCode, "192.168.0.0", nil, nil)
	if errors.Is(err, service.ErrNotFound) {
		response.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
		return
	} else if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{
		"short_url":    fmt.Sprintf("%v/s/%v", h.BaseURL, clickData.ShortCode),
	})
}
