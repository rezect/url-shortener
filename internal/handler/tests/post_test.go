package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/rezect/url-shortener/internal/handler"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/suite"
)

type PostTestSuite struct {
	suite.Suite
	hs *handler.Handler
}

type responseCreatedLink struct {
	ShortUrl    string    `json:"short_url"`
	OriginalUrl string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func (suite *PostTestSuite) SetupSuite() {
	ls := &testhelpers.MockLinkService{}
	suite.hs = handler.NewHandler(ls, BaseUrl)
}

func (suite *PostTestSuite) TestCreateLink_OK() {
	w := httptest.NewRecorder()
	body := strings.NewReader(fmt.Sprintf(`{"url":"%v","custom_alias":"%v"}`, OriginalUrl, CustomAlias))
	r := httptest.NewRequest("POST", BaseUrl, body)

	suite.hs.HandlerPost_CreateLink(w, r)

	suite.Equal(http.StatusCreated, w.Code)

	var linkData responseCreatedLink
	err := json.NewDecoder(w.Body).Decode(&linkData)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.Equal(OriginalUrl, linkData.OriginalUrl)
	suite.Equal(fmt.Sprintf(`%v/s/%v`, BaseUrl, CustomAlias), linkData.ShortUrl)
	suite.NotEqual(time.Time{}, linkData.CreatedAt)
}

func (suite *PostTestSuite) TestCreateLink_WrongBody() {
	w := httptest.NewRecorder()
	body := strings.NewReader(fmt.Sprintf(`{"wrong_key1":"%v","wrong_key2":"%v"}`, OriginalUrl, CustomAlias))
	r := httptest.NewRequest("POST", BaseUrl, body)

	suite.hs.HandlerPost_CreateLink(w, r)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *PostTestSuite) TestCreateLink_AliasExists() {
	w := httptest.NewRecorder()
	body := strings.NewReader(fmt.Sprintf(`{"url":"%v","custom_alias":"%v"}`, OriginalUrl, "exists"))
	r := httptest.NewRequest("POST", BaseUrl, body)

	suite.hs.HandlerPost_CreateLink(w, r)

	suite.Equal(http.StatusConflict, w.Code)
}

func (suite *PostTestSuite) TestCreateLink_InvalidAliasOrUrl() {
	w := httptest.NewRecorder()
	body := strings.NewReader(fmt.Sprintf(`{"url":"%v","custom_alias":"%v"}`, OriginalUrl, "invalid alias"))
	r := httptest.NewRequest("POST", BaseUrl, body)

	suite.hs.HandlerPost_CreateLink(w, r)

	suite.Equal(http.StatusBadRequest, w.Code)
}
