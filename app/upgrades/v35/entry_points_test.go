package v35_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icahosttypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/host/types"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	autopilottypes "github.com/Stride-Labs/stride/v34/x/autopilot/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) TestDisableAutopilotStakeibc() {
	s.App.AutopilotKeeper.SetParams(s.Ctx, autopilottypes.Params{StakeibcActive: true, ClaimActive: true})

	v35.DisableAutopilotStakeibc(s.Ctx, s.App.AutopilotKeeper)

	params := s.App.AutopilotKeeper.GetParams(s.Ctx)
	s.Require().False(params.StakeibcActive, "stakeibc autopilot should be off")
	s.Require().True(params.ClaimActive, "claim autopilot should be untouched")
}

func (s *UpgradeTestSuite) TestRemoveStakeibcFromICAHostAllowList() {
	// The mainnet list (queried 2026-09-29), in its mainnet order
	before := []string{
		sdk.MsgTypeURL(&banktypes.MsgSend{}),
		sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/cosmos.gov.v1beta1.MsgVote",
		sdk.MsgTypeURL(&stakeibctypes.MsgLiquidStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgRedeemStake{}),
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: before})

	v35.RemoveStakeibcFromICAHostAllowList(s.Ctx, s.App.ICAHostKeeper)

	after := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().True(after.HostEnabled, "host stays enabled")
	s.Require().Equal(append(append([]string{}, before[:9]...), before[11]), after.AllowMessages,
		"exactly the liquid stake and redeem stake entries are removed, order preserved")
}

// Running the removal twice must be a no-op the second time (idempotent for a re-run on localnet)
func (s *UpgradeTestSuite) TestRemoveStakeibcFromICAHostAllowList_Idempotent() {
	s.App.ICAHostKeeper.SetParams(s.Ctx, icahosttypes.Params{HostEnabled: true, AllowMessages: []string{
		"/cosmos.bank.v1beta1.MsgSend",
		sdk.MsgTypeURL(&stakeibctypes.MsgClaimUndelegatedTokens{}),
	}})

	v35.RemoveStakeibcFromICAHostAllowList(s.Ctx, s.App.ICAHostKeeper)

	after := s.App.ICAHostKeeper.GetParams(s.Ctx)
	s.Require().Equal([]string{"/cosmos.bank.v1beta1.MsgSend", "/stride.stakeibc.MsgClaimUndelegatedTokens"}, after.AllowMessages)
}
