package v35_test

import (
	"testing"

	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v34/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v34/x/autopilot/types"
	icaoracletypes "github.com/Stride-Labs/stride/v34/x/icaoracle/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// The handler must complete on a chain that has none of the mainnet state it acts on
// (every helper skips with a log); this is also the non-mainnet localnet case.
func (s *UpgradeTestSuite) TestUpgrade_EmptyState() {
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
}

// TestUpgrade runs the whole handler through the upgrade module on a state that exercises
// every step at once, and asserts the state each helper's own test checks in isolation.
func (s *UpgradeTestSuite) TestUpgrade() {
	// ----- arrange -----
	s.App.AutopilotKeeper.SetParams(s.Ctx, autopilottypes.Params{StakeibcActive: true, ClaimActive: true})
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: []string{
		"/cosmos.bank.v1beta1.MsgSend",
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}})
	deployKeyContract := s.storeAndInstantiateHackatom(sdk.MustAccAddressFromBech32(v35.WasmDeployKey))
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: v35.ComdexChainId})
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom, HostDenomOnHostZone: v35.DydxTradeRouteHostDenom,
	})
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, icaoracletypes.Oracle{ChainId: "osmosis-1", ConnectionId: "connection-9", Active: true})
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "stuevmos")
	s.seedFlaggedZone("cosmoshub-4", "connection-0", false)
	s.mockDelegationChannel("cosmoshub-4", "connection-0", "channel-863")
	s.seedQueries()
	haqqTracked, haqqTrackedTotal := s.setupHaqqHostZone()
	haqqZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	haqqZone.Validators[0].SlashQueryInProgress = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haqqZone)

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// ----- assert -----
	s.Require().False(s.App.AutopilotKeeper.GetParams(s.Ctx).StakeibcActive, "autopilot")
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/stride.stakeibc.MsgClaimUndelegatedTokens"},
		s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages, "ICA host allow-list")
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess.Addresses, "wasm upload")
	s.Require().Equal(v35.GovModuleAddress().String(), s.App.WasmKeeper.GetContractInfo(s.Ctx, deployKeyContract).Admin, "contract admin")
	comdex, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(comdex.Deprecated, "comdex")
	s.Require().Empty(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), "trade route")
	oracle, _ := s.App.ICAOracleKeeper.GetOracle(s.Ctx, "osmosis-1")
	s.Require().False(oracle.Active, "oracle")
	// Not GetAllBlacklistedDenoms: the fixture zones have a zero redemption rate, so the stakeibc
	// BeginBlocker that ConfirmUpgradeSucceeded runs after the handler halts them and re-blacklists
	// their stDenoms. The seeded denom is what the handler owns.
	s.Require().False(s.App.RatelimitKeeper.IsDenomBlacklisted(s.Ctx, "stuevmos"), "rate limiter")
	s.Require().Equal([]int64{0, 0, 0}, s.flags("cosmoshub-4"), "stale flags")
	s.Require().ElementsMatch([]string{"haqq-fee", "juno-delegation", "comdex-calibrate", "other-haqq-delegation", "other-juno-withdrawal"},
		s.queryIds(), "both ICQ purges")
	haqq, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().False(haqq.Validators[0].SlashQueryInProgress, "haqq slash flag")
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, _ := stakeibckeeper.GetValidatorFromAddress(haqq.Validators, entry.Address)
		s.Require().Equal(haqqTracked[entry.Address].Add(entry.Delta), validator.Delegation, "haqq delta %s", entry.Name)
	}
	s.Require().True(haqq.TotalDelegations.LT(haqqTrackedTotal), "haqq total dropped")
}
