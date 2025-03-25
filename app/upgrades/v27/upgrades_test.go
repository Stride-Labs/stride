package v27_test

import (
	"fmt"
	"sort"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v26/app/upgrades/v27"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	upgradeHeight := int64(4)

	s.ConfirmUpgradeSucceededs(v27.UpgradeName, upgradeHeight)

	// Confirm consumer ID is set to 1
	params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(params.ConsumerId, "1")
}

// SortDelegations sorts delegations by delegator address
func sortDelegations(delegations []stakingtypes.Delegation) {
	sort.SliceStable(delegations, func(i, j int) bool {
		return delegations[i].DelegatorAddress < delegations[j].DelegatorAddress
	})
}

func (s *UpgradeTestSuite) TestDistributionFix() {
	// Set specific block height for deterministic testing
	s.Ctx = s.Ctx.WithBlockHeight(16925943) // 2025-03-24T13:30:39.449960913Z

	// Define validator address and missing stake amounts
	valAddr, _ := types.ValAddressFromBech32("stridevaloper1tlz6ksce084ndhwlq2usghamvh0dut9q4z2gxd")
	bondedPoolMissingStake := types.NewInt64Coin("stake", 1038549945)
	notBondedPoolMissingStake := types.NewInt64Coin("stake", 220000)

	// Define periods for slashing event to be inserted
	upperBoundPeriod := uint64(3913)
	slashingEventPeriod := uint64(3902)
	slashingEventBlock := uint64(4673775)
	slashingEventFraction := sdkmath.LegacyMustNewDecFromStr("0.0001") // 0.01% slash

	// Load faulty distribution state from mainnet export file
	jsonDistGenesis := os.MustReadFile("test_dist_genesis.json")
	var distGenesisState disttypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonDistGenesis, &distGenesisState)

	// Load matching staking state from mainnet export file
	jsonStakingGenesis := os.MustReadFile("test_staking_genesis.json")
	var stakingGenesisState stakingtypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonStakingGenesis, &stakingGenesisState)
	sortDelegations(stakingGenesisState.Delegations)

	// Fund the distribution module with outstanding rewards
	// This aligns bank state with distribution
	for i := range distGenesisState.OutstandingRewards {
		coins, _ := distGenesisState.OutstandingRewards[i].OutstandingRewards.TruncateDecimal()
		for _, coin := range coins {
			s.FundModuleAccount(disttypes.ModuleName, coin)
		}
	}

	// Fund the staking pools with missing stake
	// This aligns bank state with staking
	s.FundModuleAccount(stakingtypes.BondedPoolName, bondedPoolMissingStake)
	s.FundModuleAccount(stakingtypes.NotBondedPoolName, notBondedPoolMissingStake)

	// Initialize modules with imported states
	s.App.StakingKeeper.InitGenesis(s.Ctx, &stakingGenesisState)
	s.App.DistrKeeper.InitGenesis(s.Ctx, distGenesisState)

	// Confirm that withdrawing rewards fails for delegations created before height 4300034
	for _, delegation := range stakingGenesisState.Delegations {
		delAddr := types.MustAccAddressFromBech32(delegation.DelegatorAddress)

		period, err := s.App.DistrKeeper.GetDelegatorStartingInfo(s.Ctx, valAddr, delAddr)
		s.Require().NoError(err)
		s.Require().Positive(period.PreviousPeriod)
		s.Require().Positive(period.Height)

		if period.Height < 4300034 {
			// Older delegations should panic when attempting to withdraw rewards
			// due to the missing slashing event
			s.Require().Panics(func() {
				_, _ = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, delAddr, valAddr)
				fmt.Printf("%s should panic (%d < %d)\n", delAddr.String(), period.Height, 5047518)
				s.Require().True(false)
			})
		} else {
			// Newer delegations should work fine
			// Use a cached context to prevent state changes
			subCtx, _ := s.Ctx.CacheContext()

			_, err := s.App.DistrKeeper.WithdrawDelegationRewards(subCtx, delAddr, valAddr)
			s.Require().NoError(err)
		}

	}

	// Apply Fix
	// Insert the missing slashing event between blocks 4300034-5047517 (periods 3893-3912)
	// The slashing event represents a 0.01% slash that occurred but wasn't properly recorded
	err := s.App.DistrKeeper.SetValidatorSlashEvent(
		s.Ctx,
		valAddr,
		slashingEventBlock,
		slashingEventPeriod,
		disttypes.NewValidatorSlashEvent(slashingEventPeriod, slashingEventFraction),
	)
	s.Require().NoError(err)

	// Copy historical rewards data from upper bound period to the slashing event period
	// Note: By using the same historical rewards from the upper bound period, we're effectively
	// not accounting for rewards that accrued on approximately half the blocks between 4300034-5047517.
	// The reward amounts are extremely small as of 2025-03-25:
	// - 0.000000000055967683 INJ   (≈ $0.0000000005910187 USD)
	// - 0.000000006094164748 EVMOS (≈ $0.0000000000284597 USD)
	// At these microscopic values, the simplification has virtually no impact on users.
	hisoricalRewards, err := s.App.DistrKeeper.GetValidatorHistoricalRewards(s.Ctx, valAddr, upperBoundPeriod)
	s.Require().NoError(err)
	err = s.App.DistrKeeper.SetValidatorHistoricalRewards(s.Ctx, valAddr, slashingEventPeriod, hisoricalRewards)
	s.Require().NoError(err)

	// After applying the fix, all delegations should be able to withdraw rewards without panics
	for _, delegation := range stakingGenesisState.Delegations {
		s.Require().NotPanics(func() {
			_, err = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, types.MustAccAddressFromBech32(delegation.DelegatorAddress), valAddr)
			s.Require().NoError(err)
		})
	}
}
