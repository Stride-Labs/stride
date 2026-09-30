// x/stakeibc/keeper/wind_down_transfer_from_ica_test.go
package keeper_test

import (
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	testOsmosisVault = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"
	hubChainId       = "cosmoshub-4" // must be a key of HostToOsmosisTransferChannel
)

// Sets the vault for the test and restores the empty default afterwards
func (s *KeeperTestSuite) withOsmosisVault() {
	previous := types.OsmosisVaultAddress
	types.OsmosisVaultAddress = testOsmosisVault
	s.T().Cleanup(func() { types.OsmosisVaultAddress = previous })
}

func (s *KeeperTestSuite) hubHostZone() types.HostZone {
	return types.HostZone{
		ChainId:              hubChainId,
		ConnectionId:         ibctesting.FirstConnectionID,
		HostDenom:            Atom,
		DelegationIcaAddress: "cosmos_DELEGATION",
		WithdrawalIcaAddress: "cosmos_WITHDRAWAL",
		FeeIcaAddress:        "cosmos_FEE",
		RedemptionIcaAddress: "cosmos_REDEMPTION",
	}
}

var transferableIcaTypes = []types.ICAAccountType{
	types.ICAAccountType_DELEGATION,
	types.ICAAccountType_WITHDRAWAL,
	types.ICAAccountType_FEE,
	types.ICAAccountType_REDEMPTION,
}

// Channels first, host zone second: the first CreateICAChannel call creates the transfer
// channel, which replaces s.App and s.Ctx with the ibctesting chain's
// (app/apptesting/test_helpers.go:396-398), so anything written to the store before it is lost
func (s *KeeperTestSuite) setupHubIcaChannels(icaTypes ...types.ICAAccountType) map[types.ICAAccountType][2]string {
	channels := map[types.ICAAccountType][2]string{}
	for _, icaType := range icaTypes {
		channelId, portId := s.CreateICAChannel(types.FormatHostZoneICAOwner(hubChainId, icaType))
		channels[icaType] = [2]string{portId, channelId}
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, s.hubHostZone())
	return channels
}

func (s *KeeperTestSuite) TestTransferFromIca_EveryIcaType() {
	s.withOsmosisVault()
	channels := s.setupHubIcaChannels(transferableIcaTypes...)

	for _, icaType := range transferableIcaTypes {
		portId, channelId := channels[icaType][0], channels[icaType][1]
		msg := types.NewMsgTransferFromIca("admin", hubChainId, icaType, sdk.NewCoin(Atom, sdkmath.NewInt(1000)))
		s.CheckICATxSubmitted(portId, channelId, func() error {
			return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
		})
		s.CheckEventValueEmitted(types.EventTypeTransferFromIca, types.AttributeKeyIcaType, icaType.String())
	}
}

func (s *KeeperTestSuite) TestTransferFromIca_ForeignDenom() {
	s.withOsmosisVault()
	channels := s.setupHubIcaChannels(types.ICAAccountType_WITHDRAWAL)
	portId, channelId := channels[types.ICAAccountType_WITHDRAWAL][0], channels[types.ICAAccountType_WITHDRAWAL][1]

	usdcOnHub := "ibc/F663521BF1836B00F5F177680F74BFB9A8B5654A694D0D2BC249E03CF2509013"
	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_WITHDRAWAL, sdk.NewCoin(usdcOnHub, sdkmath.NewInt(3_800_000)))
	s.CheckICATxSubmitted(portId, channelId, func() error {
		return s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	})
}

func (s *KeeperTestSuite) TestTransferFromIca_VaultNotConfigured() {
	// The default is empty: the tx must fail closed
	s.Require().Empty(types.OsmosisVaultAddress)
	channels := s.setupHubIcaChannels(types.ICAAccountType_DELEGATION)
	portId, channelId := channels[types.ICAAccountType_DELEGATION][0], channels[types.ICAAccountType_DELEGATION][1]

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	startSequence := s.MustGetNextSequenceNumber(portId, channelId)
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrOsmosisVaultNotConfigured)
	s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(portId, channelId), "nothing submitted")
}

func (s *KeeperTestSuite) TestTransferFromIca_ChainIdNotInMap() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	hostZone.ChainId = "comdex-1"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	msg := types.NewMsgTransferFromIca("admin", "comdex-1", types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrNoOsmosisChannelForHostZone)
}

func (s *KeeperTestSuite) TestTransferFromIca_MissingZoneAndIca() {
	s.withOsmosisVault()

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)

	hostZone := s.hubHostZone()
	hostZone.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	msg = types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_FEE, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	err = s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
	s.Require().ErrorIs(err, types.ErrICAAccountNotFound)
}

// The built ICA message: mapped channel, vault receiver, the given timeout, empty memo
func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_Transfer() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	amount := sdk.NewCoin(Atom, sdkmath.NewInt(1000))
	blockTime := s.Ctx.BlockTime()
	// The inner transfer gets twice the ICA packet's window: a late-relayed ICA packet still
	// leaves the host to Osmosis leg a full day
	timeout := uint64(blockTime.Add(2 * types.WindDownTransferTimeout).UnixNano())

	built, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_WITHDRAWAL, amount, blockTime)
	s.Require().NoError(err)
	transfer, ok := built.(*transfertypes.MsgTransfer)
	s.Require().True(ok, "non-osmosis zones build an ICS-20 transfer")
	s.Require().Equal(transfertypes.PortID, transfer.SourcePort)
	s.Require().Equal("channel-141", transfer.SourceChannel, "cosmoshub-4's channel to osmosis")
	s.Require().Equal(amount, transfer.Token)
	s.Require().Equal("cosmos_WITHDRAWAL", transfer.Sender)
	s.Require().Equal(testOsmosisVault, transfer.Receiver)
	s.Require().Equal(timeout, transfer.TimeoutTimestamp)
	s.Require().Zero(transfer.TimeoutHeight.RevisionHeight)
	s.Require().Empty(transfer.Memo)
}

// osmosis-1 maps to an empty channel: the ICA is already on Osmosis so it is a bank send
func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_OsmosisBankSend() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	hostZone.ChainId = types.OsmosisChainId
	hostZone.DelegationIcaAddress = "osmo_DELEGATION"
	amount := sdk.NewCoin(Osmo, sdkmath.NewInt(500))

	built, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_DELEGATION, amount, s.Ctx.BlockTime())
	s.Require().NoError(err)
	send, ok := built.(*banktypes.MsgSend)
	s.Require().True(ok, "osmosis-1 builds a bank send")
	s.Require().Equal("osmo_DELEGATION", send.FromAddress)
	s.Require().Equal(testOsmosisVault, send.ToAddress)
	s.Require().Equal(sdk.NewCoins(amount), send.Amount)
}

func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_Rejections() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()

	_, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_COMMUNITY_POOL_DEPOSIT, sdk.NewCoin(Atom, sdkmath.NewInt(1)), s.Ctx.BlockTime())
	s.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest, "only the four funded ICAs")

	hostZone.ChainId = "evmos_9001-2"
	_, err = keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)), s.Ctx.BlockTime())
	s.Require().ErrorIs(err, types.ErrNoOsmosisChannelForHostZone)
}

func (s *KeeperTestSuite) TestMsgServer_TransferFromIca() {
	s.withOsmosisVault()
	channels := s.setupHubIcaChannels(types.ICAAccountType_DELEGATION)
	portId, channelId := channels[types.ICAAccountType_DELEGATION][0], channels[types.ICAAccountType_DELEGATION][1]

	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))
	s.CheckICATxSubmitted(portId, channelId, func() error {
		_, err := s.GetMsgServer().TransferFromIca(s.Ctx, msg)
		return err
	})
}

// A non-osmosis zone whose mapped channel is empty must not silently become a bank send
func (s *KeeperTestSuite) TestBuildTransferFromIcaMsg_EmptyChannelOnlyForOsmosis() {
	s.withOsmosisVault()
	hostZone := s.hubHostZone()
	previous := types.HostToOsmosisTransferChannel[hubChainId]
	types.HostToOsmosisTransferChannel[hubChainId] = ""
	s.T().Cleanup(func() { types.HostToOsmosisTransferChannel[hubChainId] = previous })

	_, err := keeper.BuildTransferFromIcaMsg(hostZone, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)), s.Ctx.BlockTime())
	s.Require().ErrorIs(err, types.ErrNoOsmosisChannelForHostZone)
}

func (s *KeeperTestSuite) TestTransferFromIca_RejectsInvalidVault() {
	channels := s.setupHubIcaChannels(types.ICAAccountType_DELEGATION)
	portId, channelId := channels[types.ICAAccountType_DELEGATION][0], channels[types.ICAAccountType_DELEGATION][1]
	msg := types.NewMsgTransferFromIca("admin", hubChainId, types.ICAAccountType_DELEGATION, sdk.NewCoin(Atom, sdkmath.NewInt(1)))

	previous := types.OsmosisVaultAddress
	s.T().Cleanup(func() { types.OsmosisVaultAddress = previous })

	for _, vault := range []string{
		"osmo1notabech32address",
		"stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9a8n6xp0", // wrong prefix
	} {
		types.OsmosisVaultAddress = vault
		startSequence := s.MustGetNextSequenceNumber(portId, channelId)
		err := s.App.StakeibcKeeper.TransferFromIca(s.Ctx, msg)
		s.Require().ErrorIs(err, types.ErrOsmosisVaultNotConfigured, "vault %q", vault)
		s.Require().Equal(startSequence, s.MustGetNextSequenceNumber(portId, channelId), "nothing submitted")
	}
}
