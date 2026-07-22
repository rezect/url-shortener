package handler_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	BaseUrl    = "http://localhost:8000"
	OriginalUrl = "http://rezect.com/api/v1/shorten"
	CustomAlias = "rezect"
)

func TestMain(t *testing.T) {
	suite.Run(t, new(PostTestSuite))
	suite.Run(t, new(GetTestSuite))
}