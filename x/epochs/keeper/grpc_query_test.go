package keeper_test

import (
	gocontext "context"

	"github.com/Stride-Labs/stride/v27/x/epochs/types"
)

func (s *KeeperTestSuite) TestQueryEpochInfos() {
	s.SetupTest()

	expectedEpochs := map[string]types.EpochInfo{}
	for _, epoch := range types.DefaultGenesis().Epochs {
		expectedEpochs[epoch.Identifier] = epoch
	}

	// Invalid param
	epochInfosResponse, err := s.queryClient.EpochInfos(gocontext.Background(), &types.QueryEpochsInfoRequest{})
	s.Require().NoError(err)
	s.Require().Len(epochInfosResponse.Epochs, 5)

	// check if EpochInfos are correct
	s.Require().Equal(epochInfosResponse.Epochs[0], expectedEpochs["day"])
	s.Require().Equal(epochInfosResponse.Epochs[1], expectedEpochs["hour"])
	s.Require().Equal(epochInfosResponse.Epochs[2], expectedEpochs["mint"])
	s.Require().Equal(epochInfosResponse.Epochs[3], expectedEpochs["stride_epoch"])
	s.Require().Equal(epochInfosResponse.Epochs[4], expectedEpochs["week"])
}
