package v15_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v21/app/apptesting"
	v15 "github.com/Stride-Labs/stride/v21/app/upgrades/v15"
	icqtypes "github.com/Stride-Labs/stride/v21/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v21/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v21/x/stakeibc/types"
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

type UpdateRedemptionRateBounds struct {
	CurrentRedemptionRate          sdk.Dec
	ExpectedMinOuterRedemptionRate sdk.Dec
	ExpectedMinInnerRedemptionRate sdk.Dec
	ExpectedMaxInnerRedemptionRate sdk.Dec
	ExpectedMaxOuterRedemptionRate sdk.Dec
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	// Setup the store before the upgrade
	checkRedemptionRatesAfterUpgrade := s.SetupRedemptionRatesBeforeUpgrade()
	checkQueriesAfterUpgrade := s.SetupQueriesBeforeUpgrade()

	// Run the upgrade to set the bounds and clear pending queries
	s.ConfirmUpgradeSucceededs("v15", dummyUpgradeHeight)

	// Check the store after the upgrade
	checkRedemptionRatesAfterUpgrade()
	checkQueriesAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupRedemptionRatesBeforeUpgrade() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("0.97"), // 1 - 3% = 0.97
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.05"), // 1 + 5% = 1.05
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.067"), // 1.1 - 3% = 1.067
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.155"), // 1.1 + 5% = 1.155
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.2125"), // 1.25 - 3% = 1.2125
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.3125"), // 1.25 + 5% = 1.3125
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.3750"), // 1.25 + 10% = 1.375
		},
	}

	// Create a host zone for each test case
	for i, tc := range testCases {
		chainId := fmt.Sprintf("chain-%d", i)

		hostZone := stakeibctypes.HostZone{
			ChainId:        chainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Create an evmos host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: v15.EvmosChainId,
	})

	// Return callback function to chck that bounds were set
	return func() {
		// Confirm the correct bounds were set
		for i, tc := range testCases {
			chainId := fmt.Sprintf("chain-%d", i)

			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "min outer")
			s.Require().Equal(tc.ExpectedMinInnerRedemptionRate, hostZone.MinInnerRedemptionRate, "min inner")
			s.Require().Equal(tc.ExpectedMaxInnerRedemptionRate, hostZone.MaxInnerRedemptionRate, "max inner")
			s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "max outer")
		}

		// Confirm evmos' custom bounds were set
		evmosHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v15.EvmosChainId)
		s.Require().True(found)

		s.Require().Equal(v15.EvmosOuterMinRedemptionRate, evmosHostZone.MinRedemptionRate, "min outer")
		s.Require().Equal(v15.EvmosInnerMinRedemptionRate, evmosHostZone.MinInnerRedemptionRate, "min inner")
		s.Require().Equal(v15.EvmosMaxRedemptionRate, evmosHostZone.MaxInnerRedemptionRate, "max inner")
		s.Require().Equal(v15.EvmosMaxRedemptionRate, evmosHostZone.MaxRedemptionRate, "max outer")
	}
}

func (s *UpgradeTestSuite) SetupQueriesBeforeUpgrade() func() {
	// Set pending queries of different types
	queries := []icqtypes.Query{
		{Id: "1", CallbackId: stakeibckeeper.ICQCallbackID_Validator},
		{Id: "2", CallbackId: stakeibckeeper.ICQCallbackID_Delegation}, // deleted
		{Id: "3", CallbackId: stakeibckeeper.ICQCallbackID_Delegation}, // deleted
		{Id: "4", CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
	}
	expectedQueriesAfterUpgrade := []string{"1", "4"}

	for _, query := range queries {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}

	// Return callback function to check that queries were removed
	return func() {
		queryIds := []string{}
		for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
			queryIds = append(queryIds, query.Id)
		}

		s.Require().Len(queryIds, 2)
		s.Require().ElementsMatch(queryIds, expectedQueriesAfterUpgrade)
	}
}
