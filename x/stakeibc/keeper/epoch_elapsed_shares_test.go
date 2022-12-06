package keeper_test

import (
	"math"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// These are used to indicate that the value does not matter for the sake of the test
const (
	DefaultEpochDurationSeconds = 10.0
	DefaultNextStartTimeSeconds = 10.0
	ToNanoSeconds               = 1_000_000_000
)

// TODO: Move keeper utility functions to new file
func (s *KeeperTestSuite) SetupEpochElapsedShares(epochDurationSeconds float64, nextStartTimeSeconds float64) {
	// We call this to instantiate the block time
	s.CreateTransferChannel(HostChainId)

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		Duration:           uint64(epochDurationSeconds * ToNanoSeconds),
		NextEpochStartTime: uint64(float64(s.Coordinator.CurrentTime.UnixNano()) + (nextStartTimeSeconds * ToNanoSeconds)),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)
}

// Helper function to create an epoch tracker and check that the elapsed share matches expectations
func (s *KeeperTestSuite) checkEpochElapsedShare(epochDurationSeconds float64, nextStartTimeSeconds float64, expectedShare sdk.Dec) {
	s.SetupEpochElapsedShares(epochDurationSeconds, nextStartTimeSeconds)

	actualShare, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expectedShare, actualShare, "epoch elapsed share")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_StartOfEpoch() {
	// 10 second long epoch, with 10 seconds remaining => 0% share
	s.checkEpochElapsedShare(10.0, 10.0, sdk.NewDec(0))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_OneQuarterThroughEpoch() {
	// 10 second long epoch, with 7.5 seconds remaining => 2.5 seconds elapsed => 25% share
	s.checkEpochElapsedShare(10.0, 7.5, sdk.NewDec(25).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_MiddleOfEpoch() {
	// 10 second long epoch, with 5 seconds remaining => 50% share
	s.checkEpochElapsedShare(10.0, 5.0, sdk.NewDec(50).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_ThreeQuartersThroughEpoch() {
	// 10 second long epoch, with 2.5 seconds remaining => 7.5 seconds elapsed => 75% share
	s.checkEpochElapsedShare(10.0, 2.5, sdk.NewDec(75).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_AlmostAtEndOfEpoch() {
	// 10 second long epoch, with 0.1 seconds remaining => 99% share
	s.checkEpochElapsedShare(10.0, 0.1, sdk.NewDec(99).Quo(sdk.NewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_EndOfEpoch() {
	// 10 second long epoch, with 0 seconds remaining => 100% share
	s.checkEpochElapsedShare(10.0, 0.0, sdk.NewDec(1))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_EpochNotFound() {
	// We skip the setup step her so an epoch tracker is never created
	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().EqualError(err, "Failed to get epoch tracker for stride_epoch: not found")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_DurationOverflow() {
	// Set the duration to the max uint in the epoch tracker so that it overflows when casting to an int
	maxDurationSeconds := float64(math.MaxUint64 / ToNanoSeconds)
	s.SetupEpochElapsedShares(maxDurationSeconds, DefaultNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)

	expectedErrMsg := `unable to convert epoch duration to int64, err: overflow: `
	expectedErrMsg += `unable to cast \d+ of type uint64 to int64: unable to cast to safe cast int`
	s.Require().Regexp(regexp.MustCompile(expectedErrMsg), err.Error())
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_NextStartTimeOverflow() {
	// Set the next start time to the max uint in the epoch tracker so that it overflows when casting to an int
	maxNextStartTimeSeconds := float64(math.MaxUint64 / ToNanoSeconds)
	s.SetupEpochElapsedShares(DefaultEpochDurationSeconds, maxNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	expectedErrMsg := `unable to convert next epoch start time to int64, err: overflow: `
	expectedErrMsg += `unable to cast \d+ of type uint64 to int64: unable to cast to safe cast int`
	s.Require().Regexp(regexp.MustCompile(expectedErrMsg), err.Error())
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_CurrentBlockTimeOverflow() {
	// Set the current block time to the max uint so that it overflows when casting to an int
	maxNextStartTimeSeconds := float64(math.MaxUint64 / ToNanoSeconds)
	s.SetupEpochElapsedShares(DefaultEpochDurationSeconds, maxNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	expectedErrMsg := `unable to convert next epoch start time to int64, err: overflow: `
	expectedErrMsg += `unable to cast \d+ of type uint64 to int64: unable to cast to safe cast int`
	s.Require().Regexp(regexp.MustCompile(expectedErrMsg), err.Error())
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_BlockTimeOutsideEpoch() {
	// Setting the duration to 0 will make the epoch start and end time equal to each other
	// Which will violate the safety constraint
	invalidDuration := 0.0
	s.SetupEpochElapsedShares(invalidDuration, DefaultNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	expectedErrMsg := "current block time 1577923350000000000 is not within current epoch (ending at 1577923360000000000): invalid epoch tracker"
	s.Require().EqualError(err, expectedErrMsg)
}
