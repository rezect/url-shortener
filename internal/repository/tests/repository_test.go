package repository_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestMain(t *testing.T) {
	suite.Run(t, new(LinkRepoTestSuite))
	suite.Run(t, new(ClickRepoTestSuite))
}
