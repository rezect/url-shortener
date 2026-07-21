package handler_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/rezect/url-shortener/internal/handler"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/suite"
)

type GetTestSuite struct {
	suite.Suite
	mux http.Handler
}

func (suite *GetTestSuite) SetupSuite() {
	h := handler.NewHandler(&testhelpers.MockLinkService{}, BaseUrl)
	suite.mux = h.GetMux()
}

func (suite *GetTestSuite) TestRedirect_OK() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/s/exists", nil)
	suite.mux.ServeHTTP(w, r)

	suite.Equal(http.StatusFound, w.Code)
	suite.Equal("original url", w.Header().Get("Location"))
}

func (suite *GetTestSuite) TestRedirect_AliasNotExists() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/s/random_alias", nil)
	suite.mux.ServeHTTP(w, r)

	suite.Equal(http.StatusNotFound, w.Code)
}
