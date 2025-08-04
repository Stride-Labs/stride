package v28_test

import (
	"encoding/base64"
	"fmt"
	"sort"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	v28 "github.com/Stride-Labs/stride/v28/app/upgrades/v28"
	icqtypes "github.com/Stride-Labs/stride/v28/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinOuterRedemptionRate sdkmath.LegacyDec
	ExpectedMaxOuterRedemptionRate sdkmath.LegacyDec
}

type UpdateRedemptionRateInnerBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinInnerRedemptionRate sdkmath.LegacyDec
	ExpectedMaxInnerRedemptionRate sdkmath.LegacyDec
}

var (
	// variables for Prop 262 Testing
	Strd = "ustrd"

	ReceivingInitialBalance  = sdkmath.NewInt(1_000_000)
	IncentivesInitialBalance = sdkmath.NewInt(11_000_000_000_000)
	SecurityInitialBalance   = sdkmath.NewInt(1_481_637_000_000)

	ReceivingExpectedBalance  = sdkmath.NewInt(9_481_001_000_000)
	IncentivesExpectedBalance = sdkmath.NewInt(3_000_000_000_000)
	SecurityExpectedBalance   = sdkmath.NewInt(637_000_000)
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
	// Set state before upgrade
	checkRedemptionRates := s.SetupTestUpdateRedemptionRateBounds()
	checkICQStore := s.SetupTestICQStore()
	checkMaxIcas := s.SetupTestMaxIcasBand()
	checkActionGovProp262 := s.TestActionGovProp262()

	// Run upgrade
	s.ConfirmUpgradeSucceeded(v28.UpgradeName)

	// Confirm state after upgrade
	checkRedemptionRates()
	s.checkConsumerParams()
	checkICQStore()
	checkMaxIcas()
	checkActionGovProp262()
}

func (s *UpgradeTestSuite) SetupTestMaxIcasBand() func() {
	// Create a host zone for band
	hostZone := stakeibctypes.HostZone{
		ChainId: v28.BandChainId,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Return callback to check store after upgrade
	return func() {
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v28.BandChainId)
		s.Require().True(found)
		s.Require().Equal(v28.MaxMessagesPerIca, hostZone.MaxMessagesPerIcaTx)
	}
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.5"), // 1 - 50% = 0.95
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("2.0"), // 1 + 100% = 1.25
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.55"), // 1.1 - 50% = 0.55
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("2.2"),  // 1.1 + 100% = 2.2
		},
	}

	// Create a host zone for each test case
	for _, tc := range testCases {
		hostZone := stakeibctypes.HostZone{
			ChainId:        tc.ChainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm they were all updated
		for _, tc := range testCases {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "%s - min outer", tc.ChainId)
			s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
		}
	}
}

func (s *UpgradeTestSuite) checkConsumerParams() {
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
	err := v28.ApplyDistributionFix(s.Ctx, s.App.DistrKeeper)
	s.Require().NoError(err)

	// After applying the fix, all delegations should be able to withdraw rewards without panics
	for _, delegation := range stakingGenesisState.Delegations {
		s.Require().NotPanics(func() {
			_, err = s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, types.MustAccAddressFromBech32(delegation.DelegatorAddress), valAddr)
			s.Require().NoError(err)
		})
	}
}

func (s *UpgradeTestSuite) SetupTestICQStore() func() {
	// Create the ICQ Query in the store
	// And create a mock Host Zone with the relevant validator

	// -- create the ICQ Query --
	icqQueries := []icqtypes.Query{
		{
			Id: "2c39af4c3d2ecb96d8bbf7f3386468c5909e51fe3364b8d1f9d6fce173dd1f7a",
		},
		{
			Id: "some_other_id",
		},
	}

	for _, icqQuery := range icqQueries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqQuery)
	}

	// -- create the Host Zone --
	hostZone := stakeibctypes.HostZone{
		ChainId: "evmos_9001-2",
	}

	// Create list of Validators to add to the Host Zone
	validators := []*stakeibctypes.Validator{
		{
			// Should get set to false
			Address:              v28.QueryValidatorAddress,
			SlashQueryInProgress: true,
		},
		{
			Address:              "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vFIRST",
			SlashQueryInProgress: true,
		},
		{
			Address:              "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vSECND",
			SlashQueryInProgress: false,
		},
	}
	hostZone.Validators = validators

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Return callback to check ICQ store after upgrade
	return func() {
		/// -- verify SlashQueryInProgress is modified correctly --
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "evmos_9001-2")
		s.Require().True(found)
		s.Require().Equal(v28.QueryValidatorAddress, hostZone.Validators[0].Address)
		s.Require().Equal(false, hostZone.Validators[0].SlashQueryInProgress)
		s.Require().Equal(true, hostZone.Validators[1].SlashQueryInProgress)
		s.Require().Equal(false, hostZone.Validators[2].SlashQueryInProgress)

		// -- verify ICQ Query is deleted --
		icqQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
		s.Require().Equal(1, len(icqQueries))
		s.Require().Equal("some_other_id", icqQueries[0].Id)
	}
}

func (s *UpgradeTestSuite) TestStuckQueryRequestData() {
	_, validatorAddressBz, _ := bech32.DecodeAndConvert(v28.QueryValidatorAddress)
	_, delegatorAddressBz, _ := bech32.DecodeAndConvert(v28.EvmosDelegationIca)
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)
	s.Require().Equal(base64.StdEncoding.EncodeToString(queryData), "MSBuvLM8WbdQm7tYvdAu6Bu5OtoAIx8fN3RBNSB6fa911RRbYQruJvSIXf8h2priHOp//cZrag==")
}

func (s *UpgradeTestSuite) TestActionGovProp262() func() {
	incentivesAddress := sdk.MustAccAddressFromBech32(v28.IncentivesAddress)
	receivingAddress262 := sdk.MustAccAddressFromBech32(v28.ReceivingAddress262)
	securityAddress := sdk.MustAccAddressFromBech32(v28.SecurityAddress)

	// Fund accounts
	s.FundAccount(incentivesAddress, sdk.NewCoin(Strd, IncentivesInitialBalance))
	s.FundAccount(receivingAddress262, sdk.NewCoin(Strd, ReceivingInitialBalance))
	s.FundAccount(securityAddress, sdk.NewCoin(Strd, SecurityInitialBalance))

	// Return callback to check balances
	return func() {
		receivingBalance := s.App.BankKeeper.GetBalance(s.Ctx, receivingAddress262, Strd)
		s.Require().Equal(ReceivingExpectedBalance, receivingBalance.Amount)

		incentivesBalance := s.App.BankKeeper.GetBalance(s.Ctx, incentivesAddress, Strd)
		s.Require().Equal(IncentivesExpectedBalance, incentivesBalance.Amount)

		securityBalance := s.App.BankKeeper.GetBalance(s.Ctx, securityAddress, Strd)
		s.Require().Equal(SecurityExpectedBalance, securityBalance.Amount)
	}
}
