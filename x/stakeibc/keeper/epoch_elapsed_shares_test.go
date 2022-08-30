package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

// TODO: Move keeper utility functions to new file
func (s *KeeperTestSuite) SetupEpochElapsedShares(epochDurationSeconds float64, nextStartTimeSeconds float64) {
	// We call this to instantiate the block time
	s.CreateTransferChannel(HostChainId)

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		Duration:           uint64(epochDurationSeconds * 1_000_000_000.0),
		NextEpochStartTime: uint64(float64(s.Coordinator.CurrentTime.UnixNano()) + (nextStartTimeSeconds * 1_000_000_000.0)),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), strideEpochTracker)
}

// Helper function to create an epoch tracker and check that the elapsed share matches expectations
func (s *KeeperTestSuite) checkEpochElapsedShare(epochDurationSeconds float64, nextStartTimeSeconds float64, expectedShare sdk.Dec) {
	s.SetupEpochElapsedShares(epochDurationSeconds, nextStartTimeSeconds)

	actualShare, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx())
	s.Require().NoError(err)
	s.Require().Equal(expectedShare, actualShare, "epoch elapsed share")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_StartOfEpoch() {
	// 10 second long epoch, with 10 seconds to go => 0% share
	s.checkEpochElapsedShare(10.0, 10.0, sdk.NewDec(0))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_OneQuarterThroughEpoch() {
	// 10 second long epoch, with 7.5 seconds to go => 2.5 seconds elapsed => 25% share
	s.checkEpochElapsedShare(10.0, 7.5, sdk.NewDec(25).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_MiddleOfEpoch() {
	// 10 second long epoch, with 5 seconds to go => 50% share
	s.checkEpochElapsedShare(10.0, 5.0, sdk.NewDec(50).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_ThreeQuartersThroughEpoch() {
	// 10 second long epoch, with 2.5 seconds to go => 7.5 seconds elapsed => 75% share
	s.checkEpochElapsedShare(10.0, 2.5, sdk.NewDec(75).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_AlmostAtEndOfEpoch() {
	// 10 second long epoch, with 0.1 seconds to go => 99% share
	s.checkEpochElapsedShare(10.0, 0.1, sdk.NewDec(99).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_EndOfEpoch() {
	// 10 second long epoch, with 0 seconds to go => 100% share
	s.checkEpochElapsedShare(10.0, 0.0, sdk.NewDec(1))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_EpochNotFound() {

}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_DurationOverflow() {

}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_NextStartTimeOverflow() {

}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_CurrentBlockTimeOverflow() {

}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_BlockTimeOutsideEpoch() {

}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_InvalidElapsedShare() {
	// Not sure if this is possible to invoke
}
