package service_test

import (
	"testing"
	"time"

	"github.com/rezect/url-shortener/internal/service"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	ls *service.LinkService
}

func (suite *ServiceTestSuite) SetupSuite() {
	mockDb := &testhelpers.MockRepository{}
	suite.ls = service.NewLinkService(mockDb)
}

func (suite *ServiceTestSuite) TestCreateLink_OK() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "shortener"
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.Equal(customAlias, alias)
	suite.NotEqual(time.Time{}, createdAt)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestCreateLink_UrlWithoutHttps() {
	originalURL := "github.com/rezect/url-shortener"
	customAlias := "shortener"
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrInvalidURL, err)
}

func (suite *ServiceTestSuite) TestCreateLink_LinkWithoutDomain() {
	originalURL := "https://github/rezect/url-shortener"
	customAlias := "shortener"
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrInvalidURL, err)
}

func (suite *ServiceTestSuite) TestCreateLink_InvalidAlias() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "too_long_alias_000000000000000000000000"
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrInvalidAlias, err)
}

func (suite *ServiceTestSuite) TestCreateLink_AliasExists() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "exists"
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrAliasExists, err)
}

func (suite *ServiceTestSuite) TestCreateLink_WithoutAlias() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := ""
	alias, createdAt, err := suite.ls.CreateLink(originalURL, customAlias)

	suite.NotEqual(customAlias, alias)
	suite.NotEqual(time.Time{}, createdAt)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestDeleteLink_OK() {
	customAlias := "exists"
	err := suite.ls.DeleteLink(customAlias)

	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestDeleteLink_LinkDoesNotExist() {
	customAlias := "does not exist"
	err := suite.ls.DeleteLink(customAlias)

	suite.Equal(service.ErrNotFound, err)
}

func (suite *ServiceTestSuite) TestRedirect_OK() {
	customAlias := "exists"
	originUrl, err := suite.ls.Redirect(customAlias)

	suite.NoError(err)
	suite.Equal("http://github.com/rezect", originUrl)
}

func (suite *ServiceTestSuite) TestRedirect_LinkDoesNotExist() {
	customAlias := "does not exist"
	originUrl, err := suite.ls.Redirect(customAlias)

	suite.Equal(service.ErrNotFound, err)
	suite.Equal("", originUrl)
}

func TestMain(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
