package keeper_test

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	hosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	_ "github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	epochtypes "github.com/Stride-Labs/stride/v6/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v6/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
)

var (
	hostModuleAddress = stakeibctypes.NewZoneAddress(HostChainId)
)

type RewardAllocationTestCase struct {
	hz                 stakeibctypes.HostZone
	channel            Channel
	withdrawalAcctBal  sdkmath.Int
	withdrawalAcctCoin sdk.Coin
	timeoutHeight      clienttypes.Height
}

func (s *KeeperTestSuite) SetupWithdrawAccount() RewardAllocationTestCase {
	// Set up host zone withdrawal ICA
	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	withdrawalChannelID := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]

	// Set up ibc denom
	ibcDenomTrace := s.GetIBCDenomTrace(Atom) // we need a true IBC denom here
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, ibcDenomTrace)

	// Fund withdrawal ica (mint -> send to withdrawal ica)
	initialModuleAccountBal := sdkmath.NewInt(15_000)
	initialModuleAccountCoin := sdk.NewCoin(Atom, initialModuleAccountBal)
	s.FundAccount(sdk.MustAccAddressFromBech32(withdrawalAddress), initialModuleAccountCoin)
	err := s.HostApp.BankKeeper.MintCoins(s.HostChain.GetContext(), minttypes.ModuleName, sdk.NewCoins(initialModuleAccountCoin))
	s.Require().NoError(err)
	err = s.HostApp.BankKeeper.SendCoinsFromModuleToAccount(s.HostChain.GetContext(), minttypes.ModuleName, sdk.MustAccAddressFromBech32(withdrawalAddress), sdk.NewCoins(initialModuleAccountCoin))
	s.Require().NoError(err)

	// Allow ica call ibc transfer in host chain
	s.HostApp.ICAHostKeeper.SetParams(s.HostChain.GetContext(), hosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			"/ibc.applications.transfer.v1.MsgTransfer",
		},
	})

	// Set up host zone
	hostZone := stakeibctypes.HostZone{
		ChainId: HostChainId,
		Address: hostModuleAddress.String(),
		WithdrawalAccount: &stakeibctypes.ICAAccount{
			Address: withdrawalAddress,
			Target:  stakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		ConnectionId:      ibctesting.FirstConnectionID,
		TransferChannelId: ibctesting.FirstChannelID,
		HostDenom:         Atom,
		IbcDenom:          ibcDenomTrace.IBCDenom(),
		RedemptionRate:    sdk.OneDec(),
	}

	currentEpoch := uint64(2)
	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}
	mintEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.MINT_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 60_000_000_000), // dictates timeouts
	}

	// we need a deposit record to liquid stake 
	initialDepositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 2,
		HostZoneId:         "GAIA",
		Amount:             sdkmath.ZeroInt(),
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, mintEpochTracker)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, initialDepositRecord)

	return RewardAllocationTestCase{
		hz: hostZone,
		channel: Channel{
			PortID:    icatypes.PortPrefix + withdrawalAccountOwner,
			ChannelID: withdrawalChannelID,
		},
		withdrawalAcctBal:  initialModuleAccountBal,
		withdrawalAcctCoin: initialModuleAccountCoin,
		timeoutHeight:      clienttypes.NewHeight(1, 100),
	}
}

// Test the process of sweeping staking rewards from the withdrawal account to the Reward Collector account
func (s *KeeperTestSuite) TestIbcTransferRewardsFromWithdrawalToRewardCollector() {
	// Setup
	tc := s.SetupWithdrawAccount()

	rewardCollector := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName)
	// Send msgs to withdraw ICA that would occur in the icq withdrawal callback, to test ibc-transfer from hostzone withdrawal acct -> stride rewardscollector
	var msgs []sdk.Msg
	ibcTransferTimeoutNanos := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(s.HostChain.GetContext().BlockTime().UnixNano()) + ibcTransferTimeoutNanos
	msg := ibctypes.NewMsgTransfer("transfer", "channel-0", tc.withdrawalAcctCoin, tc.hz.WithdrawalAccount.Address, rewardCollector.GetAddress().String(), tc.timeoutHeight, timeoutTimestamp)
	msgs = append(msgs, msg)
	data, _ := icatypes.SerializeCosmosTx(s.App.AppCodec(), msgs)
	icaTimeOutNanos := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeyICATimeoutNanos)
	icaTimeoutTimestamp := uint64(s.StrideChain.GetContext().BlockTime().UnixNano()) + icaTimeOutNanos

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}
	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		tc.channel.PortID,
		tc.channel.ChannelID,
		s.TransferPath.EndpointB.ChannelConfig.PortID,
		s.TransferPath.EndpointB.ChannelID,
		tc.timeoutHeight,
		timeoutTimestamp,
	)
	_, err := s.App.StakeibcKeeper.SubmitTxs(s.Ctx, tc.hz.ConnectionId, msgs, *tc.hz.WithdrawalAccount, icaTimeoutTimestamp, "", nil)
	s.Require().NoError(err)

	// Simulate the process of receiving ica packets on the hostchain
	module, _, err := s.HostChain.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(s.HostChain.GetContext(), "icahost")
	s.Require().NoError(err)
	cbs, ok := s.HostChain.App.GetIBCKeeper().Router.GetRoute(module)
	s.Require().True(ok)
	cbs.OnRecvPacket(s.HostChain.GetContext(), packet, nil)

	// After withdraw ica send ibc transfer, simulate receiving transfer packet at stride
	transferPacketData := ibctypes.NewFungibleTokenPacketData(
		Atom, tc.withdrawalAcctBal.String(), tc.hz.WithdrawalAccount.Address, rewardCollector.GetAddress().String(),
	)
	transferPacketData.Memo = ""
	transferPacket := channeltypes.NewPacket(
		transferPacketData.GetBytes(),
		1,
		s.TransferPath.EndpointB.ChannelConfig.PortID,
		s.TransferPath.EndpointB.ChannelID,
		s.TransferPath.EndpointA.ChannelConfig.PortID,
		s.TransferPath.EndpointA.ChannelID,
		tc.timeoutHeight,
		timeoutTimestamp,
	)
	cbs, ok = s.StrideChain.App.GetIBCKeeper().Router.GetRoute("transfer")
	s.Require().True(ok)
	cbs.OnRecvPacket(s.StrideChain.GetContext(), transferPacket, nil)

	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal(tc.withdrawalAcctBal, rewardedTokens.AmountOf(tc.hz.IbcDenom))
}

// Test the process of liquid staking staking rewards then sweeping them to the Fee Collector
func (s *KeeperTestSuite) TestLiquidStakeAndSweepSuccess() {
	tc := s.SetupWithdrawAccount()

	// Fund reward collector account with ibc'd reward tokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(tc.hz.IbcDenom, tc.withdrawalAcctBal))

	// Liquid stake all hostzone token then get stTokens back
	rewardsFound := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().True(rewardsFound)

	// Reward Collector acct should have no more ibc/XXX tokens after liquid staking them all
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal("0", rewardedTokens.AmountOf(tc.hz.IbcDenom).String())

	// Sweep stTokens from Reward Collector to Fee Collector
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Fee Collector acct should have stTokens after they're swept there from Reward Collector
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName).GetAddress()
	liquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollectorAddress)
	s.Require().Equal(tc.withdrawalAcctBal.String(), liquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorIbcTokensNoIbcTokens() {
	_ = s.SetupWithdrawAccount()

	// Fund reward collector account with non-ibc reward tokens
	nonIbcTokenDenom := "XXX"
	nonIbctokenAmt := sdk.NewInt(10)
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(nonIbcTokenDenom, nonIbctokenAmt))

	// Liquid stake all hostzone token then get stTokens back
	rewardsFound := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().False(rewardsFound)

	// Reward Collector acct should stil have nonIbcTokenDenom tokens after liquid staking only its stTokens (of which there are none)
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal(nonIbctokenAmt.String(), rewardedTokens.AmountOf(nonIbcTokenDenom).String())
}

func (s *KeeperTestSuite) TestLiquidStakeRewardCollectorIbcTokensNoTokens() {
	tc := s.SetupWithdrawAccount()

	// Do not fund reward collector account, should have 0 balance
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal("0", rewardedTokens.AmountOf(tc.hz.IbcDenom).String())

	// Liquid stake all hostzone token should fail
	rewardsFound := s.App.StakeibcKeeper.LiquidStakeRewardCollectorBalance(s.Ctx, s.GetMsgServer())
	s.Require().False(rewardsFound)
}

func (s *KeeperTestSuite) TestSweepRewardCollToFeeCollectorSuccess() {
	tc := s.SetupWithdrawAccount()

	// Fund reward collector account with stTokens
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin("st"+tc.hz.HostDenom, tc.withdrawalAcctBal))

	// Sweep stTokens from Reward Collector to Fee Collector
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Reward Collector acct should have no more stTokens after they're swept
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal("0", rewardLiquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())

	// Fee Collector acct should have stTokens after they're swept there from Reward Collector
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName).GetAddress()
	feeLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollectorAddress)
	s.Require().Equal(tc.withdrawalAcctBal.String(), feeLiquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())
}

func (s *KeeperTestSuite) TestSweepRewardCollToFeeCollectorNonStTokens() {
	tc := s.SetupWithdrawAccount()

	// Fund reward collector account with stTokens
	nonStTokenDenom := "XXX"
	nonStTokenAmt := sdk.NewInt(10)
	s.FundModuleAccount(stakeibctypes.RewardCollectorName, sdk.NewCoin(nonStTokenDenom, nonStTokenAmt))

	// Sweep stTokens from Reward Collector to Fee Collector
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Reward Collector acct should still contain nonStTokenDenom after stTokens after they're swept
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal(nonStTokenAmt.String(), rewardLiquidStakedTokens.AmountOf(nonStTokenDenom).String())

	// Fee Collector acct should have no stTokens or nonStTokenDenom tokens
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName).GetAddress()
	feeLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollectorAddress)
	s.Require().Equal("0", feeLiquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())
	s.Require().Equal("0", feeLiquidStakedTokens.AmountOf(nonStTokenDenom).String())
}

func (s *KeeperTestSuite) TestSweepRewardCollToFeeCollectorNoTokens() {
	tc := s.SetupWithdrawAccount()

	// Do not fund the reward collector account, check it has no stTokens
	rewardCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName).GetAddress()
	rewardLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, rewardCollectorAddress)
	s.Require().Equal("0", rewardLiquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())

	// Sweep stTokens from Reward Collector to Fee Collector, should not error
	err := s.App.StakeibcKeeper.SweepStTokensFromRewardCollToFeeColl(s.Ctx)
	s.Require().NoError(err)

	// Fee Collector acct should have no stTokens tokens
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName).GetAddress()
	feeLiquidStakedTokens := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollectorAddress)
	s.Require().Equal("0", feeLiquidStakedTokens.AmountOf("st"+tc.hz.HostDenom).String())
}

// Test the process of a delegator claiming staking reward stTokens (tests that Fee Account can distribute arbitrary denoms)
func (s *KeeperTestSuite) TestClaimStakingRewardStTokens() {
	tc := s.SetupWithdrawAccount()

	// Fund fee collector account with stTokens
	s.FundModuleAccount(authtypes.FeeCollectorName, sdk.NewCoin("st"+tc.hz.HostDenom, tc.withdrawalAcctBal))

	rewards, err := SimulateProcessingDelegationRewards(s)
	s.Require().NoError(err)

	// Check balances contains stTokens
	s.Require().True(strings.Contains(rewards.String(), "stuatom"))

}

// Helper: setup staking accounts, advance Stride 1 block and distribute rewards to stakers
func SimulateProcessingDelegationRewards(s *KeeperTestSuite) (sdk.Coins, error) {

	// Set up validators & delegators on Stride
	addrs := s.TestAccs
	for _, acc := range addrs {
		s.FundAccount(acc, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	}
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(s.T(), s.Ctx, s.App.StakingKeeper)

	PK := simapp.CreateTestPubKeys(2)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], PK[0], sdk.NewInt(100), true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], PK[1], sdk.NewInt(100), true)

	s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{})
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)

	// Simulate the token distribution from feeCollector to validators
	abciValA := abci.Validator{
		Address: PK[0].Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: PK[1].Address(),
		Power:   100,
	}
	votes := []abci.VoteInfo{
		{
			Validator:       abciValA,
			SignedLastBlock: true,
		},
		{
			Validator:       abciValB,
			SignedLastBlock: true,
		},
	}
	s.App.DistrKeeper.AllocateTokens(s.Ctx, 200, 200, sdk.ConsAddress(PK[1].Address()), votes)

	// Withdraw reward
	return s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[1]), valAddrs[1])
}
