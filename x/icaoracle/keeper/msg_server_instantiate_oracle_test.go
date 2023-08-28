package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

type InstantiateOracleTestCase struct {
	OracleChannelId string
	OraclePortId    string
	ValidMsg        types.MsgInstantiateOracle
	InitialOracle   types.Oracle
}

func (s *KeeperTestSuite) SetupTestInstantiateOracle() InstantiateOracleTestCase {
	// Create oracle ICA channel
	owner := types.FormatICAAccountOwner(HostChainId, types.ICAAccountType_Oracle)
	channelId, portId := s.CreateICAChannel(owner)

	// Create oracle
	oracle := types.Oracle{
		ChainId:      HostChainId,
		ConnectionId: ibctesting.FirstConnectionID,
		ChannelId:    channelId,
		PortId:       portId,
		IcaAddress:   "ica_address",
	}
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, oracle)

	// Confirm the oracle was stored
	_, found := s.App.ICAOracleKeeper.GetOracle(s.Ctx, HostChainId)
	s.Require().True(found, "oracle should be in the store during setup")

	return InstantiateOracleTestCase{
		OracleChannelId: channelId,
		OraclePortId:    portId,
		ValidMsg: types.MsgInstantiateOracle{
			OracleChainId:  HostChainId,
			ContractCodeId: uint64(1),
		},
		InitialOracle: oracle,
	}
}

func (s *KeeperTestSuite) TestInstantiateOracle_Successful() {
	tc := s.SetupTestInstantiateOracle()

	// Submit the instantiate message
	_, err := s.GetMsgServer().InstantiateOracle(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().NoError(err, "no error expected when instantiating an oracle")

	// Confirm the callback data was stored
	callbackKey := icacallbacktypes.PacketID(tc.OraclePortId, tc.OracleChannelId, uint64(1))
	_, found := s.App.IcacallbacksKeeper.GetCallbackData(s.Ctx, callbackKey)
	s.Require().True(found, "callback data should have been found")
}

func (s *KeeperTestSuite) TestInstantiateOracle_Failure_OracleNotFound() {
	tc := s.SetupTestInstantiateOracle()

	// Remove the oracle from the store
	s.App.ICAOracleKeeper.RemoveOracle(s.Ctx, HostChainId)

	// Submit the instantiate message - it should fail
	_, err := s.GetMsgServer().InstantiateOracle(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "oracle not found")
}

func (s *KeeperTestSuite) TestInstantiateOracle_Failure_OracleAlreadyInstantiated() {
	tc := s.SetupTestInstantiateOracle()

	// Set the oracle contract address to appear as if it was already instantiated
	instantiatedOracle := tc.InitialOracle
	instantiatedOracle.ContractAddress = "contract"
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, instantiatedOracle)

	// Submit the instantiate message - it should fail
	_, err := s.GetMsgServer().InstantiateOracle(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "oracle already instantiated")
}

func (s *KeeperTestSuite) TestInstantiateOracle_Failure_InvalidICASetup() {
	tc := s.SetupTestInstantiateOracle()

	// Remove the oracle ICA address to appear as if it the ICA was not registered yet
	instantiatedOracle := tc.InitialOracle
	instantiatedOracle.IcaAddress = ""
	s.App.ICAOracleKeeper.SetOracle(s.Ctx, instantiatedOracle)

	// Submit the instantiate message - it should fail
	_, err := s.GetMsgServer().InstantiateOracle(sdk.WrapSDKContext(s.Ctx), &tc.ValidMsg)
	s.Require().ErrorContains(err, "oracle ICA channel has not been registered")
}
