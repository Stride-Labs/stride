package keeper_test

import (
	"fmt"
	"strings"

	_ "github.com/stretchr/testify/suite"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	sdkmath "cosmossdk.io/math"
	epochtypes "github.com/Stride-Labs/stride/v6/x/epochs/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v6/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	hosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"
	recordtypes "github.com/Stride-Labs/stride/v6/x/records/types"
)

var (
	validators = []*stakeibctypes.Validator{
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
	hostModuleAddress = stakeibctypes.NewZoneAddress(HostChainId)
)

func (s *KeeperTestSuite) SetupWithdrawAccount() (stakeibctypes.HostZone, Channel) {
	// Set up host zone ica
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	_ = s.CreateICAChannel(delegationAccountOwner)
	delegationAddress := s.IcaAddresses[delegationAccountOwner]

	withdrawalAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "WITHDRAWAL")
	withdrawalChannelID := s.CreateICAChannel(withdrawalAccountOwner)
	withdrawalAddress := s.IcaAddresses[withdrawalAccountOwner]
		
	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]

	// Set up ibc denom
	ibcDenomTrace := s.GetIBCDenomTrace(Atom) // we need a true IBC denom here
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, ibcDenomTrace)

	// Fund withdraw ica
	initialModuleAccountBalance := sdk.NewCoin(Atom, sdkmath.NewInt(15_000))
	s.FundAccount(sdk.MustAccAddressFromBech32(withdrawalAddress), initialModuleAccountBalance)
	s.HostApp.BankKeeper.MintCoins(s.HostChain.GetContext(), minttypes.ModuleName, sdk.NewCoins(initialModuleAccountBalance))
	s.HostApp.BankKeeper.SendCoinsFromModuleToAccount(s.HostChain.GetContext(), minttypes.ModuleName, sdk.MustAccAddressFromBech32(withdrawalAddress), sdk.NewCoins(initialModuleAccountBalance))

	// Allow ica call ibc transfer in host chain
	s.HostApp.ICAHostKeeper.SetParams(s.HostChain.GetContext(), hosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			"/ibc.applications.transfer.v1.MsgTransfer",
		},
	})

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
	mintEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.MINT_EPOCH,
		EpochNumber:        currentEpoch,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 60_000_000_000), // dictates timeouts
	}

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

	return hostZone, Channel{
		PortID: icatypes.PortPrefix + withdrawalAccountOwner,
		ChannelID: withdrawalChannelID,
	}
}

func (s *KeeperTestSuite) TestAllocateRewardIBC() {
	hz, channel := s.SetupWithdrawAccount()

	rewardCollector := s.App.AccountKeeper.GetModuleAccount(s.Ctx, stakeibctypes.RewardCollectorName)
	
	// Send tx to withdraw ica to perform ibc transfer from hostzone to stride
	var msgs []sdk.Msg
	ibcTransferTimeoutNanos := s.App.StakeibcKeeper.GetParam(s.Ctx, stakeibctypes.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(s.HostChain.GetContext().BlockTime().UnixNano()) + ibcTransferTimeoutNanos
	msg := ibctypes.NewMsgTransfer("transfer", "channel-0", sdk.NewCoin(Atom, sdkmath.NewInt(15_000)), hz.WithdrawalAccount.Address, rewardCollector.GetAddress().String(), clienttypes.NewHeight(1, 100), timeoutTimestamp)
	msgs = append(msgs, msg)
	data, _ := icatypes.SerializeCosmosTx(s.App.AppCodec(), msgs)
	
	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}
	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		1,
		channel.PortID,
		channel.ChannelID,
		s.TransferPath.EndpointB.ChannelConfig.PortID,
		s.TransferPath.EndpointB.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)
	s.App.StakeibcKeeper.SubmitTxs(s.Ctx, hz.ConnectionId, msgs, *hz.WithdrawalAccount, 0, "", nil)

	// Simulate the process of receiving ica packets on the hostchain
	module, _, err := s.HostChain.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(s.HostChain.GetContext(), "icahost")
	s.Require().NoError(err)
	cbs, ok := s.HostChain.App.GetIBCKeeper().Router.GetRoute(module)
	s.Require().True(ok)
	cbs.OnRecvPacket(s.HostChain.GetContext(), packet, nil)

    // After withdraw ica send ibc transfer, simulate receiving transfer packet at stride
	transferPacketData := ibctypes.NewFungibleTokenPacketData(
		Atom, sdkmath.NewInt(15_000).String(), hz.WithdrawalAccount.Address, rewardCollector.GetAddress().String(),
	)
	transferPacketData.Memo = ""
	transferPacket := channeltypes.NewPacket(
		transferPacketData.GetBytes(),
		1,
		s.TransferPath.EndpointB.ChannelConfig.PortID,
		s.TransferPath.EndpointB.ChannelID,
		s.TransferPath.EndpointA.ChannelConfig.PortID,
		s.TransferPath.EndpointA.ChannelID,
		clienttypes.NewHeight(1, 100),
		0,
	)
	cbs, ok = s.StrideChain.App.GetIBCKeeper().Router.GetRoute("transfer")
	s.Require().True(ok)
	cbs.OnRecvPacket(s.StrideChain.GetContext(), transferPacket, nil)

	// Liquid stake all hostzone token then get stTokens back
	// s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})
	err = s.App.StakeibcKeeper.AllocateHostZoneReward(s.Ctx)

	// Set up validator & delegation
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
			Validator: abciValA,
			SignedLastBlock: true,
		},
		{
			Validator: abciValB,
			SignedLastBlock: true,
		},
	}
	s.App.DistrKeeper.AllocateTokens(s.Ctx, 200, 200, sdk.ConsAddress(PK[1].Address()), votes)

	// Withdraw reward
	s.App.DistrKeeper.WithdrawDelegationRewards(s.Ctx, sdk.AccAddress(valAddrs[1]), valAddrs[1])

	// Check balances contains stTokens
	rewards := s.App.BankKeeper.GetAllBalances(s.Ctx, sdk.AccAddress(valAddrs[1]))
	s.Require().True(strings.Contains(rewards.String(), "stuatom"))

}
