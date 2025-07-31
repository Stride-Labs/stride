package v28_test

import (
	"encoding/base64"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	v28 "github.com/Stride-Labs/stride/v27/app/upgrades/v28"
	icqtypes "github.com/Stride-Labs/stride/v27/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
}

type UpdateRedemptionRateInnerBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinInnerRedemptionRate sdk.Dec
	ExpectedMaxInnerRedemptionRate sdk.Dec
}

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

	// Set state before upgrade
	checkRedemptionRates := s.SetupTestUpdateRedemptionRateBounds()
	checkICQStore := s.SetupTestICQStore()

	// Run upgrade
	s.ConfirmUpgradeSucceededs(v28.UpgradeName, upgradeHeight)

	// Confirm state after upgrade
	checkRedemptionRates()

	// Check ICQ Store
	checkICQStore()
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.5"), // 1 - 50% = 0.95
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("2.0"), // 1 + 100% = 1.25
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.55"), // 1.1 - 50% = 0.55
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("2.2"),  // 1.1 + 100% = 2.2
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
			Address:              "evmosvaloper1tdss4m3x7jy9mlepm2dwy8820l7uv6m2vx6z88",
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
