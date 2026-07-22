package repository_test

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezect/url-shortener/internal/repository"
	"github.com/rezect/url-shortener/internal/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClickRepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer

	mainDB *repository.ClickRepository
	conn   *repository.ClickRepository

	tx  pgx.Tx
	ctx context.Context
}

func (suite *ClickRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.pgContainer = testhelpers.CreatePostgresContainer(suite.T(), suite.ctx)

	pool, err := pgxpool.New(suite.ctx, suite.pgContainer.ConnString)
	if err != nil {
		suite.T().Fatal(err)
	}

	repository := repository.NewClickRepository(suite.ctx, suite.pgContainer.ConnString, pool)
	require.NoError(suite.T(), err)
	suite.mainDB = repository
	suite.conn = nil
}

func (suite *ClickRepoTestSuite) SetupTest() {
	tx, err := suite.mainDB.BeginTransaction(context.Background())
	if err != nil {
		suite.T().Fatal(err)
	}
	dbTx := suite.mainDB.WithTx(tx)

	suite.conn = dbTx
	suite.tx = tx
}

func (suite *ClickRepoTestSuite) TearDownTest() {
	suite.tx.Rollback(context.Background())

	suite.conn = nil
	suite.tx = nil
}

func (s *ClickRepoTestSuite) TestClicks_Create() {
	err := s.conn.Create(s.ctx, "exists", "000.00.000.000", nil, nil)
	s.NoError(err)
}

func (s *ClickRepoTestSuite) TestClicks_GetTotalClicks() {
	ua := "Firefox"
	shortCode := "exists"
	var clicksToDo int64 = 101

	for range clicksToDo {
		err := s.conn.Create(s.ctx, shortCode, "000.00.000.000", &ua, nil)
		s.NoError(err)
	}
	
	totalClicks, err := s.conn.GetTotalClicks(s.ctx, shortCode)
	s.NoError(err)
	s.Equal(clicksToDo, totalClicks)
}