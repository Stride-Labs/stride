package keeper_test

import (
	"fmt"
	"testing"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/Stride-Labs/stride/app/apptesting"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

var (
	TestOwnerAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"
	TestVersion      = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
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

type MultiChainKeeperTestSuite struct {
	apptesting.AppTestHelper
	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	path   *ibctesting.Path
}

func (s *MultiChainKeeperTestSuite) SetupTestIbc() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))

	s.path = ibctesting.NewPath(s.chainA, s.chainB)
	s.path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	s.path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	s.path.EndpointA.ChannelConfig.Version = transfertypes.Version
	s.path.EndpointB.ChannelConfig.Version = transfertypes.Version

	s.coordinator.Setup(s.path)
	s.Require().Equal("07-tendermint-0", s.path.EndpointA.ClientID)
	s.Require().Equal("connection-0", s.path.EndpointA.ConnectionID)
	s.Require().Equal("channel-0", s.path.EndpointA.ChannelID)
}

func (s *MultiChainKeeperTestSuite) CreateAccounts() (sdk.AccAddress, sdk.AccAddress) {
	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	coins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000)))
	err := s.chainA.GetSimApp().BankKeeper.MintCoins(s.chainA.GetContext(), minttypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.chainA.GetSimApp().BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), minttypes.ModuleName, addr1, coins)
	s.Require().NoError(err)

	return addr1, addr2
}

func (s *MultiChainKeeperTestSuite) TestIbc() {
	s.SetupTestIbc()

	addr1, addr2 := s.CreateAccounts()
	fmt.Println(addr1.String())
	fmt.Printf("%v\n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.chainB.GetSimApp().BankKeeper.GetAllBalances(s.chainB.GetContext(), addr2))

	s.chainA.GetSimApp().TransferKeeper.SendTransfer(
		s.chainA.GetContext(),
		"transfer",
		"channel-0",
		sdk.NewCoin("uatom", sdk.NewInt(500)),
		addr1,
		addr2.String(),
		clienttypes.NewHeight(0, 300),
		0,
	)

	fmt.Println(addr1.String())
	fmt.Printf("%v\n", s.chainA.GetSimApp().BankKeeper.GetAllBalances(s.chainA.GetContext(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.chainB.GetSimApp().BankKeeper.GetAllBalances(s.chainB.GetContext(), addr2))
}

// RegisterInterchainAccount is a helper function for starting the channel handshake
func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	if err := endpoint.Chain.GetSimApp().ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, owner); err != nil {
		return err
	}

	// commit state changes for proof verification
	endpoint.Chain.App.Commit()
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID

	return nil
}

func (s *MultiChainKeeperTestSuite) TestSetupIca() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))

	// transferPath := ibctesting.NewPath(s.chainA, s.chainB)
	// transferPath.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	// transferPath.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	// transferPath.EndpointA.ChannelConfig.Version = transfertypes.Version
	// transferPath.EndpointB.ChannelConfig.Version = transfertypes.Version

	// s.coordinator.Setup(transferPath)
	// s.Require().Equal("07-tendermint-0", transferPath.EndpointA.ClientID)
	// s.Require().Equal("connection-0", transferPath.EndpointA.ConnectionID)
	// s.Require().Equal("channel-0", transferPath.EndpointA.ChannelID)

	path := ibctesting.NewPath(s.chainA, s.chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.PortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	s.coordinator.SetupConnections(path)

	// err := s.chainA.GetSimApp().ICAControllerKeeper.RegisterInterchainAccount(s.chainA.GetContext(), "connection-0", "GAIA.DELEGATION")
	err := RegisterInterchainAccount(path.EndpointA, "GAIA.DELEGATION")
	s.Require().NoError(err)

	channels := s.chainA.GetSimApp().IBCKeeper.ChannelKeeper.GetAllChannels(s.chainA.GetContext())
	for _, channel := range channels {
		fmt.Printf("%v\n", channel)
	}

	err = path.EndpointB.ChanOpenTry()
	s.Require().NoError(err)

	err = path.EndpointA.ChanOpenAck()
	s.Require().NoError(err)

	err = path.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err)

	channels = s.chainA.GetSimApp().IBCKeeper.ChannelKeeper.GetAllChannels(s.chainA.GetContext())
	for _, channel := range channels {
		fmt.Printf("%v\n", channel)
	}

	icaAddress, found := s.chainA.GetSimApp().ICAControllerKeeper.GetInterchainAccountAddress(s.chainA.GetContext(), "connection-0", "icacontroller-GAIA.DELEGATION")
	s.Require().True(found)

	fmt.Println(icaAddress)

	timeoutTimestamp := uint64(s.chainA.GetContext().BlockTime().UnixNano() + 100_000_000_000)

	_, addr2 := s.CreateAccounts()

	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{
		FromAddress: icaAddress,
		ToAddress:   addr2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000))),
	})

	data, err := icatypes.SerializeCosmosTx(s.chainA.Codec, msgs)
	s.Require().NoError(err)

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}
	chanCap, found := s.chainA.GetSimApp().ScopedIBCKeeper.GetCapability(s.chainA.GetContext(), host.ChannelCapabilityPath("icacontroller-GAIA.DELEGATION", "channel-0"))
	s.Require().True(found)
	s.chainA.GetSimApp().ICAControllerKeeper.SendTx(s.chainA.GetContext(), chanCap, "connection-0", "icacontroller-GAIA.DELEGATION", packetData, timeoutTimestamp)
}

func TestMultiChainKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(MultiChainKeeperTestSuite))
}
