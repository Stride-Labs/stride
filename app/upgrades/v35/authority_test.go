package v35_test

import (
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	"github.com/Stride-Labs/stride/v35/utils"
)

// unreachableDeposit is the gov min deposit the handler writes: 1e18 ustrd (authority spec §3)
func unreachableDeposit() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(v35.GovUnreachableDeposit)))
}

// unreachableExpeditedDeposit is the gov expedited min deposit the handler writes: 2e18 ustrd
func unreachableExpeditedDeposit() sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(v35.GovUnreachableExpeditedDeposit)))
}

func (s *UpgradeTestSuite) TestSetConsensusAuthority() {
	// Mainnet has no authority field today; the test app's default params carry one, so clear it
	before, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	before.Authority = nil
	s.Require().NoError(s.App.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, before))

	s.Require().NoError(v35.SetConsensusAuthority(s.Ctx, s.App.ConsensusParamsKeeper))

	// Every other consensus param is kept exactly as read
	expected := before
	expected.Authority = &cmtproto.AuthorityParams{Authority: v35.UpgradeAuthority}
	after, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expected, after, "only the authority changes")

	// A second run (localnet re-run) leaves the same state
	s.Require().NoError(v35.SetConsensusAuthority(s.Ctx, s.App.ConsensusParamsKeeper))
	again, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(expected, again, "idempotent")
}

// An authority already set to some other address is overwritten, not kept
func (s *UpgradeTestSuite) TestSetConsensusAuthority_OverwritesExisting() {
	params, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	params.Authority = &cmtproto.AuthorityParams{Authority: "stride1otherauthority"}
	s.Require().NoError(s.App.ConsensusParamsKeeper.ParamsStore.Set(s.Ctx, params))

	s.Require().NoError(v35.SetConsensusAuthority(s.Ctx, s.App.ConsensusParamsKeeper))

	after, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v35.UpgradeAuthority, after.Authority.Authority)
}

func (s *UpgradeTestSuite) TestCloseGovSubmission() {
	// Non-default values everywhere so an accidental reset to defaults is visible
	votingPeriod := 3 * time.Hour
	expeditedVotingPeriod := time.Hour
	maxDepositPeriod := 2 * time.Hour
	before := govv1.Params{
		MinDeposit:                 sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(5))),
		ExpeditedMinDeposit:        sdk.NewCoins(sdk.NewCoin(utils.BaseStrideDenom, sdkmath.NewInt(10))),
		MaxDepositPeriod:           &maxDepositPeriod,
		VotingPeriod:               &votingPeriod,
		ExpeditedVotingPeriod:      &expeditedVotingPeriod,
		Quorum:                     "0.41",
		Threshold:                  "0.61",
		VetoThreshold:              "0.31",
		ExpeditedThreshold:         "0.71",
		MinInitialDepositRatio:     "0.21",
		ProposalCancelRatio:        "0.51",
		ProposalCancelDest:         v35.UpgradeAuthority,
		MinDepositRatio:            "0.011",
		BurnVoteQuorum:             true,
		BurnProposalDepositPrevote: true,
		BurnVoteVeto:               false,
	}
	s.Require().NoError(s.App.GovKeeper.Params.Set(s.Ctx, before))

	s.Require().NoError(v35.CloseGovSubmission(s.Ctx, s.App.GovKeeper))

	after, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(unreachableDeposit().Equal(after.MinDeposit), "min deposit %s", after.MinDeposit)
	s.Require().True(unreachableExpeditedDeposit().Equal(after.ExpeditedMinDeposit), "expedited min deposit %s", after.ExpeditedMinDeposit)
	s.Require().NoError(after.ValidateBasic(), "stored params must pass gov validation")

	// Everything but the two deposits is kept as read (compared by String: Coins hold big.Ints)
	expected := before
	expected.MinDeposit = unreachableDeposit()
	expected.ExpeditedMinDeposit = unreachableExpeditedDeposit()
	s.Require().Equal(expected.String(), after.String(), "only the deposits change")
}

func (s *UpgradeTestSuite) TestSetPOAAdmin() {
	// The test app seeds the POA admin with the gov module account, as mainnet had a different
	// multisig; either way it must end up as the upgrade authority
	before, err := s.App.POAKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotEqual(v35.UpgradeAuthority, before.Admin, "fixture starts with a different admin")

	s.Require().NoError(v35.SetPOAAdmin(s.Ctx, s.App.POAKeeper))

	after, err := s.App.POAKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(v35.UpgradeAuthority, after.Admin, "POA admin")
}
