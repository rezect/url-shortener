package handler

import (
	"context"
	"errors"
	"net"
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
		Ip:        getIP(r),
		UserAgent: userAgent,
		Referer:   referer,
	})
	http.Redirect(w, r, originalURL, http.StatusFound)
}

type ResponseStatistic struct {
	ShortCode    string            `json:"short_code"`
	OriginalUrl  string            `json:"original_url"`
	CreatedAt    time.Time         `json:"created_at"`
	TotalClicks  int64             `json:"total_clicks"`
	ClicksPerDay map[string]int `json:"clicks_per_day"`
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

	clicksPerDay, err := h.Service.GetDailyClicks(context.Background(), shortCode)
	if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
		return
	}

	responseData := ResponseStatistic{
		ShortCode:    shortCode,
		OriginalUrl:  originalUrl,
		CreatedAt:    createdAt,
		TotalClicks:  totalLinkClicks,
		ClicksPerDay: *clicksPerDay,
	}

	response.WriteJSON(w, http.StatusOK, responseData)
}

func (h *Handler) HandlerGet_Health(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, nil)
}

func getIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
