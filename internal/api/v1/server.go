package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/rezect/ttl-cache-server/pkg/cache"
	"github.com/rezect/url-shortener/internal/repository/postgres"
	"github.com/rezect/url-shortener/internal/response"
)

type Server struct {
	BaseURL string
	Srv     *http.Server
	Cache   *cache.Cache
	Db      *postgres.Database
	Logger  *log.Logger
}

func NewServer(addr string, dbConnString string, baseServerUrl string, ctx context.Context) (*Server, error) {
	db, err := postgres.NewDatabase(dbConnString, ctx)
	if err != nil {
		return nil, err
	}

	srv := Server{
		BaseURL: baseServerUrl,
		Cache:   cache.CacheNew(),
		Db:      db,
		Logger:  log.Default(),
	}

	srv.Srv = &http.Server{
		Addr:         addr,
		Handler:      srv.handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &srv, nil
}

func (s *Server) Start() error {
	return s.Srv.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err := s.Srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	s.Cache.Stop()
	s.Db.Stop()

	return nil
}

func (s *Server) handler() http.Handler {
	mux := s.setupRouts()
	// TODO: add middleware
	return mux
}

func (s *Server) setupRouts() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/shorten", s.handlePost_CreateLink)

	return mux
}

type postLink struct {
	Url         string `json:"url"`
	CustomAlias string `json:"custom_alias"`
}

func (s *Server) handlePost_CreateLink(w http.ResponseWriter, r *http.Request) {
	var linkData postLink
	err := json.NewDecoder(r.Body).Decode(&linkData)
	if err != nil {
		s.Logger.Printf("Error while decoding input body: %v", err)
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error while decoding input body: %v", err)})
		return
	}
	defer r.Body.Close()

	createdTime := time.Now()
	code, err := s.Db.Create(linkData.Url, createdTime, nil, linkData.CustomAlias)
	if err != nil {
		s.Logger.Printf("Error while creating new link: %v", err)
		response.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error while creating new link: %v", err)})
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{
		"short_url":    fmt.Sprintf("%v/s/%v", s.BaseURL, code),
		"original_url": linkData.Url,
		"created_at":   createdTime.Format(time.RFC3339),
	})
}
