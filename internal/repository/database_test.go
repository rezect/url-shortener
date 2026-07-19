package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	mainDB      *Database
	conn        *Database
	tx          pgx.Tx
	ctx         context.Context
}

var (
	originalURL = "https://github.com/rezect/url-shortener"
	customAlias = "shortener"
)

func (suite *RepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.pgContainer = testhelpers.CreatePostgresContainer(suite.T(), suite.ctx)

	repository, err := NewDatabase(suite.pgContainer.ConnString, suite.ctx)
	require.NoError(suite.T(), err)
	suite.mainDB = repository
	suite.conn = nil
}

func (suite *RepoTestSuite) SetupTest() {
	tx, err := suite.mainDB.BeginTransaction(context.Background())
	if err != nil {
		suite.T().Fatal(err)
	}
	dbTx := suite.mainDB.WithTx(tx)

	suite.conn = dbTx
	suite.tx = tx
}

func (suite *RepoTestSuite) TearDownTest() {
	suite.tx.Rollback(context.Background())

	suite.conn = nil
	suite.tx = nil
}

func (suite *RepoTestSuite) TestCreateLink() {
	createdAt, err := suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)

	suite.NoError(err)
	suite.NotEqual(time.Time{}, createdAt)
}

func (suite *RepoTestSuite) TestExists() {
	isExists, err := suite.conn.Exists(context.Background(), customAlias)
	suite.NoError(err)
	suite.False(isExists)

	_, err = suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)
	suite.NoError(err)

	isExists, err = suite.conn.Exists(context.Background(), customAlias)
	suite.NoError(err)
	suite.True(isExists)
}

func (suite *RepoTestSuite) TestDeleteLink() {
	_, err := suite.conn.Create(context.Background(), originalURL, customAlias, nil, nil)
	suite.NoError(err)

	err = suite.conn.Delete(context.Background(), customAlias)
	suite.NoError(err)

	err = suite.conn.Delete(context.Background(), customAlias)
	suite.NoError(err)
}

func TestMain(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}
