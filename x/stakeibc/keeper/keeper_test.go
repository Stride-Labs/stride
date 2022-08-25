package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tendermint/tendermint/libs/log"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"
	dbm "github.com/tendermint/tm-db"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/ibc-go/v3/testing/simapp"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	"github.com/Stride-Labs/stride/app"
	"github.com/Stride-Labs/stride/app/apptesting"
	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

var (
	TestIcaVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
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

func SetupStrideTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	testingApp := app.NewStrideApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		5,
		app.MakeEncodingConfig(),
		simapp.EmptyAppOptions{},
	)
	return testingApp, app.NewDefaultGenesisState()
}

type MultiChainKeeperTestSuite struct {
	apptesting.AppTestHelper
	coordinator *ibctesting.Coordinator

	chainA       *ibctesting.TestChain
	chainB       *ibctesting.TestChain
	transferPath *ibctesting.Path
}

func NewTransferPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version
	return path
}

func NewIcaPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.PortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestIcaVersion
	path.EndpointB.ChannelConfig.Version = TestIcaVersion
	return path
}

func CopyConnectionAndClientToPath(path *ibctesting.Path, pathToCopy *ibctesting.Path) *ibctesting.Path {
	path.EndpointA.ClientID = pathToCopy.EndpointA.ClientID
	path.EndpointB.ClientID = pathToCopy.EndpointB.ClientID
	path.EndpointA.ConnectionID = pathToCopy.EndpointA.ConnectionID
	path.EndpointB.ConnectionID = pathToCopy.EndpointB.ConnectionID
	path.EndpointA.ClientConfig = pathToCopy.EndpointA.ClientConfig
	path.EndpointB.ClientConfig = pathToCopy.EndpointB.ClientConfig
	path.EndpointA.ConnectionConfig = pathToCopy.EndpointA.ConnectionConfig
	path.EndpointB.ConnectionConfig = pathToCopy.EndpointB.ConnectionConfig
	return path
}

func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	if err := endpoint.Chain.App.(*app.StrideApp).ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, owner); err != nil {
		return err
	}

	endpoint.Chain.App.Commit()
	endpoint.Chain.NextBlock()

	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID

	return nil
}

func (s *MultiChainKeeperTestSuite) CreateICAChannel(owner string) error {
	// Create ICA Path and the copy over the client and connection from the transfer path
	path := NewIcaPath(s.chainA, s.chainB)
	path = CopyConnectionAndClientToPath(path, s.transferPath)

	// Register the ICA
	if err := RegisterInterchainAccount(path.EndpointA, owner); err != nil {
		return err
	}

	// Complete the handshake
	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}
	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}
	if err := path.EndpointB.ChanOpenConfirm(); err != nil {
		return err
	}

	return nil
}

func (s *MultiChainKeeperTestSuite) SetupIbc() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 0)
	ibctesting.DefaultTestingAppInit = SetupStrideTestingApp
	s.chainA = ibctesting.NewTestChain(s.T(), s.coordinator, "STRIDE")

	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp
	s.chainB = ibctesting.NewTestChain(s.T(), s.coordinator, "GAIA")

	s.coordinator.Chains = map[string]*ibctesting.TestChain{
		"STRIDE": s.chainA,
		"GAIA":   s.chainB,
	}
	s.transferPath = NewTransferPath(s.chainA, s.chainB)
	s.coordinator.Setup(s.transferPath)

	s.Require().Equal("07-tendermint-0", s.transferPath.EndpointA.ClientID)
	s.Require().Equal("connection-0", s.transferPath.EndpointA.ConnectionID)
	s.Require().Equal("channel-0", s.transferPath.EndpointA.ChannelID)

	s.Require().Equal("07-tendermint-0", s.transferPath.EndpointB.ClientID)
	s.Require().Equal("connection-0", s.transferPath.EndpointB.ConnectionID)
	s.Require().Equal("channel-0", s.transferPath.EndpointB.ChannelID)

	err := s.CreateICAChannel("GAIA.DELEGATION")
	s.Require().NoError(err, "create delegation ICA")

	err = s.CreateICAChannel("GAIA.FEE")
	s.Require().NoError(err, "create fee ICA")

	err = s.CreateICAChannel("GAIA.REDEMPTION")
	s.Require().NoError(err, "create redemption ICA")

	err = s.CreateICAChannel("GAIA.WITHDRAWAL")
	s.Require().NoError(err, "create withdrawal ICA")

	channels := s.chainA.App.(*app.StrideApp).IBCKeeper.ChannelKeeper.GetAllChannels(s.chainA.GetContext())
	for _, channel := range channels {
		fmt.Printf("%v\n", channel)
	}
}

func (s *MultiChainKeeperTestSuite) SetupAccounts() (sdk.AccAddress, sdk.AccAddress) {
	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	coins := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000)))
	err := s.chainA.App.(*app.StrideApp).BankKeeper.MintCoins(s.chainA.GetContext(), minttypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.chainA.App.(*app.StrideApp).BankKeeper.SendCoinsFromModuleToAccount(s.chainA.GetContext(), minttypes.ModuleName, addr1, coins)
	s.Require().NoError(err)

	return addr1, addr2
}

func (s *MultiChainKeeperTestSuite) TestTransfer() {
	s.SetupIbc()

	addr1, addr2 := s.SetupAccounts()
	fmt.Println(addr1.String())
	fmt.Printf("%v\n", s.chainA.App.(*app.StrideApp).BankKeeper.GetAllBalances(s.chainA.GetContext(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.chainB.App.(*simapp.SimApp).BankKeeper.GetAllBalances(s.chainB.GetContext(), addr2))

	s.chainA.App.(*app.StrideApp).TransferKeeper.SendTransfer(
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
	fmt.Printf("%v\n", s.chainA.App.(*app.StrideApp).BankKeeper.GetAllBalances(s.chainA.GetContext(), addr1))
	fmt.Println(addr2.String())
	fmt.Printf("%v\n", s.chainB.App.(*simapp.SimApp).BankKeeper.GetAllBalances(s.chainB.GetContext(), addr2))
}

func (s *MultiChainKeeperTestSuite) TestIca() {
	s.SetupIbc()

	delegationAddress, found := s.chainA.App.(*app.StrideApp).ICAControllerKeeper.GetInterchainAccountAddress(s.chainA.GetContext(), "connection-0", "icacontroller-GAIA.DELEGATION")
	s.Require().True(found)
	fmt.Println("DELEGATION ADDRESS:", delegationAddress)

	feeAddress, found := s.chainA.App.(*app.StrideApp).ICAControllerKeeper.GetInterchainAccountAddress(s.chainA.GetContext(), "connection-0", "icacontroller-GAIA.FEE")
	s.Require().True(found)
	fmt.Println("FEE ADDRESS:", feeAddress)

	redemptionAddress, found := s.chainA.App.(*app.StrideApp).ICAControllerKeeper.GetInterchainAccountAddress(s.chainA.GetContext(), "connection-0", "icacontroller-GAIA.REDEMPTION")
	s.Require().True(found)
	fmt.Println("REDEMPTION ADDRESS:", redemptionAddress)

	withdrawAddress, found := s.chainA.App.(*app.StrideApp).ICAControllerKeeper.GetInterchainAccountAddress(s.chainA.GetContext(), "connection-0", "icacontroller-GAIA.WITHDRAWAL")
	s.Require().True(found)
	fmt.Println("WITHDRAWAL ADDRESS:", withdrawAddress)

	timeoutTimestamp := uint64(s.chainA.GetContext().BlockTime().UnixNano() + 100_000_000_000)

	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{
		FromAddress: delegationAddress,
		ToAddress:   feeAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1000))),
	})

	data, err := icatypes.SerializeCosmosTx(s.chainA.Codec, msgs)
	s.Require().NoError(err)

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	chanCap, found := s.chainA.App.(*app.StrideApp).ScopedIBCKeeper.GetCapability(s.chainA.GetContext(), host.ChannelCapabilityPath("icacontroller-GAIA.DELEGATION", "channel-1"))
	s.Require().True(found)

	seq, err := s.chainA.App.(*app.StrideApp).ICAControllerKeeper.SendTx(s.chainA.GetContext(), chanCap, "connection-0", "icacontroller-GAIA.DELEGATION", packetData, timeoutTimestamp)
	s.Require().NoError(err)
	fmt.Println("SEQUENCE:", seq)
}

func TestMultiChainKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(MultiChainKeeperTestSuite))
}
