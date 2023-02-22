package keeper_test

import (
	"fmt"
	// "time"

	// "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"
	_ "github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"

	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"

	sdkmath "cosmossdk.io/math"

	epochtypes "github.com/Stride-Labs/stride/v5/x/epochs/types"
	// icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/v5/x/interchainquery/types"

	// icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
	// recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	// abci "github.com/tendermint/tendermint/abci/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	// stakeibckeeper "github.com/Stride-Labs/stride/v5/x/stakeibc/keeper"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	// icaapp "github.com/cosmos/interchain-accounts/app"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	// ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

func (s *KeeperTestSuite) SetupWithdrawAccount() (WithdrawalBalanceICQCallbackArgs, Channel, stakeibctypes.HostZone, channeltypes.Packet) {
	
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	_ = s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	withdrawalChannelID := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]
	interchainAccount := s.HostApp.AccountKeeper.GetAccount(s.HostChain.GetContext(), sdk.MustAccAddressFromBech32(withdrawalAddress))
	fmt.Println("interchainAccount", interchainAccount)
		
	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]

	s.HostApp.AccountKeeper.NewAccountWithAddress(s.HostChain.GetContext(), sdk.MustAccAddressFromBech32(withdrawalAddress))

	ibcDenomTrace := s.GetIBCDenomTrace(Atom) // we need a true IBC denom here
	hostModuleAddress := stakeibctypes.NewZoneAddress(HostChainId)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, ibcDenomTrace)

	initialModuleAccountBalance := sdk.NewCoin(ibcDenomTrace.IBCDenom(), sdkmath.NewInt(15_000))
	s.FundAccount(sdk.MustAccAddressFromBech32(withdrawalAddress), initialModuleAccountBalance)
	s.HostApp.BankKeeper.MintCoins(s.HostChain.GetContext(), minttypes.ModuleName, sdk.NewCoins(initialModuleAccountBalance))
	fmt.Println("module host", s.HostApp.AccountKeeper.GetModuleAddress(minttypes.ModuleName).String(), s.HostApp.BankKeeper.GetAllBalances(s.HostChain.GetContext(), s.HostApp.AccountKeeper.GetModuleAddress(minttypes.ModuleName)))
	s.HostApp.BankKeeper.SendCoinsFromModuleToAccount(s.HostChain.GetContext(), minttypes.ModuleName, sdk.MustAccAddressFromBech32(withdrawalAddress), sdk.NewCoins(initialModuleAccountBalance))
	fmt.Println("withdrawalAddress", s.HostApp.BankKeeper.GetAllBalances(s.HostChain.GetContext(), sdk.MustAccAddressFromBech32(withdrawalAddress)))

	validators := []*stakeibctypes.Validator{
		{
			Name:    "val1",
			Address: "gaia_VAL1",
			Weight:  1,
		},
		{
			Name:    "val2",
			Address: "gaia_VAL2",
			Weight:  4,
		},
	}

	hostZone := stakeibctypes.HostZone{
		ChainId:           HostChainId,
		Address:           hostModuleAddress.String(),
		DelegationAccount: &stakeibctypes.ICAAccount{Address: delegationAddress},
		WithdrawalAccount: &stakeibctypes.ICAAccount{
			Address: withdrawalAddress,
			Target:  stakeibctypes.ICAAccountType_WITHDRAWAL,
		},
		FeeAccount: &stakeibctypes.ICAAccount{
			Address: feeAddress,
			Target: stakeibctypes.ICAAccountType_FEE,
		},
		ConnectionId:      ibctesting.FirstConnectionID,
		TransferChannelId: ibctesting.FirstChannelID,
		HostDenom:         Atom,
		IbcDenom:          ibcDenomTrace.IBCDenom(),
		Validators:        validators,
		RedemptionRate: sdk.OneDec(),
	}

	currentEpoch := uint64(2)
	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	queryResponse := s.CreateBalanceQueryResponse(initialModuleAccountBalance.Amount.Int64(), ibcDenomTrace.IBCDenom())

	validArgs := WithdrawalBalanceICQCallbackArgs{
		query: icqtypes.Query{
			Id:      "0",
			ChainId: HostChainId,
		},
		callbackArgs: queryResponse,
	}

	feeCollector := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName)
	// Call the ICQ callback
	var msgs []sdk.Msg
	ibcTransferTimeoutNanos := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(s.HostChain.GetContext().BlockTime().UnixNano()) + ibcTransferTimeoutNanos
	msg := ibctypes.NewMsgTransfer("transfer", "channel-0", initialModuleAccountBalance, withdrawalAddress, feeCollector.GetAddress().String(), clienttypes.NewHeight(1, 100), timeoutTimestamp)
	msgs = append(msgs, msg)
	data, _ := icatypes.SerializeCosmosTx(s.App.AppCodec(), msgs)
	

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}
	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		icatypes.PortPrefix + withdrawalAccountOwner,
		withdrawalChannelID,
		s.TransferPath.EndpointB.ChannelConfig.PortID,
		s.TransferPath.EndpointB.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)

	s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hostZone.ConnectionId, msgs, *hostZone.WithdrawalAccount, 0, "", nil)

	return validArgs, Channel{
		PortID: icatypes.PortPrefix + withdrawalAccountOwner,
		ChannelID: withdrawalChannelID,
	}, hostZone, packet
}

func (s *KeeperTestSuite) TestAllocateRewardIBC() {
	_, channel, _, packet := s.SetupWithdrawAccount()
	chanels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx)
			fmt.Println("channel", chanels)
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, channel.PortID, channel.ChannelID)
	s.Require().True(found, "sequence number not found before reinvestment")
	fmt.Println("startSequence", startSequence)
	feeCollector := s.App.AccountKeeper.GetModuleAccount(s.Ctx, authtypes.FeeCollectorName).GetAddress()
	oldBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollector)
	fmt.Println("oldBalances", oldBalances)

	module, _, err := s.HostChain.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(s.HostChain.GetContext(), "icahost")
	fmt.Println("module", module, err)
	s.Require().NoError(err)

	cbs, ok := s.HostChain.App.GetIBCKeeper().Router.GetRoute(module)
	s.Require().True(ok)

	ack := cbs.OnRecvPacket(s.HostChain.GetContext(), packet, nil)
	fmt.Println("ack", ack.Acknowledgement(), ack.Success())

	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, channel.PortID, channel.ChannelID)
	s.Require().True(found, "sequence number not found before reinvestment")
	fmt.Println("endSequence", endSequence)
	newBalances := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollector)
	fmt.Println("newBalances", newBalances)

}

func (s *KeeperTestSuite) TestAllocateReward() {
	s.Setup()
	app := s.App	
	addrs := s.TestAccs
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	tstaking := teststaking.NewHelper(s.T(), s.Ctx, app.StakingKeeper)

	PK := simapp.CreateTestPubKeys(2)

	// create validator with 50% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[0], PK[0], sdk.NewInt(100), true)

	// create second validator with 0% commission
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDec(0), sdk.NewDec(0), sdk.NewDec(0))
	tstaking.CreateValidator(valAddrs[1], PK[1], sdk.NewInt(100), true)

	abciValA := abci.Validator{
		Address: PK[0].Address(),
		Power:   100,
	}
	abciValB := abci.Validator{
		Address: PK[1].Address(),
		Power:   100,
	}

	// allocate tokens as if both had voted and second was proposer
	// fund fee collector
	s.FundModuleAccount("fee_collector", sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)))
	s.FundModuleAccount("fee_collector", sdk.NewCoin("atom", sdk.NewInt(100)))


	// end block to bond validator
	staking.EndBlocker(s.Ctx, app.StakingKeeper)
	// next block
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)

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
	app.DistrKeeper.AllocateTokens(s.Ctx, 200, 200, sdk.ConsAddress(PK[1].Address()), votes)

	// 98 outstanding rewards (100 less 2 to community pool)
	// Previous proposer got 100 * (5% + 93% / 2)
	require.Equal(s.T(), sdk.DecCoins{{Denom: "atom", Amount: sdk.NewDecWithPrec(515, 1)}, {Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(515, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(s.Ctx, valAddrs[1]).Rewards)
	// 100 * (93% / 2)
	require.Equal(s.T(), sdk.DecCoins{{Denom: "atom", Amount: sdk.NewDecWithPrec(465, 1)}, {Denom: sdk.DefaultBondDenom, Amount: sdk.NewDecWithPrec(465, 1)}}, app.DistrKeeper.GetValidatorOutstandingRewards(s.Ctx, valAddrs[0]).Rewards)
	
	// Withdraw reward
	balancesBefore := app.BankKeeper.GetAllBalances(s.Ctx, sdk.AccAddress(valAddrs[1]))
	app.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[1]), valAddrs[1])
	balancesAfter := app.BankKeeper.GetAllBalances(s.Ctx, sdk.AccAddress(valAddrs[1]))
	require.Equal(s.T(), sdk.NewCoins(sdk.NewCoin("atom", sdk.NewInt(51)), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(51))), balancesAfter.Sub(balancesBefore...))
}
