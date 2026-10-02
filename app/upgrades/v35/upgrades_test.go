package v35_test

import (
	"testing"

	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"
	ratelimittypes "github.com/cosmos/ibc-go/v11/modules/apps/rate-limiting/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	v35 "github.com/Stride-Labs/stride/v35/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v35/x/autopilot/types"
	icaoracletypes "github.com/Stride-Labs/stride/v35/x/icaoracle/types"
	recordstypes "github.com/Stride-Labs/stride/v35/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v35/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v35/x/stakeibc/types"
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

	// The authority hand-off does not depend on any state and must land even here
	s.assertUpgradeAuthorityState()
}

// assertUpgradeAuthorityState checks the three param writes of authority spec §3 after the whole
// handler has run: consensus authority, gov deposits and staking max entries.
func (s *UpgradeTestSuite) assertUpgradeAuthorityState() {
	consensusParams, err := s.App.ConsensusParamsKeeper.ParamsStore.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().NotNil(consensusParams.Authority, "consensus authority set")
	s.Require().Equal(v35.UpgradeAuthority, consensusParams.Authority.Authority, "consensus authority")

	govParams, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().True(unreachableDeposit().Equal(govParams.MinDeposit), "gov min deposit %s", govParams.MinDeposit)
	s.Require().True(unreachableExpeditedDeposit().Equal(govParams.ExpeditedMinDeposit), "gov expedited min deposit %s", govParams.ExpeditedMinDeposit)

	stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint32(v35.StakingMaxEntries), stakingParams.MaxEntries, "staking max entries")
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
		sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
		sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	}})
	deployKeyContract := s.storeAndInstantiateHackatom(sdk.MustAccAddressFromBech32(v35.WasmDeployKey))
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, withInBoundsRates(stakeibctypes.HostZone{ChainId: v35.ComdexChainId}))
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, stakeibctypes.TradeRoute{
		RewardDenomOnRewardZone: v35.DydxTradeRouteRewardDenom, HostDenomOnHostZone: v35.DydxTradeRouteHostDenom,
	})
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, icaoracletypes.Oracle{ChainId: "osmosis-1", ConnectionId: "connection-9", Active: true})
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "stuevmos")
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, ratelimittypes.WhitelistedAddressPair{Sender: "sender", Receiver: "receiver"})
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, ratelimittypes.RateLimit{
		Path:  &ratelimittypes.Path{Denom: "stuatom", ChannelOrClientId: "channel-0"},
		Quota: &ratelimittypes.Quota{MaxPercentSend: sdkmath.NewInt(10), MaxPercentRecv: sdkmath.NewInt(10), DurationHours: 24},
		Flow:  &ratelimittypes.Flow{Inflow: sdkmath.ZeroInt(), Outflow: sdkmath.ZeroInt(), ChannelValue: sdkmath.NewInt(1000)},
	})
	s.seedFlaggedZone("cosmoshub-4", "connection-0", false)
	s.mockDelegationChannel("cosmoshub-4", "connection-0", "channel-863")
	s.seedQueries()
	haqqTracked, haqqTrackedTotal := s.setupHaqqHostZone()
	haqqZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	haqqZone.Validators[0].SlashQueryInProgress = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, haqqZone)
	s.setFailedLSMDeposit(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED, v35.FailedLSMDepositAmount)
	valAddr, _ := s.seedValidator(9, stakingtypes.Bonded, 1)
	delegator := apptesting.CreateRandomAccounts(1)[0]
	delegated := s.delegate(delegator, valAddr, 4_000)

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)

	// ----- assert -----
	s.Require().False(s.App.AutopilotKeeper.GetParams(s.Ctx).StakeibcActive, "autopilot")
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/stride.stakeibc.MsgClaimUndelegatedTokens", "/cosmos.staking.v1beta1.MsgUndelegate"},
		s.App.ICAHostKeeper.GetParams(s.Ctx).AllowMessages, "ICA host allow-list")
	s.Require().Equal([]string{v35.GovModuleAddress().String()}, s.App.WasmKeeper.GetParams(s.Ctx).CodeUploadAccess.Addresses, "wasm upload")
	s.Require().Equal(v35.GovModuleAddress().String(), s.App.WasmKeeper.GetContractInfo(s.Ctx, deployKeyContract).Admin, "contract admin")
	comdex, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(comdex.Deprecated, "comdex")
	s.Require().Empty(s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx), "trade route")
	oracle, _ := s.App.ICAOracleKeeper.GetOracle(s.Ctx, "osmosis-1")
	s.Require().False(oracle.Active, "oracle")
	// The fixture zones have in-bounds rates, so the stakeibc BeginBlocker that ConfirmUpgradeSucceeded
	// runs after the handler does not re-blacklist their stDenoms
	s.Require().Empty(s.App.RatelimitKeeper.GetAllRateLimits(s.Ctx), "rate limits")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllBlacklistedDenoms(s.Ctx), "blacklisted denoms")
	s.Require().Empty(s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx), "whitelisted pairs")
	s.Require().Equal([]int64{0, 0, 0}, s.flags("cosmoshub-4"), "stale flags")
	s.Require().ElementsMatch([]string{"haqq-fee", "juno-delegation", "other-haqq-delegation", "other-juno-withdrawal"},
		s.queryIds(), "all ICQ purges")
	haqq, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	s.Require().False(haqq.Validators[0].SlashQueryInProgress, "haqq slash flag")
	for _, entry := range v35.HaqqDelegationDeltas {
		validator, _, _ := stakeibckeeper.GetValidatorFromAddress(haqq.Validators, entry.Address)
		s.Require().Equal(haqqTracked[entry.Address].Add(entry.Delta), validator.Delegation, "haqq delta %s", entry.Name)
	}
	s.Require().True(haqq.TotalDelegations.LT(haqqTrackedTotal), "haqq total dropped")
	lsmDeposit := s.mustGetFailedLSMDeposit()
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_QUEUE, lsmDeposit.Status, "LSM deposit requeued")
	s.Require().Equal(int64(67_850_951), lsmDeposit.Amount.Int64(), "LSM deposit amount")
	s.assertUpgradeAuthorityState()
	s.Require().False(s.hasDelegation(delegator, valAddr), "delegation undelegated")
	entries := s.unbondingEntries(delegator, valAddr)
	s.Require().Len(entries, 1, "one unbonding entry from the handler")
	s.Require().Equal(delegated, entries[0].Balance, "unbonding balance")
}

// withInBoundsRates gives a fixture host zone a redemption rate inside its safety bounds so the
// stakeibc BeginBlocker (which runs right after the handler) does not halt it and re-blacklist its stDenom
func withInBoundsRates(hostZone stakeibctypes.HostZone) stakeibctypes.HostZone {
	hostZone.RedemptionRate = sdkmath.LegacyOneDec()
	hostZone.MinRedemptionRate = sdkmath.LegacyMustNewDecFromStr("0.9")
	hostZone.MinInnerRedemptionRate = sdkmath.LegacyMustNewDecFromStr("0.95")
	hostZone.MaxInnerRedemptionRate = sdkmath.LegacyMustNewDecFromStr("1.4")
	hostZone.MaxRedemptionRate = sdkmath.LegacyMustNewDecFromStr("1.5")
	return hostZone
}
