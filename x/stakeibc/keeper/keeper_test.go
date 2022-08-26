package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/Stride-Labs/stride/app/apptesting"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

type KeeperTestSuite struct {
	apptesting.AppTestHelper
	msgServer types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.msgServer = keeper.NewMsgServerImpl(s.App.StakeibcKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestTransfer() {
	s.CreateTransferChannel("GAIA")
	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	coins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000)))
	err := s.App.BankKeeper.MintCoins(s.Ctx(), minttypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx(), minttypes.ModuleName, addr1, coins)
	s.Require().NoError(err)

	fmt.Println(addr1.String())
	fmt.Printf("%v\n", s.App.BankKeeper.GetAllBalances(s.Ctx(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.HostChain.GetSimApp().BankKeeper.GetAllBalances(s.HostCtx(), addr2))

	s.App.TransferKeeper.SendTransfer(
		s.Ctx(),
		"transfer",
		"channel-0",
		sdk.NewCoin("uatom", sdk.NewInt(500)),
		addr1,
		addr2.String(),
		clienttypes.NewHeight(0, 300),
		0,
	)

	fmt.Println(addr1.String())
	fmt.Printf("%v\n", s.App.BankKeeper.GetAllBalances(s.Ctx(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.HostChain.GetSimApp().BankKeeper.GetAllBalances(s.HostCtx(), addr2))
}

func (s *KeeperTestSuite) TestCreateChannels() {
	s.CreateICAChannel("GAIA.DELEGATION")
	s.CreateICAChannel("GAIA.FEE")
	s.CreateICAChannel("GAIA.REDEMPTION")
	s.CreateICAChannel("GAIA.WITHDRAWAL")

	channels := s.App.IBCKeeper.ChannelKeeper.GetAllChannels(s.Ctx())
	for _, channel := range channels {
		fmt.Printf("%v\n", channel)
	}
}

func (s *KeeperTestSuite) TestIca() {
	s.CreateICAChannel("GAIA.DELEGATION")
	s.CreateICAChannel("GAIA.FEE")
	s.CreateICAChannel("GAIA.REDEMPTION")
	s.CreateICAChannel("GAIA.WITHDRAWAL")

	delegationAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx(), "connection-0", "icacontroller-GAIA.DELEGATION")
	s.Require().True(found)
	fmt.Println("DELEGATION ADDRESS:", delegationAddress)

	feeAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx(), "connection-0", "icacontroller-GAIA.FEE")
	s.Require().True(found)
	fmt.Println("FEE ADDRESS:", feeAddress)

	redemptionAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx(), "connection-0", "icacontroller-GAIA.REDEMPTION")
	s.Require().True(found)
	fmt.Println("REDEMPTION ADDRESS:", redemptionAddress)

	withdrawAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx(), "connection-0", "icacontroller-GAIA.WITHDRAWAL")
	s.Require().True(found)
	fmt.Println("WITHDRAWAL ADDRESS:", withdrawAddress)

	timeoutTimestamp := uint64(s.Ctx().BlockTime().UnixNano() + 100_000_000_000)

	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{
		FromAddress: delegationAddress,
		ToAddress:   feeAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000))),
	})

	data, err := icatypes.SerializeCosmosTx(s.StrideChain.Codec, msgs)
	s.Require().NoError(err)

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	chanCap, found := s.App.ScopedIBCKeeper.GetCapability(s.Ctx(), host.ChannelCapabilityPath("icacontroller-GAIA.DELEGATION", "channel-1"))
	s.Require().True(found)

	seq, err := s.App.ICAControllerKeeper.SendTx(s.Ctx(), chanCap, "connection-0", "icacontroller-GAIA.DELEGATION", packetData, timeoutTimestamp)
	s.Require().NoError(err)
	fmt.Println("SEQUENCE:", seq)
}
