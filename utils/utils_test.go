package utils_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
)

type UtilsTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UtilsTestSuite) SetupTest() {
	s.Setup()
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
