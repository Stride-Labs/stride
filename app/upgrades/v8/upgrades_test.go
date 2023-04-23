package v8_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	v8 "github.com/Stride-Labs/stride/v9/app/upgrades/v8"
	autopilottypes "github.com/Stride-Labs/stride/v9/x/autopilot/types"
	"github.com/Stride-Labs/stride/v9/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
)

var (
	ustrd               = "ustrd"
	dummyUpgradeHeight  = int64(5)
	osmoAirdropId       = "osmosis"
	unofficialAirdropId = "unofficial-airdrop"
	addresses           = []string{
		"stride12a06af3mm5j653446xr4dguacuxfkj293ey2vh",
		"stride1udf2vyj5wyjckl7nzqn5a2vh8fpmmcffey92y8",
		"stride1uc8ccxy5s2hw55fn8963ukfdycaamq95jqcfnr",
	}
	weights = []sdk.Dec{
		sdk.NewDec(1000),
		sdk.NewDec(2000),
		sdk.NewDec(3000),
	}
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
	s.Setup()

	s.SetupStoreBeforeUpgrade()
	s.ConfirmUpgradeSucceededs("v8", dummyUpgradeHeight)
	s.CheckStoreAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupStoreBeforeUpgrade() {
	// Add a test aidrop to the store
	params := claimtypes.Params{
		Airdrops: []*claimtypes.Airdrop{
			{
				AirdropIdentifier: osmoAirdropId,
				ClaimedSoFar:      sdkmath.NewInt(1000000),
			},
			{
				AirdropIdentifier: unofficialAirdropId, // this should be removed
				ClaimedSoFar:      sdkmath.NewInt(1000000),
			},
		},
	}
	err := s.App.ClaimKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err, "no error expected when setting claim params")

	// Add claim records for the airdrop
	claimRecords := []claimtypes.ClaimRecord{
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           addresses[0],
			Weight:            weights[0],
			ActionCompleted:   []bool{true, true, false},
		},
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           addresses[1],
			Weight:            weights[1],
			ActionCompleted:   []bool{true, true, true},
		},
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           addresses[2],
			Weight:            weights[2],
			ActionCompleted:   []bool{false, false, false},
		},
	}
	err = s.App.ClaimKeeper.SetClaimRecords(s.Ctx, claimRecords)
	s.Require().NoError(err, "no error expected when setting claim record")

	// Set vesting to 0s
	types.DefaultVestingInitialPeriod, err = time.ParseDuration("0s")
	s.Require().NoError(err, "no error expected when setting vesting initial period")
}

func (s *UpgradeTestSuite) CheckStoreAfterUpgrade() {
	afterCtx := s.Ctx.WithBlockHeight(dummyUpgradeHeight)

	// Check that the evmos airdrop was added and the unofficial airdrop was removed
	claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	s.Require().Len(claimParams.Airdrops, 2, "there should be only two airdrops (evmos and osmo)")
	osmoAirdrop := claimParams.Airdrops[0]
	evmosAirdrop := claimParams.Airdrops[1]

	// Check that the params of the osmo airdrop were reset
	s.Require().Equal(osmoAirdropId, osmoAirdrop.AirdropIdentifier, "osmo airdrop identifier")
	s.Require().Zero(osmoAirdrop.ClaimedSoFar.Int64(), "osmo claimed so far")

	// Check that the params of the evmos airdrop were initialized
	s.Require().Equal(v8.EvmosAirdropIdentifier, evmosAirdrop.AirdropIdentifier, "evmos airdrop identifier")
	s.Require().Zero(evmosAirdrop.ClaimedSoFar.Int64(), "evmos claimed so far")
	s.Require().Equal(v8.EvmosAirdropDistributor, evmosAirdrop.DistributorAddress, "evmos airdrop distributor")
	s.Require().Equal(v8.AirdropDuration, evmosAirdrop.AirdropDuration, "evmos airdrop duration")
	s.Require().Equal(ustrd, evmosAirdrop.ClaimDenom, "evmos airdrop claim denom")
	s.Require().Equal(v8.AirdropStartTime, evmosAirdrop.AirdropStartTime, "evmos airdrop start time")

	// Check that the evmos claims records were added
	evmosClaimRecords := s.App.ClaimKeeper.GetClaimRecords(afterCtx, v8.EvmosAirdropIdentifier)
	s.Require().Positive(len(evmosClaimRecords))

	// Check that the osmo claim actions were reset
	osmoClaimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, osmoAirdropId)
	s.Require().Equal(len(osmoClaimRecords), 3, "claim records length")

	fullyResetAction := []bool{false, false, false}
	for i, claimRecord := range osmoClaimRecords {
		s.Require().Equal(fullyResetAction, claimRecord.ActionCompleted, "record %d reset", i)
		s.Require().Equal(addresses[i], claimRecord.Address, "record %d address", i)
		s.Require().Equal(weights[i], claimRecord.Weight, "record %d weight", i)
	}

	// Check autopilot params
	expectedAutoPilotParams := autopilottypes.Params{
		StakeibcActive: false,
		ClaimActive:    true,
	}
	actualAutopilotParams := s.App.AutopilotKeeper.GetParams(s.Ctx)
	s.Require().Equal(expectedAutoPilotParams, actualAutopilotParams, "autopilot params")
}
