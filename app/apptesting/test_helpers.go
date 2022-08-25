package apptesting

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/Stride-Labs/stride/app"
)

var (
	StrideChainID = "STRIDE"

	TestIcaVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

type AppTestHelper struct {
	suite.Suite

	App *app.StrideApp
	Ctx sdk.Context

	Coordinator  *ibctesting.Coordinator
	StrideChain  *ibctesting.TestChain
	HostChain    *ibctesting.TestChain
	TransferPath *ibctesting.Path

	QueryHelper *baseapp.QueryServiceTestHelper
	TestAccs    []sdk.AccAddress
}

func (s *AppTestHelper) Setup() {
	s.App = app.InitStrideTestApp(true)
	s.Ctx = s.App.BaseApp.NewContext(false, tmtypes.Header{Height: 1, ChainID: "STRIDE"})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)
}

func (s *AppTestHelper) FundModuleAccount(moduleName string, amount sdk.Coin) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, moduleName, sdk.NewCoins(amount))
	s.Require().NoError(err)
}

func (s *AppTestHelper) FundAccount(acc sdk.AccAddress, amount sdk.Coin) {
	amountCoins := sdk.NewCoins(amount)
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amountCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, amountCoins)
	s.Require().NoError(err)
}

func (s *AppTestHelper) CompareCoins(expectedCoin sdk.Coin, actualCoin sdk.Coin, msg string) {
	s.Require().Equal(expectedCoin.Amount.Int64(), actualCoin.Amount.Int64(), msg)
}

func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// Enables IBC support through ibctesting and creates a transfer channel
func (s *AppTestHelper) SetupIBC(hostChainID string) {
	s.Coordinator = ibctesting.NewCoordinator(s.T(), 0)

	// Iniitalize stride testing app with by casting a StrideApp -> TestingApp
	ibctesting.DefaultTestingAppInit = app.InitStrideIBCTestingApp
	s.StrideChain = ibctesting.NewTestChain(s.T(), s.Coordinator, StrideChainID)

	// Initialize a host testing app using SimApp -> TestingApp
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp
	s.HostChain = ibctesting.NewTestChain(s.T(), s.Coordinator, hostChainID)

	// Update coordinator
	s.Coordinator.Chains = map[string]*ibctesting.TestChain{
		StrideChainID: s.StrideChain,
		hostChainID:   s.HostChain,
	}

	// Create a clients, connections, and transfer channel
	s.TransferPath = NewTransferPath(s.StrideChain, s.HostChain)
	s.Coordinator.Setup(s.TransferPath)

	// Replace AppTestHelper's app with the TestingApp
	s.App = s.StrideChain.App.(*app.StrideApp)

	// Finally confirm the channel was setup properly
	s.Require().Equal("07-tendermint-0", s.TransferPath.EndpointA.ClientID, "stride clientID")
	s.Require().Equal("connection-0", s.TransferPath.EndpointA.ConnectionID, "stride connectionID")
	s.Require().Equal("channel-0", s.TransferPath.EndpointA.ChannelID, "stride transfer channelID")

	s.Require().Equal("07-tendermint-0", s.TransferPath.EndpointB.ClientID, "host clientID")
	s.Require().Equal("connection-0", s.TransferPath.EndpointB.ConnectionID, "host connectionID")
	s.Require().Equal("channel-0", s.TransferPath.EndpointB.ChannelID, "host transfer channelID")
}

// Creates an ICA channel through ibctesting
// Also creates a transfer channel is if hasn't been done yet
func (s *AppTestHelper) CreateICAChannel(owner string) {
	if s.TransferPath == nil {
		// If we have yet to setup IBC, do that here
		ownerSplit := strings.Split(owner, ".")
		s.Require().Equal(2, len(ownerSplit), "owner should be of the form: {HostZone}.{AccountName}")

		hostChainId := ownerSplit[0]
		s.SetupIBC(hostChainId)
	}

	// Create ICA Path and the copy over the client and connection from the transfer path
	icaPath := NewIcaPath(s.StrideChain, s.HostChain)
	icaPath = CopyConnectionAndClientToPath(icaPath, s.TransferPath)

	// Register the ICA
	s.RegisterInterchainAccount(icaPath.EndpointA, owner)

	// Complete the handshake
	err := icaPath.EndpointB.ChanOpenTry()
	s.Require().NoError(err, "ChanOpenTry error")

	err = icaPath.EndpointA.ChanOpenAck()
	s.Require().NoError(err, "ChanOpenAck error")

	icaPath.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err, "ChanOpenConfirm error")

	// Finally, confirm the ICA channel was created properly
	portID := icaPath.EndpointA.ChannelConfig.PortID
	channelID := icaPath.EndpointA.ChannelID
	_, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.StrideChain.GetContext(), portID, channelID)
	s.Require().True(found, fmt.Sprintf("Channel not found after creation, PortID: %s, ChannelID: %s", portID, channelID))
}

// Creates a transfer channel between two chains
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

// Creates an ICA channel between two chains
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

// In ibctesting, there's no easy way to create a new channel on an existing connection
// To get around this, this helper function will copy the client/connection info from an existing channel
// We use this when creating ICA channels, because we want to reuse the same connections/clients from the transfer channel
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

// Register's a new ICA account on the next channel available
// This function assumes a connection already exists
func (s *AppTestHelper) RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) {
	// Get the port ID from the owner name (i.e. "icacontroller-{owner}")
	portID, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "owner to portID error")

	// Get the next channel available
	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	// Register an ICA at that
	err = s.App.ICAControllerKeeper.RegisterInterchainAccount(endpoint.Chain.GetContext(), endpoint.ConnectionID, owner)
	s.Require().NoError(err, "register interchain account error")

	// Commit the state
	endpoint.Chain.App.Commit()
	endpoint.Chain.NextBlock()

	// Update the endpoint object to the newly created port + channel
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID
}
