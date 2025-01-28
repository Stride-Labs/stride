package v25_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/Stride-Labs/stride/v25/app/apptesting"
	v25 "github.com/Stride-Labs/stride/v25/app/upgrades/v25"
	epochtypes "github.com/Stride-Labs/stride/v25/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v25/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v25/x/stakeibc/types"
	oldstaketiatypes "github.com/Stride-Labs/stride/v25/x/staketia/legacytypes"
	"github.com/Stride-Labs/stride/v25/x/staketia/types"
	staketiatypes "github.com/Stride-Labs/stride/v25/x/staketia/types"
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

	// Setup state before upgrade
	checkStaketiaMigration := s.SetupStaketiaMigration()
	checkRedemptionRatesAfterUpgrade := s.SetupTestUpdateRedemptionRateBounds()
	checkInnerRedemptionRatesAfterUpgrade := s.SetupTestUpdateInnerRedemptionRateBounds()
	checkLSMRecord := s.SetupLSMRecord()

	// Run upgrade
	s.ConfirmUpgradeSucceededs(v25.UpgradeName, upgradeHeight)

	// Validate state after upgrade
	checkStaketiaMigration()
	checkRedemptionRatesAfterUpgrade()
	checkInnerRedemptionRatesAfterUpgrade()
	checkLSMRecord()
}

func (s *UpgradeTestSuite) SetupStaketiaMigration() func() {
	delegatedBalance := sdkmath.NewInt(100)

	// Create a transfer channel (which will create a connection)
	s.CreateTransferChannel(staketiatypes.CelestiaChainId)

	// Mint stTIA for the redemption rate
	s.FundAccount(s.TestAccs[0], sdk.NewCoin("st"+staketiatypes.CelestiaNativeTokenDenom, sdk.NewInt(1000)))

	// Store the staketia host zone
	s.App.StaketiaKeeper.SetLegacyHostZone(s.Ctx, oldstaketiatypes.HostZone{
		ChainId:             staketiatypes.CelestiaChainId,
		NativeTokenDenom:    staketiatypes.CelestiaNativeTokenDenom,
		NativeTokenIbcDenom: staketiatypes.CelestiaNativeTokenIBCDenom,
		DepositAddress:      s.TestAccs[1].String(),
		TransferChannelId:   ibctesting.FirstChannelID,
		MinRedemptionRate:   sdk.MustNewDecFromStr("0.90"),
		MaxRedemptionRate:   sdk.MustNewDecFromStr("1.5"),
		RedemptionRate:      sdk.MustNewDecFromStr("1.2"),
		DelegatedBalance:    delegatedBalance,
	})

	// Create epoch trackers and EURs which are needed for the stakeibc registration
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     uint64(1),
	})
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     uint64(1),
	})
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        uint64(1),
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	// Before we call the migration function, temporarily update the variable to be connection-0 to match the above
	// and then set it back after the function call for other tests that use it
	mainnetConnectionId := staketiatypes.CelestiaConnectionId
	types.CelestiaConnectionId = ibctesting.FirstConnectionID

	// Return a callback to check the state after the upgrade
	return func() {
		// Set back the connectionID
		types.CelestiaConnectionId = mainnetConnectionId

		// Confirm the stakeibc host zone was created
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, staketiatypes.CelestiaChainId)
		s.Require().True(found, "Host zone should have been found")
		s.Require().Equal(delegatedBalance, hostZone.TotalDelegations, "Delegated balance")

		// Confirm the validator set was registered
		s.Require().Equal(len(hostZone.Validators), len(v25.Validators), "Number of validators")
		s.Require().Equal(hostZone.Validators[0].Address, "celestiavaloper1uvytvhunccudw8fzaxvsrumec53nawyj939gj9", "First validator")
	}
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v25.OsmosisChainId,
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdk.MustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMaxOuterRedemptionRate: sdk.MustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
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

func (s *UpgradeTestSuite) SetupTestUpdateInnerRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	// Celestia already set with rr of 1.2
	testCases := []UpdateRedemptionRateInnerBounds{
		{
			ChainId:                        "celestia",
			CurrentRedemptionRate:          sdk.MustNewDecFromStr("1.2"),
			ExpectedMinInnerRedemptionRate: sdk.MustNewDecFromStr("1.1988"), // 1.2-(1.2*.001) = 1.1988
			ExpectedMaxInnerRedemptionRate: sdk.MustNewDecFromStr("1.2012"), // 1.2+(1.2*.001) = 1.2012
		},
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm they were all updated
		for _, tc := range testCases {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinInnerRedemptionRate, hostZone.MinInnerRedemptionRate, "%s - min inner", tc.ChainId)
			s.Require().Equal(tc.ExpectedMaxInnerRedemptionRate, hostZone.MaxInnerRedemptionRate, "%s - max inner", tc.ChainId)
		}
	}
}

func (s *UpgradeTestSuite) SetupLSMRecord() func() {
	initialDetokenizeAmount := sdkmath.NewInt(100)
	expectedDetokenizeAmount := sdkmath.NewInt(99)

	// Create the failed detokenization record
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, recordtypes.LSMTokenDeposit{
		ChainId: v25.CosmosChainId,
		Denom:   v25.FailedLSMDepositDenom,
		Amount:  initialDetokenizeAmount,
	})

	return func() {
		// Confirm the lsm deposit record was reset
		lsmRecord, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, v25.CosmosChainId, v25.FailedLSMDepositDenom)
		s.Require().True(found, "lsm deposit record should have been found")
		s.Require().Equal(recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, lsmRecord.Status, "lsm record status")
		s.Require().Equal(expectedDetokenizeAmount, lsmRecord.Amount, "lsm deposit record amount")
	}
}
