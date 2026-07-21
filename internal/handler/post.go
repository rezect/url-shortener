package handler

import (
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

func (h *HandlerService) HandlerPost_CreateLink(w http.ResponseWriter, r *http.Request) {
	var linkData postLinkData
	err := json.NewDecoder(r.Body).Decode(&linkData)
	if err != nil || linkData.Url == "" || linkData.CustomAlias == "" {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error while decoding input body: %v", err)})
		return
	}
	defer r.Body.Close()

	alias, createdAt, err := h.LinkService.CreateLink(linkData.Url, linkData.CustomAlias)
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
