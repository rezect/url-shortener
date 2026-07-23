package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rezect/url-shortener/internal/models"
	"github.com/rezect/url-shortener/internal/response"
	"github.com/rezect/url-shortener/internal/service"
)

func (h *Handler) HandlerGet_Redirect(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Alias is empty",
		})
		return
	}

	userAgent := r.Header.Get("User-Agent")
	referer := r.Header.Get("Referer")

	originalURL, err := h.Service.Redirect(context.Background(), alias)
	if errors.Is(err, service.ErrNotFound) {
		response.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error": "Link is not found",
		})
		return
	} else if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error",
		})
		return
	}

	h.Queue.Push(models.Click{
		ShortCode: alias,
		Ip: "192.168.122.89",
		UserAgent: userAgent,
		Referer: referer,
	})
	w.Header().Set("Location", originalURL)
	response.WriteJSON(w, http.StatusFound, nil)
}

type ResponseStatistic struct {
	ShortCode    string            `json:"short_code"`
	OriginalUrl  string            `json:"original_url"`
	CreatedAt    time.Time         `json:"created_at"`
	TotalClicks  int64             `json:"total_clicks"`
	ClicksPerDay map[time.Time]int `json:"clicks_per_day"`
}

func (h *Handler) HandlerGet_LinkStatistic(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")

	originalUrl, totalLinkClicks, createdAt, err := h.Service.GetTotalClicks(context.Background(), shortCode)
	if errors.Is(err, service.ErrInvalidAlias) {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	} else if errors.Is(err, service.ErrNotFound) {
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

	responseData := ResponseStatistic{
		ShortCode: shortCode,
		OriginalUrl: originalUrl,
		CreatedAt: createdAt,
		TotalClicks: totalLinkClicks,
		ClicksPerDay: map[time.Time]int{},
	}

	response.WriteJSON(w, http.StatusOK, responseData)
}
