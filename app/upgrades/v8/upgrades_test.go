package v8_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v7/app/apptesting"
	v8 "github.com/Stride-Labs/stride/v7/app/upgrades/v8"
	"github.com/Stride-Labs/stride/v7/x/claim/types"
	claimtypes "github.com/Stride-Labs/stride/v7/x/claim/types"
)

var (
	dummyUpgradeHeight = int64(5)
	osmoAirdropId      = "osmosis"
	dummyAddresses     = []string{
		"stride12a06af3mm5j653446xr4dguacuxfkj293ey2vh",
		"stride1udf2vyj5wyjckl7nzqn5a2vh8fpmmcffey92y8",
		"stride1uc8ccxy5s2hw55fn8963ukfdycaamq95jqcfnr",
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
		},
	}
	s.App.ClaimKeeper.SetParams(s.Ctx, params)

	// Add claim records for the airdrop
	claimRecords := []claimtypes.ClaimRecord{
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           dummyAddresses[0],
			Weight:            sdk.NewDec(1000),
			ActionCompleted:   []bool{true, true, false},
		},
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           dummyAddresses[1],
			Weight:            sdk.NewDec(50),
			ActionCompleted:   []bool{true, true, true},
		},
		{
			AirdropIdentifier: osmoAirdropId,
			Address:           dummyAddresses[2],
			Weight:            sdk.NewDec(100),
			ActionCompleted:   []bool{false, false, false},
		},
	}
	err := s.App.ClaimKeeper.SetClaimRecords(s.Ctx, claimRecords)
	s.Require().NoError(err, "no error expected when setting claim record")

	// Set vesting to 0s
	types.DefaultVestingInitialPeriod, err = time.ParseDuration("0s")
	s.Require().NoError(err, "no error expected when setting vesting initial period")
}

func (s *UpgradeTestSuite) CheckStoreAfterUpgrade() {
	afterCtx := s.Ctx.WithBlockHeight(dummyUpgradeHeight)

	// Check that the evmos airdrop was added
	evmosClaimRecords := s.App.ClaimKeeper.GetClaimRecords(afterCtx, v8.EvmosAirdropIdentifier)
	s.Require().Positive(len(evmosClaimRecords))

	// Check that the osmo airdrop params were reset
	claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	s.Require().Equal(osmoAirdropId, claimParams.Airdrops[0].AirdropIdentifier, "airdrop identifier")
	s.Require().Zero(claimParams.Airdrops[0].ClaimedSoFar.Int64(), "claimed so far")

	// Check that the claim actions were reset
	osmoClaimRecords := s.App.ClaimKeeper.GetClaimRecords(s.Ctx, osmoAirdropId)
	s.Require().Equal(len(osmoClaimRecords), 3, "claim records length")

	fullyResetAction := []bool{false, false, false}
	for i, claimRecord := range osmoClaimRecords {
		s.Require().Equal(fullyResetAction, claimRecord.ActionCompleted, "record %d reset", i)
		s.Require().Equal(dummyAddresses[i], claimRecord.Address, "record %d address", i)
		s.Require().Equal(sdk.NewDec(1000), claimRecord.Weight, "record %d weight", i)
	}
}
