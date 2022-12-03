package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	suite.Suite
	queryClient types.QueryClient
}

// Test helpers
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
