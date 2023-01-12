package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type ClearBalanceState struct {
	feeChannel Channel
	hz         stakeibctypes.HostZone
}

type ClearBalanceTestCase struct {
	initialState ClearBalanceState
	validMsg     stakeibctypes.MsgClearBalance
}

func (s *KeeperTestSuite) SetupClearBalance() ClearBalanceTestCase {
	// fee account
	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	feeChannelID := s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]
	// hz
	zoneAddress := types.NewZoneAddress(HostChainId)
	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		ConnectionId:   ibctesting.FirstConnectionID,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Address:        zoneAddress.String(),
		FeeAccount: &stakeibctypes.ICAAccount{
			Address: feeAddress,
			Target:  stakeibctypes.ICAAccountType_FEE,
		},
	}

	amount := sdk.NewInt(1_000_000)

	user := Account{
		acc: s.TestAccs[0],
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return ClearBalanceTestCase{
		initialState: ClearBalanceState{
			hz: hostZone,
			feeChannel: Channel{
				PortID:    icatypes.PortPrefix + feeAccountOwner,
				ChannelID: feeChannelID,
			},
		},
		validMsg: stakeibctypes.MsgClearBalance{
			Creator: user.acc.String(),
			ChainId: HostChainId,
			Amount:  amount,
			Channel: feeChannelID,
		},
	}
}

func (s *KeeperTestSuite) TestClearBalance_Successful() {
	tc := s.SetupClearBalance()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	feeChannel := tc.initialState.feeChannel
	feePortId := feeChannel.PortID
	feeChannelId := feeChannel.ChannelID

	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found before clear balance")

	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().NoError(err, "balance clears")

	// Confirm the sequence number was incremented
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found after clear balance")
	s.Require().Equal(endSequence, startSequence+1, "sequence number after clear balance")
}

func (s *KeeperTestSuite) TestClearBalance_HostChainMissing() {
	tc := s.SetupClearBalance()
	// remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "chainId: GAIA: host zone not registered")
}

func (s *KeeperTestSuite) TestClearBalance_FeeAccountMissing() {
	tc := s.SetupClearBalance()
	// no fee account
	tc.initialState.hz.FeeAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.initialState.hz)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "chainId: GAIA: fee account is not registered")
}

func (s *KeeperTestSuite) TestClearBalance_ParseCoinError() {
	tc := s.SetupClearBalance()
	// invalid denom
	tc.initialState.hz.HostDenom = ":"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.initialState.hz)
	_, err := s.GetMsgServer().ClearBalance(sdk.WrapSDKContext(s.Ctx), &tc.validMsg)
	s.Require().EqualError(err, "failed to parse coin (1000000:): invalid decimal coin expression: 1000000:")
}
