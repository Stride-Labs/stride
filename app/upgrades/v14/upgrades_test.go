package v14_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v13/app/apptesting"
	v14 "github.com/Stride-Labs/stride/v13/app/upgrades/v14"
	claimtypes "github.com/Stride-Labs/stride/v13/x/claim/types"
)

var (
	dummyUpgradeHeight = int64(5)

	osmoAirdropId = "osmosis"
	ustrd         = "ustrd"
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
	s.SetupStoreBeforeUpgrade()
	s.ConfirmUpgradeSucceededs("v14", dummyUpgradeHeight)
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
	err := s.App.ClaimKeeper.SetParams(s.Ctx, params)
	s.Require().NoError(err, "no error expected when setting claim params")
	// Set vesting to 0s
	claimtypes.DefaultVestingInitialPeriod, err = time.ParseDuration("0s")
	s.Require().NoError(err, "no error expected when setting vesting initial period")
}

func (s *UpgradeTestSuite) CheckStoreAfterUpgrade() {
	afterCtx := s.Ctx.WithBlockHeight(dummyUpgradeHeight)

	// Check that all airdrops were added, osmosis airdrop wasn't removed
	claimParams, err := s.App.ClaimKeeper.GetParams(s.Ctx)
	s.Require().NoError(err, "no error expected when getting params")
	s.Require().Len(claimParams.Airdrops, 5, "there should be exactly 5 airdrops")

	// ------ OSMO -------
	osmoAirdrop := claimParams.Airdrops[0]
	s.Require().Equal(osmoAirdropId, osmoAirdrop.AirdropIdentifier, "osmo airdrop identifier") // verify this wasn't deleted

	// ------ INJECTIVE -------
	injectiveAirdrop := claimParams.Airdrops[1]
	s.CheckAirdropAdded(afterCtx, injectiveAirdrop, v14.InjectiveAirdropDistributor, v14.InjectiveAirdropIdentifier, v14.InjectiveChainId, true)

	// ------ COMDEX -------
	comdexAirdrop := claimParams.Airdrops[2]
	s.CheckAirdropAdded(afterCtx, comdexAirdrop, v14.ComdexAirdropDistributor, v14.ComdexAirdropIdentifier, v14.ComdexChainId, false)

	// ------ SOMM -------
	sommAirdrop := claimParams.Airdrops[3]
	s.CheckAirdropAdded(afterCtx, sommAirdrop, v14.SommAirdropDistributor, v14.SommAirdropIdentifier, v14.SommChainId, false)

	// ------ UMEE -------
	umeeAirdrop := claimParams.Airdrops[4]
	s.CheckAirdropAdded(afterCtx, umeeAirdrop, v14.UmeeAirdropDistributor, v14.UmeeAirdropIdentifier, v14.UmeeChainId, false)
}

func (s *UpgradeTestSuite) CheckAirdropAdded(ctx sdk.Context, airdrop *claimtypes.Airdrop, distributor string, identifier string, chainId string, autopilotEnabled bool) {
	// Check that the params of the airdrop were initialized
	s.Require().Equal(identifier, airdrop.AirdropIdentifier, fmt.Sprintf("%s airdrop identifier", identifier))
	s.Require().Equal(chainId, airdrop.ChainId, fmt.Sprintf("%s airdrop chain-id", identifier))
	s.Require().Zero(airdrop.ClaimedSoFar.Int64(), fmt.Sprintf("%s claimed so far", identifier))
	s.Require().Equal(distributor, airdrop.DistributorAddress, fmt.Sprintf("%s airdrop distributor", identifier))
	s.Require().Equal(v14.AirdropDuration, airdrop.AirdropDuration, fmt.Sprintf("%s airdrop duration", identifier))
	s.Require().Equal(ustrd, airdrop.ClaimDenom, fmt.Sprintf("%s airdrop claim denom", identifier))
	s.Require().Equal(v14.AirdropStartTime, airdrop.AirdropStartTime, fmt.Sprintf("%s airdrop start time", identifier))
	s.Require().Equal(autopilotEnabled, airdrop.AutopilotEnabled, fmt.Sprintf("%s airdrop autopilot enabled", identifier))

	claimRecords := s.App.ClaimKeeper.GetClaimRecords(ctx, identifier)
	s.Require().Positive(len(claimRecords), fmt.Sprintf("there should be at least one claim record for %s", identifier))

	// Check that an epoch was created
	epochInfo, found := s.App.EpochsKeeper.GetEpochInfo(ctx, fmt.Sprintf("airdrop-%s", identifier))
	s.Require().True(found, "epoch tracker should be found")
	s.Require().Zero(epochInfo.CurrentEpoch, "epoch should be zero")
	s.Require().Equal(epochInfo.Duration, claimtypes.DefaultEpochDuration, "epoch duration should be equal to airdrop duration")
}
