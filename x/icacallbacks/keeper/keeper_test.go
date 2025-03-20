package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
