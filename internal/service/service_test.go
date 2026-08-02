package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/rezect/url-shortener/internal/service"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/suite"
)

var (
	originUrlTest = "http://github.com/rezect"
	customAliasTest = "rezect"
)

type ServiceTestSuite struct {
	suite.Suite
	ls *service.Service
}

func (suite *ServiceTestSuite) SetupSuite() {
	mockLinkRepo := &testhelpers.MockLinkRepo{}
	mockClickRepo := &testhelpers.MockClickRepo{}
	mockCache := &testhelpers.MockCache{}
	suite.ls = service.NewService(mockLinkRepo, mockClickRepo, mockCache)
}

func (suite *ServiceTestSuite) TestCreateLink_OK() {
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originUrlTest, customAliasTest)

	suite.Equal(customAliasTest, alias)
	suite.NotEqual(time.Time{}, createdAt)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestCreateLink_InvalidUrl() {
	originalURL := "not a url lol"
	customAlias := "shortener"
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrInvalidURL, err)
}

func (suite *ServiceTestSuite) TestCreateLink_InvalidAlias() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "too_long_alias_000000000000000000000000"
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrInvalidAlias, err)
}

func (suite *ServiceTestSuite) TestCreateLink_AliasExistsInCache() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "existsCache"
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrAliasExists, err)
}

func (suite *ServiceTestSuite) TestCreateLink_CustomAliasNotFound() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "notExistsCache"
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.NoError(err)
	suite.Equal(customAlias, alias)
	suite.NotEqual(time.Time{}, createdAt)
}

func (suite *ServiceTestSuite) TestCreateLink_AliasExistsOnlyInDatabase() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := "existsRepo"
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.Equal("", alias)
	suite.Equal(time.Time{}, createdAt)
	suite.Equal(service.ErrAliasExists, err)
}

func (suite *ServiceTestSuite) TestCreateLink_WithoutAlias() {
	originalURL := "https://github.com/rezect/url-shortener"
	customAlias := ""
	alias, createdAt, err := suite.ls.CreateLink(context.Background(), originalURL, customAlias)

	suite.NotEqual(customAlias, alias)
	suite.NotEqual(time.Time{}, createdAt)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestDeleteLink_ExistsCache() {
	customAlias := "existsCache"
	err := suite.ls.DeleteLink(context.Background(), customAlias)

	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestDeleteLink_ExistsDatabase() {
	customAlias := "existsRepo"
	err := suite.ls.DeleteLink(context.Background(), customAlias)

	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestDeleteLink_NotExistsCache() {
	customAlias := "notExistsCache"
	err := suite.ls.DeleteLink(context.Background(), customAlias)

	suite.Equal(service.ErrNotFound, err)
}

func (suite *ServiceTestSuite) TestDeleteLink_NotExists() {
	customAlias := "bibolibo"
	err := suite.ls.DeleteLink(context.Background(), customAlias)

	suite.Equal(service.ErrNotFound, err)
}

func (suite *ServiceTestSuite) TestRedirect_ExistsCache() {
	customAlias := "existsCache"
	originUrl, err := suite.ls.Redirect(context.Background(), customAlias)

	suite.NoError(err)
	suite.Equal("http://github.com/rezect", originUrl)
}

func (suite *ServiceTestSuite) TestRedirect_NotExistsCache() {
	customAlias := "notExistsCache"
	originUrl, err := suite.ls.Redirect(context.Background(), customAlias)

	suite.Equal(service.ErrNotFound, err)
	suite.Equal("", originUrl)
}

func (suite *ServiceTestSuite) TestRedirect_NotExists() {
	customAlias := "doesNotExists"
	originUrl, err := suite.ls.Redirect(context.Background(), customAlias)

	suite.Equal(service.ErrNotFound, err)
	suite.Equal("", originUrl)
}

func (suite *ServiceTestSuite) TestRedirect_ExistsInDatabase() {
	customAlias := "existsRepo"
	originUrl, err := suite.ls.Redirect(context.Background(), customAlias)

	suite.NoError(err)
	suite.Equal("http://github.com/rezect", originUrl)
}

func (suite *ServiceTestSuite) TestGetTotalClicks_ExistsCache() {
	originUrl, totalClicks, createdAt, err := suite.ls.GetTotalClicks(
		context.Background(),
		"existsCache",
	)

	suite.NoError(err)
	suite.Equal(originUrlTest, originUrl)
	suite.Equal((int64)(0), totalClicks)
	suite.NotEqual(time.Time{}, createdAt)
}

func (suite *ServiceTestSuite) TestGetTotalClicks_ExistsDatabase() {
	originUrl, totalClicks, createdAt, err := suite.ls.GetTotalClicks(
		context.Background(),
		"existsRepo",
	)
	
	suite.NoError(err)
	suite.Equal(originUrlTest, originUrl)
	suite.Equal((int64)(0), totalClicks)
	suite.NotEqual(time.Time{}, createdAt)
}

func (suite *ServiceTestSuite) TestGetDailyClicks_ExistsCache() {
	_, err := suite.ls.GetDailyClicks(
		context.Background(),
		"existsCache",
	)
	suite.NoError(err)
}

func (suite *ServiceTestSuite) TestGetDailyClicks_ExistsDatabase() {
	_, err := suite.ls.GetDailyClicks(
		context.Background(),
		"existsRepo",
	)
	suite.NoError(err)
}

func TestMain(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
