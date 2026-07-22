package repository_test

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezect/url-shortener/internal/repository"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	originalURL = "https://github.com/rezect/url-shortener"
	customAlias = "shortener"
)

type LinkRepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer

	mainDB      *repository.LinkRepository
	conn        *repository.LinkRepository
	
	tx          pgx.Tx
	ctx         context.Context
}

func (suite *LinkRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.pgContainer = testhelpers.CreatePostgresContainer(suite.T(), suite.ctx)

	pool, err := pgxpool.New(suite.ctx, suite.pgContainer.ConnString)
	if err != nil {
		suite.T().Fatal(err)
	}

	repository := repository.NewLinkRepository(suite.ctx, suite.pgContainer.ConnString, pool)
	require.NoError(suite.T(), err)
	suite.mainDB = repository
	suite.conn = nil
}

func (suite *LinkRepoTestSuite) SetupTest() {
	tx, err := suite.mainDB.BeginTransaction(context.Background())
	if err != nil {
		suite.T().Fatal(err)
	}
	dbTx := suite.mainDB.WithTx(tx)

	suite.conn = dbTx
	suite.tx = tx
}

func (suite *LinkRepoTestSuite) TearDownTest() {
	suite.tx.Rollback(context.Background())

	suite.conn = nil
	suite.tx = nil
}

func (suite *LinkRepoTestSuite) TestLink_CreateLink() {
	createdAt, err := suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)

	suite.NoError(err)
	suite.NotEqual(time.Time{}, createdAt)
}

func (suite *LinkRepoTestSuite) TestLink_Exists() {
	isExists, err := suite.conn.Exists(context.Background(), customAlias)
	suite.NoError(err)
	suite.False(isExists)

	_, err = suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)
	suite.NoError(err)

	isExists, err = suite.conn.Exists(context.Background(), customAlias)
	suite.NoError(err)
	suite.True(isExists)
}

func (suite *LinkRepoTestSuite) TestLink_DeleteLink() {
	_, err := suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)
	suite.NoError(err)

	err = suite.conn.Delete(context.Background(), customAlias)
	suite.NoError(err)

	err = suite.conn.Delete(context.Background(), customAlias)
	suite.NoError(err)
}

func (suite *LinkRepoTestSuite) TestLink_Get() {
	_, err := suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)
	suite.NoError(err)

	link, err := suite.conn.Get(context.Background(), customAlias)
	suite.NoError(err)
	suite.Equal(originalURL, link.OriginalUrl)
	suite.Equal(customAlias, link.ShortCode)
}
