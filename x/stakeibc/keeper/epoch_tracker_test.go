package keeper_test

import (
	"math"
	"strconv"

	sdkmath "cosmossdk.io/math"

	epochtypes "github.com/Stride-Labs/stride/v28/x/epochs/types"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

// These are used to indicate that the value does not matter for the sake of the test
const (
	DefaultEpochDurationSeconds = 10.0
	DefaultNextStartTimeSeconds = 10.0
	ToNanoSeconds               = 1_000_000_000
)

func (s *KeeperTestSuite) createNEpochTracker(n int) []types.EpochTracker {
	items := make([]types.EpochTracker, n)
	for i := range items {
		items[i].EpochIdentifier = strconv.Itoa(i)
		s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, items[i])
	}
	return items
}

func (s *KeeperTestSuite) TestEpochTrackerGet() {
	items := s.createNEpochTracker(10)
	for _, item := range items {
		rst, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx,
			item.EpochIdentifier,
		)
		s.Require().True(found)
		s.Require().Equal(
			&item,
			&rst,
		)
	}
}

func (s *KeeperTestSuite) TestEpochTrackerRemove() {
	items := s.createNEpochTracker(10)
	for _, item := range items {
		s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx,
			item.EpochIdentifier,
		)
		_, found := s.App.StakeibcKeeper.GetEpochTracker(s.Ctx,
			item.EpochIdentifier,
		)
		s.Require().False(found)
	}
}

func (s *KeeperTestSuite) TestEpochTrackerGetAll() {
	items := s.createNEpochTracker(10)
	s.Require().ElementsMatch(
		items,
		s.App.StakeibcKeeper.GetAllEpochTracker(s.Ctx),
	)
}

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
func (s *KeeperTestSuite) checkEpochElapsedShare(epochDurationSeconds float64, nextStartTimeSeconds float64, expectedShare sdkmath.LegacyDec) {
	s.SetupEpochElapsedShares(epochDurationSeconds, nextStartTimeSeconds)

	actualShare, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expectedShare, actualShare, "epoch elapsed share")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_StartOfEpoch() {
	// 10 second long epoch, with 10 seconds remaining => 0% share
	s.checkEpochElapsedShare(10.0, 10.0, sdkmath.LegacyNewDec(0))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_OneQuarterThroughEpoch() {
	// 10 second long epoch, with 7.5 seconds remaining => 2.5 seconds elapsed => 25% share
	s.checkEpochElapsedShare(10.0, 7.5, sdkmath.LegacyNewDec(25).Quo(sdkmath.LegacyNewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_MiddleOfEpoch() {
	// 10 second long epoch, with 5 seconds remaining => 50% share
	s.checkEpochElapsedShare(10.0, 5.0, sdkmath.LegacyNewDec(50).Quo(sdkmath.LegacyNewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_ThreeQuartersThroughEpoch() {
	// 10 second long epoch, with 2.5 seconds remaining => 7.5 seconds elapsed => 75% share
	s.checkEpochElapsedShare(10.0, 2.5, sdkmath.LegacyNewDec(75).Quo(sdkmath.LegacyNewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_AlmostAtEndOfEpoch() {
	// 10 second long epoch, with 0.1 seconds remaining => 99% share
	s.checkEpochElapsedShare(10.0, 0.1, sdkmath.LegacyNewDec(99).Quo(sdkmath.LegacyNewDec(100)))
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Successful_EndOfEpoch() {
	// 10 second long epoch, with 0 seconds remaining => 100% share
	s.checkEpochElapsedShare(10.0, 0.0, sdkmath.LegacyNewDec(1))
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
	s.Require().ErrorContains(err, "unable to convert epoch duration to int64")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_NextStartTimeOverflow() {
	// Set the next start time to the max uint in the epoch tracker so that it overflows when casting to an int
	maxNextStartTimeSeconds := float64(math.MaxUint64 / ToNanoSeconds)
	s.SetupEpochElapsedShares(DefaultEpochDurationSeconds, maxNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().ErrorContains(err, "unable to convert next epoch start time to int64")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_CurrentBlockTimeOverflow() {
	// Set the current block time to the max uint so that it overflows when casting to an int
	maxNextStartTimeSeconds := float64(math.MaxUint64 / ToNanoSeconds)
	s.SetupEpochElapsedShares(DefaultEpochDurationSeconds, maxNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().ErrorContains(err, "unable to convert next epoch start time to int64")
}

func (s *KeeperTestSuite) TestEpochElapsedShare_Failed_BlockTimeOutsideEpoch() {
	// Setting the duration to 0 will make the epoch start and end time equal to each other
	// Which will violate the safety constraint
	invalidDuration := 0.0
	s.SetupEpochElapsedShares(invalidDuration, DefaultNextStartTimeSeconds)

	_, err := s.App.StakeibcKeeper.GetStrideEpochElapsedShare(s.Ctx)
	s.Require().ErrorContains(err, "is not within current epoch")
}
