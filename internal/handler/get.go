package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rezect/url-shortener/internal/response"
	"github.com/rezect/url-shortener/internal/service"
)

func (h *Handler) HandlerGet_Redirect(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if alias == "" {
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "alias is not set",
		})
		return
	}

	fmt.Printf(`Alias got: "%v"\n`, alias)

	originalURL, err := h.LinkService.Redirect(alias)
	if errors.Is(err, service.ErrNotFound) {
		response.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error": "Link is not found",
		})
		return
	} else if err != nil {
		response.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Link is not found",
		})
		return
	}

	w.Header().Set("Location", originalURL)
	response.WriteJSON(w, http.StatusFound, nil)
}
