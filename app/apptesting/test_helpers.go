package apptesting

import (
	"strings"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmtypesproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/cosmos/ibc-go/v7/testing/simapp"
	appProvider "github.com/cosmos/interchain-security/v3/app/provider"
	ibctesting "github.com/cosmos/interchain-security/v3/legacy_ibc_testing/testing"
	icstestingutils "github.com/cosmos/interchain-security/v3/testutil/ibc_testing"
	e2e "github.com/cosmos/interchain-security/v3/testutil/integration"
	ccvprovidertypes "github.com/cosmos/interchain-security/v3/x/ccv/provider/types"
	ccvtypes "github.com/cosmos/interchain-security/v3/x/ccv/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app"
	"github.com/Stride-Labs/stride/v14/utils"
)

var (
	StrideChainID   = "STRIDE"
	ProviderChainID = "PROVIDER"
	FirstClientId   = "07-tendermint-0"

	TestIcaVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

type SuitelessAppTestHelper struct {
	App *app.StrideApp
	Ctx sdk.Context
}

type AppTestHelper struct {
	suite.Suite

	App         *app.StrideApp
	HostApp     *simapp.SimApp
	ProviderApp e2e.ProviderApp

	IbcEnabled    bool
	Coordinator   *ibctesting.Coordinator
	StrideChain   *ibctesting.TestChain
	HostChain     *ibctesting.TestChain
	ProviderChain *ibctesting.TestChain
	TransferPath  *ibctesting.Path

	QueryHelper  *baseapp.QueryServiceTestHelper
	TestAccs     []sdk.AccAddress
	IcaAddresses map[string]string
	Ctx          sdk.Context
}

// AppTestHelper Constructor
func (s *AppTestHelper) Setup() {
	s.App = app.InitStrideTestApp(true)
	s.Ctx = s.App.BaseApp.NewContext(false, tmtypesproto.Header{Height: 1, ChainID: StrideChainID})
	s.QueryHelper = &baseapp.QueryServiceTestHelper{
		GRPCQueryRouter: s.App.GRPCQueryRouter(),
		Ctx:             s.Ctx,
	}
	s.TestAccs = CreateRandomAccounts(3)
	s.IbcEnabled = false
	s.IcaAddresses = make(map[string]string)
}

// Instantiates an TestHelper without the test suite
// This is for testing scenarios where we simply need the setup function to run,
// and need access to the TestHelper attributes and keepers (e.g. genesis tests)
func SetupSuitelessTestHelper() SuitelessAppTestHelper {
	s := SuitelessAppTestHelper{}
	s.App = app.InitStrideTestApp(true)
	s.Ctx = s.App.BaseApp.NewContext(false, tmtypesproto.Header{Height: 1, ChainID: StrideChainID})
	return s
}

// Mints coins directly to a module account
func (s *AppTestHelper) FundModuleAccount(moduleName string, amount sdk.Coin) {
	amountCoins := sdk.NewCoins(amount)
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amountCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, minttypes.ModuleName, moduleName, amountCoins)
	s.Require().NoError(err)
}

// Mints and sends coins to a user account
func (s *AppTestHelper) FundAccount(acc sdk.AccAddress, amount sdk.Coin) {
	amountCoins := sdk.NewCoins(amount)
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, amountCoins)
	s.Require().NoError(err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, amountCoins)
	s.Require().NoError(err)
}

// Helper function to compare coins with a more legible error
func (s *AppTestHelper) CompareCoins(expectedCoin sdk.Coin, actualCoin sdk.Coin, msg string) {
	s.Require().Equal(expectedCoin.Amount.Int64(), actualCoin.Amount.Int64(), msg)
}

// Generate random account addresss
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}

// Initializes a ibctesting coordinator to keep track of Stride and a host chain's state
func (s *AppTestHelper) SetupIBCChains(hostChainID string) {
	s.Coordinator = ibctesting.NewCoordinator(s.T(), 0)

	// Initialize a provider testing app
	s.ProviderChain = ibctesting.NewTestChain(s.T(), s.Coordinator, icstestingutils.ProviderAppIniter, ProviderChainID)
	s.ProviderApp = s.ProviderChain.App.(*appProvider.App)

	// Initialize a stride testing app by casting a StrideApp -> TestingApp
	s.StrideChain = ibctesting.NewTestChainWithValSet(s.T(), s.Coordinator, app.InitStrideIBCTestingApp, StrideChainID,
		s.ProviderChain.Vals, s.ProviderChain.Signers)

	// Initialize a host testing app using SimApp -> TestingApp
	s.HostChain = ibctesting.NewTestChain(s.T(), s.Coordinator, ibctesting.SetupTestingApp, hostChainID)

	// Update coordinator
	s.Coordinator.Chains = map[string]*ibctesting.TestChain{
		StrideChainID:   s.StrideChain,
		hostChainID:     s.HostChain,
		ProviderChainID: s.ProviderChain,
	}
	s.IbcEnabled = true

	// valsets must match
	providerValUpdates := tmtypes.TM2PB.ValidatorUpdates(s.ProviderChain.Vals)
	strideValUpdates := tmtypes.TM2PB.ValidatorUpdates(s.StrideChain.Vals)
	s.Require().True(len(providerValUpdates) == len(strideValUpdates), "initial valset not matching")

	// for i := 0; i < len(providerValUpdates); i++ {
	// 	addr1 := ccvutils.GetChangePubKeyAddress(providerValUpdates[i])
	// 	addr2 := ccvutils.GetChangePubKeyAddress(strideValUpdates[i])
	// 	s.Require().True(bytes.Equal(addr1, addr2), "validator mismatch")
	// }

	// move chains to the next block
	s.ProviderChain.NextBlock()
	s.StrideChain.NextBlock()
	s.HostChain.NextBlock()

	providerKeeper := s.ProviderApp.GetProviderKeeper()
	// create consumer client on provider chain and set as consumer client for consumer chainID in provider keeper.
	err := providerKeeper.CreateConsumerClient(
		s.ProviderChain.GetContext(),
		&ccvprovidertypes.ConsumerAdditionProposal{
			ChainId:                           s.StrideChain.ChainID,
			InitialHeight:                     s.StrideChain.LastHeader.GetHeight().(clienttypes.Height),
			BlocksPerDistributionTransmission: 50,
			CcvTimeoutPeriod:                  time.Hour,
			TransferTimeoutPeriod:             time.Hour,
			UnbondingPeriod:                   time.Hour * 504,
			ConsumerRedistributionFraction:    "0.75",
			HistoricalEntries:                 10000,
		},
	)
	s.Require().NoError(err)

	// move provider to next block to commit the state
	s.ProviderChain.NextBlock()

	// initialize the consumer chain with the genesis state stored on the provider
	strideConsumerGenesis, found := providerKeeper.GetConsumerGenesis(
		s.ProviderChain.GetContext(),
		s.StrideChain.ChainID,
	)
	s.Require().True(found, "consumer genesis not found")
	s.StrideChain.App.(*app.StrideApp).GetConsumerKeeper().InitGenesis(s.StrideChain.GetContext(), &strideConsumerGenesis)
}

// Creates clients, connections, and a transfer channel between stride and a host chain
func (s *AppTestHelper) CreateTransferChannel(hostChainID string) {
	// If we have yet to create the host chain, do that here
	if !s.IbcEnabled {
		s.SetupIBCChains(hostChainID)
	}
	s.Require().Equal(s.HostChain.ChainID, hostChainID,
		"The testing app has already been initialized with a different chainID (%s)", s.HostChain.ChainID)

	// Create clients, connections, and a transfer channel
	s.TransferPath = NewTransferPath(s.StrideChain, s.HostChain, s.ProviderChain)
	s.Coordinator.Setup(s.TransferPath)

	// Replace stride and host apps with those from TestingApp
	s.App = s.StrideChain.App.(*app.StrideApp)
	// s.HostApp = s.HostChain.GetSimApp()
	s.Ctx = s.StrideChain.GetContext()

	// Finally confirm the channel was setup properly
	s.Require().Equal("07-tendermint-1", s.TransferPath.EndpointA.ClientID, "stride clientID")
	s.Require().Equal(ibctesting.FirstConnectionID, s.TransferPath.EndpointA.ConnectionID, "stride connectionID")
	s.Require().Equal(ibctesting.FirstChannelID, s.TransferPath.EndpointA.ChannelID, "stride transfer channelID")
}

// Creates an ICA channel through ibctesting
// Also creates a transfer channel is if hasn't been done yet
func (s *AppTestHelper) CreateICAChannel(owner string) (channelID, portID string) {
	// If we have yet to create a client/connection (through creating a transfer channel), do that here
	_, transferChannelExists := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, ibctesting.TransferPort, ibctesting.FirstChannelID)
	if !transferChannelExists {
		ownerSplit := strings.Split(owner, ".")
		s.Require().Equal(2, len(ownerSplit), "owner should be of the form: {HostZone}.{AccountName}")

		hostChainID := ownerSplit[0]
		s.CreateTransferChannel(hostChainID)
	}

	// Create ICA Path and then copy over the client and connection from the transfer path
	icaPath := NewIcaPath(s.StrideChain, s.HostChain, s.ProviderChain)
	icaPath = CopyConnectionAndClientToPath(icaPath, s.TransferPath)

	// Register the ICA and complete the handshake
	s.RegisterInterchainAccount(icaPath.EndpointA, owner)

	err := icaPath.EndpointB.ChanOpenTry()
	s.Require().NoError(err, "ChanOpenTry error")

	err = icaPath.EndpointA.ChanOpenAck()
	s.Require().NoError(err, "ChanOpenAck error")

	err = icaPath.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err, "ChanOpenConfirm error")

	s.Ctx = s.StrideChain.GetContext()

	// Confirm the ICA channel was created properly
	portID = icaPath.EndpointA.ChannelConfig.PortID
	channelID = icaPath.EndpointA.ChannelID
	_, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portID, channelID)
	s.Require().True(found, "Channel not found after creation, PortID: %s, ChannelID: %s", portID, channelID)

	// Store the account address
	icaAddress, found := s.App.ICAControllerKeeper.GetInterchainAccountAddress(s.Ctx, ibctesting.FirstConnectionID, portID)
	s.Require().True(found, "can't get ICA address")
	s.IcaAddresses[owner] = icaAddress

	// Finally set the active channel
	s.App.ICAControllerKeeper.SetActiveChannelID(s.Ctx, ibctesting.FirstConnectionID, portID, channelID)

	return channelID, portID
}

// Register's a new ICA account on the next channel available
// This function assumes a connection already exists
func (s *AppTestHelper) RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) {
	// Get the port ID from the owner name (i.e. "icacontroller-{owner}")
	portID, err := icatypes.NewControllerPortID(owner)
	s.Require().NoError(err, "owner to portID error")

	// Get the next channel available and register the ICA
	channelSequence := s.App.IBCKeeper.ChannelKeeper.GetNextChannelSequence(s.Ctx)

	err = s.App.ICAControllerKeeper.RegisterInterchainAccount(s.Ctx, endpoint.ConnectionID, owner, TestIcaVersion)
	s.Require().NoError(err, "register interchain account error")

	// Commit the state
	endpoint.Chain.NextBlock()

	// Update the endpoint object to the newly created port + channel
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID
}

// Creates a transfer channel between two chains
func NewTransferPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain, providerChain *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	trustingPeriodFraction := providerChain.App.(*appProvider.App).GetProviderKeeper().GetTrustingPeriodFraction(providerChain.GetContext())
	consumerUnbondingPeriod := path.EndpointA.Chain.App.(*app.StrideApp).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointA.Chain.GetContext())
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriod
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccvtypes.CalculateTrustPeriod(consumerUnbondingPeriod, trustingPeriodFraction)

	return path
}

// Creates an ICA channel between two chains
func NewIcaPath(chainA *ibctesting.TestChain, chainB *ibctesting.TestChain, providerChain *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestIcaVersion
	path.EndpointB.ChannelConfig.Version = TestIcaVersion

	trustingPeriodFraction := providerChain.App.(*appProvider.App).GetProviderKeeper().GetTrustingPeriodFraction(providerChain.GetContext())
	consumerUnbondingPeriod := path.EndpointA.Chain.App.(*app.StrideApp).GetConsumerKeeper().GetUnbondingPeriod(path.EndpointA.Chain.GetContext())
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = consumerUnbondingPeriod
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = ccvtypes.CalculateTrustPeriod(consumerUnbondingPeriod, trustingPeriodFraction)

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

// Helper function to change the state of a channel (i.e. to open/close it)
func (s *AppTestHelper) UpdateChannelState(portId, channelId string, channelState channeltypes.State) {
	channel, found := s.App.IBCKeeper.ChannelKeeper.GetChannel(s.Ctx, portId, channelId)
	s.Require().True(found, "ica channel should have been found")
	channel.State = channelState
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, portId, channelId, channel)
}

// Constructs an ICA Packet Acknowledgement compatible with ibc-go v5+
func ICAPacketAcknowledgement(t *testing.T, msgType string, msgResponses []proto.Message) channeltypes.Acknowledgement {
	txMsgData := &sdk.TxMsgData{
		MsgResponses: make([]*codectypes.Any, len(msgResponses)),
	}
	for i, msgResponse := range msgResponses {
		var value []byte
		var err error
		if msgResponse != nil {
			value, err = proto.Marshal(msgResponse)
			require.NoError(t, err, "marshal error")
		}

		txMsgData.MsgResponses[i] = &codectypes.Any{
			TypeUrl: msgType,
			Value:   value,
		}
	}
	marshalledTxMsgData, err := proto.Marshal(txMsgData)
	require.NoError(t, err)
	ack := channeltypes.NewResultAcknowledgement(marshalledTxMsgData)
	return ack
}

// Constructs an legacy ICA Packet Acknowledgement compatible with ibc-go version v4 and lower
func ICAPacketAcknowledgementLegacy(t *testing.T, msgType string, msgResponses []proto.Message) channeltypes.Acknowledgement {
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, len(msgResponses)), //nolint:staticcheck
	}
	for i, msgResponse := range msgResponses {
		var data []byte
		var err error
		if msgResponse != nil {
			data, err = proto.Marshal(msgResponse)
			require.NoError(t, err, "marshal error")
		}

		txMsgData.Data[i] = &sdk.MsgData{ //nolint:staticcheck
			MsgType: msgType,
			Data:    data,
		}
	}
	marshalledTxMsgData, err := proto.Marshal(txMsgData)
	require.NoError(t, err)
	ack := channeltypes.NewResultAcknowledgement(marshalledTxMsgData)
	return ack
}

// Get an IBC denom from it's native host denom
// This assumes the transfer channel is channel-0
func (s *AppTestHelper) GetIBCDenomTrace(denom string) transfertypes.DenomTrace {
	sourcePrefix := transfertypes.GetDenomPrefix(ibctesting.TransferPort, ibctesting.FirstChannelID)
	prefixedDenom := sourcePrefix + denom

	return transfertypes.ParseDenomTrace(prefixedDenom)
}

// Creates and stores an IBC denom from a base denom on transfer channel-0
// This is only required for tests that use the transfer keeper and require that the IBC
// denom is present in the store
//
// Returns the IBC hash
func (s *AppTestHelper) CreateAndStoreIBCDenom(baseDenom string) (ibcDenom string) {
	denomTrace := s.GetIBCDenomTrace(baseDenom)
	s.App.TransferKeeper.SetDenomTrace(s.Ctx, denomTrace)
	return denomTrace.IBCDenom()
}

func (s *AppTestHelper) MarshalledICS20PacketData() sdk.AccAddress {
	data := ibctransfertypes.FungibleTokenPacketData{}
	return data.GetBytes()
}

// Helper function to mock out a connection, client, and revision height
func (s *AppTestHelper) MockClientLatestHeight(height uint64) {
	clientState := tendermint.ClientState{
		LatestHeight: clienttypes.NewHeight(1, height),
	}
	connection := connectiontypes.ConnectionEnd{
		ClientId: FirstClientId,
	}
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, ibctesting.FirstConnectionID, connection)
	s.App.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, FirstClientId, &clientState)
}

func (s *AppTestHelper) ConfirmUpgradeSucceededs(upgradeName string, upgradeHeight int64) {
	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight - 1)
	plan := upgradetypes.Plan{Name: upgradeName, Height: upgradeHeight}
	err := s.App.UpgradeKeeper.ScheduleUpgrade(s.Ctx, plan)
	s.Require().NoError(err)
	_, exists := s.App.UpgradeKeeper.GetUpgradePlan(s.Ctx)
	s.Require().True(exists)

	s.Ctx = s.Ctx.WithBlockHeight(upgradeHeight)
	s.Require().NotPanics(func() {
		beginBlockRequest := abci.RequestBeginBlock{}
		s.App.BeginBlocker(s.Ctx, beginBlockRequest)
	})
}

// Generates a valid and invalid test address (used for non-keeper tests)
func GenerateTestAddrs() (string, string) {
	pk1 := ed25519.GenPrivKey().PubKey()
	validAddr := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid").String()
	return validAddr, invalidAddr
}

// Grabs an admin address to test validate basic on admin txs
func GetAdminAddress() (address string, ok bool) {
	for address := range utils.Admins {
		return address, true
	}
	return "", false
}

// Modifies sdk config to have stride address prefixes (used for non-keeper tests)
func SetupConfig() {
	app.SetupConfig()
}

// Searches for an event using the current context
func (s *AppTestHelper) getEventsFromEventType(eventType string) (events []sdk.Event) {
	for _, event := range s.Ctx.EventManager().Events() {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events
}

// Searches for an event attribute, given an event
// Returns the value if found
func (s *AppTestHelper) getEventValuesFromAttribute(event sdk.Event, attributeKey string) (values []string) {
	for _, attribute := range event.Attributes {
		if string(attribute.Key) == attributeKey {
			values = append(values, string(attribute.Value))
		}
	}
	return values
}

// Searches for an event that has an attribute value matching the expected value
// Returns whether there was a match, as well as a list for all the values found
// for that attribute (for the error message)
func (s *AppTestHelper) checkEventAttributeValueMatch(
	events []sdk.Event,
	attributeKey,
	expectedValue string,
) (allValues []string, found bool) {
	for _, event := range events {
		allValues = append(allValues, s.getEventValuesFromAttribute(event, attributeKey)...)
		for _, actualValue := range allValues {
			if actualValue == expectedValue {
				found = true
			}
		}
	}
	return allValues, found
}

// Checks if an event was emitted
func (s *AppTestHelper) CheckEventTypeEmitted(eventType string) []sdk.Event {
	events := s.getEventsFromEventType(eventType)
	eventEmitted := len(events) > 0
	s.Require().True(eventEmitted, "%s event should have been emitted", eventType)
	return events
}

// Checks that an event was not emitted
func (s *AppTestHelper) CheckEventTypeNotEmitted(eventType string) {
	events := s.getEventsFromEventType(eventType)
	eventNotEmitted := len(events) == 0
	s.Require().True(eventNotEmitted, "%s event should not have been emitted", eventType)
}

// Checks that an event was emitted and that the value matches expectations
func (s *AppTestHelper) CheckEventValueEmitted(eventType, attributeKey, expectedValue string) {
	events := s.CheckEventTypeEmitted(eventType)

	// Check all events and attributes for a match
	allValues, valueFound := s.checkEventAttributeValueMatch(events, attributeKey, expectedValue)
	s.Require().True(valueFound, "attribute %s with value %s should have been found in event %s. Values emitted for attribute: %+v",
		attributeKey, expectedValue, eventType, allValues)
}

// Checks that there was no event emitted that matches the event type, attribute, and value
func (s *AppTestHelper) CheckEventValueNotEmitted(eventType, attributeKey, expectedValue string) {
	// Check that either the event or attribute were not emitted
	events := s.getEventsFromEventType(eventType)
	if len(events) == 0 {
		return
	}

	// Check all events and attributes to make sure there's no match
	allValues, valueFound := s.checkEventAttributeValueMatch(events, attributeKey, expectedValue)
	s.Require().False(valueFound, "attribute %s with value %s should not have been found in event %s. Values emitted for attribute: %+v",
		attributeKey, expectedValue, eventType, allValues)
}
