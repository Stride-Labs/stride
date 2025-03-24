package v27_test

import (
	"fmt"
	"sort"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
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

// SortDelegations sorts delegations by validator address and then delegator address
func sortDelegations(delegations []stakingtypes.Delegation) {
	sort.SliceStable(delegations, func(i, j int) bool {
		// First compare validator addresses
		if delegations[i].ValidatorAddress != delegations[j].ValidatorAddress {
			return delegations[i].ValidatorAddress < delegations[j].ValidatorAddress
		}
		// If validator addresses are the same, compare delegator addresses
		return delegations[i].DelegatorAddress < delegations[j].DelegatorAddress
	})
}

func (s *UpgradeTestSuite) TestDistributionFix() {
	s.Ctx = s.Ctx.WithBlockHeight(16925943) // 2025-03-24T13:30:39.449960913Z

	jsonDistGenesis := os.MustReadFile("test_dist_genesis.json")
	jsonStakingGenesis := os.MustReadFile("test_staking_genesis.json")

	// Load faulty state from json
	var distGenesisState disttypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonDistGenesis, &distGenesisState)
	var stakingGenesisState stakingtypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonStakingGenesis, &stakingGenesisState)

	sortDelegations(stakingGenesisState.Delegations)

	// Align x/bank modules with faulty state
	for i := range distGenesisState.OutstandingRewards {
		coins, _ := distGenesisState.OutstandingRewards[i].OutstandingRewards.TruncateDecimal()
		for _, coin := range coins {
			s.FundModuleAccount(disttypes.ModuleName, coin)
		}
	}
	s.FundModuleAccount(stakingtypes.BondedPoolName, types.NewInt64Coin("stake", 1038549945))
	s.FundModuleAccount(stakingtypes.NotBondedPoolName, types.NewInt64Coin("stake", 220000))

	// Overwrite x/staking's state with imported
	s.App.StakingKeeper.InitGenesis(s.Ctx, &stakingGenesisState)

	// Overwrite x/distribution's state with faulty state
	s.App.DistrKeeper.InitGenesis(s.Ctx, distGenesisState)

	// Get validator address
	valAddrResp, err := stakingkeeper.NewQuerier(&s.App.StakingKeeper).Validators(s.Ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.Bonded.String(),
	})
	s.Require().NoError(err)

	valAddr, err := types.ValAddressFromBech32(valAddrResp.Validators[0].OperatorAddress)
	s.Require().NoError(err)

	// Verify that things are failing
	cutoutHeight := uint64(4300034)
	for _, delegation := range stakingGenesisState.Delegations {
		delAddr := types.MustAccAddressFromBech32(delegation.DelegatorAddress)

		period, err := s.App.DistrKeeper.GetDelegatorStartingInfo(s.Ctx, valAddr, delAddr)
		s.Require().NoError(err)
		s.Require().Positive(period.PreviousPeriod)
		s.Require().Positive(period.Height)

		// All delegators from before height 4300034 fail
		// See faulty_state.csv for reference
		if period.Height < cutoutHeight {
			s.Require().Panics(func() {
				_, _ = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, delAddr, valAddr)
				fmt.Printf("%s should panic (%d < %d)\n", delAddr.String(), period.Height, 5047518)
				s.Require().True(false)
			})
		} else {
			// Fork ctx to prevent modifying the state
			subCtx, _ := s.Ctx.CacheContext()

			_, err := s.App.DistrKeeper.WithdrawDelegationRewards(subCtx, delAddr, valAddr)
			s.Require().NoError(err)
		}

	}

	// Fix x/ditribution state

	// There should be anothre slashing event between blocks 4300034-5047517/periods 3893-3912 with slash fraction of 0.01%
	// See faulty_state.csv for reference
	slashingEventBlock := cutoutHeight
	slashingEventPeriod := uint64(3893)
	slashingEventFraction := sdkmath.LegacyMustNewDecFromStr("0.0001")

	err = s.App.DistrKeeper.SetValidatorSlashEvent(
		s.Ctx,
		valAddr,
		slashingEventBlock,
		slashingEventPeriod,
		disttypes.NewValidatorSlashEvent(slashingEventPeriod, slashingEventFraction),
	)
	s.Require().NoError(err)

	// Verify that things are working
	for _, delegation := range stakingGenesisState.Delegations {
		s.Require().NotPanics(func() {
			_, _ = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, types.MustAccAddressFromBech32(delegation.DelegatorAddress), valAddr)
		})
	}
}
