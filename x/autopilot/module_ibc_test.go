package autopilot_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	"github.com/Stride-Labs/stride/v27/x/autopilot"
	recordsmodule "github.com/Stride-Labs/stride/v27/x/records"
)

type ModuleIBCTestSuite struct {
	apptesting.AppTestHelper
	module autopilot.IBCModule
}

func (s *ModuleIBCTestSuite) SetupTest() {
	s.Setup()
	// Create the full IBC stack similar to how it's done in the other tests
	transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
	recordsStack := recordsmodule.NewIBCModule(s.App.RecordsKeeper, transferIBCModule)
	s.module = autopilot.NewIBCModule(s.App.AutopilotKeeper, recordsStack)
}

func TestModuleIBCTestSuite(t *testing.T) {
	suite.Run(t, new(ModuleIBCTestSuite))
}

func (s *ModuleIBCTestSuite) TestOnRecvPacket_ValidBech32WithoutMemo() {
	// Test with valid address and no memo - should bypass autopilot validation
	validAddress := s.TestAccs[0].String()
	tokenPacketData := transfertypes.FungibleTokenPacketData{
		Denom:    "uatom",
		Amount:   "1000000",
		Sender:   "cosmos1sender",
		Receiver: validAddress,
		Memo:     "", // No memo = no autopilot needed
	}

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&tokenPacketData),
	}

	relayer := sdk.AccAddress("relayer")

	ack := s.module.OnRecvPacket(s.Ctx, packet, relayer)

	if !ack.Success() {
		// If it fails, it should not be due to autopilot receiver address validation (error code 1506)
		s.Require().NotContains(string(ack.Acknowledgement()), "ABCI code: 1506",
			"valid address without memo should not fail due to autopilot address validation")
		// It may fail due to other reasons (transfer module validation, missing denom, etc.)
		s.T().Logf("Packet failed due to other validation (expected): %s", string(ack.Acknowledgement()))
	}
}

// This test demonstrates that invalid bech32 addresses are now handled correctly:
// - Without memo: Bypasses autopilot validation (gets rejected later by transfer module with error code 1)
// - With memo: Gets caught by autopilot validation (error code 1506)
// This is the fix for Namada's bech32m address compatibility issue
func (s *ModuleIBCTestSuite) TestInvalidBech32AddressHandling() {
	invalidBech32Address := "tnam1qr8h3ga7rg76s8j4w4ks7hx6a5vxrd5r4n0ux6p4ez"

	// Test 1: Without memo - should bypass autopilot validation
	tokenPacketDataNoMemo := transfertypes.FungibleTokenPacketData{
		Denom:    "uatom",
		Amount:   "1000000",
		Sender:   "cosmos1sender",
		Receiver: invalidBech32Address,
		Memo:     "", // No memo = no autopilot processing
	}

	packetNoMemo := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&tokenPacketDataNoMemo),
	}

	// Test 2: With memo - should be caught by autopilot validation
	tokenPacketDataWithMemo := transfertypes.FungibleTokenPacketData{
		Denom:    "uatom",
		Amount:   "1000000",
		Sender:   "cosmos1sender",
		Receiver: invalidBech32Address,
		Memo:     `{"autopilot": {"receiver": "stride1valid", "stakeibc": {"action": "LiquidStake"}}}`,
	}

	packetWithMemo := channeltypes.Packet{
		Sequence:           2,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&tokenPacketDataWithMemo),
	}

	relayer := sdk.AccAddress("relayer")

	ackNoMemo := s.module.OnRecvPacket(s.Ctx, packetNoMemo, relayer)
	ackWithMemo := s.module.OnRecvPacket(s.Ctx, packetWithMemo, relayer)

	// Verify the different error sources
	s.Require().False(ackNoMemo.Success(), "without memo should still fail (from transfer module)")
	s.Require().Contains(string(ackNoMemo.Acknowledgement()), "ABCI code: 1",
		"without memo should fail with transfer module error (code 1)")

	s.Require().False(ackWithMemo.Success(), "with memo should fail at autopilot validation")
	s.Require().Contains(string(ackWithMemo.Acknowledgement()), "ABCI code: 1506",
		"with memo should fail with autopilot validation error (code 1506)")
}

func (s *ModuleIBCTestSuite) TestOnRecvPacket_ValidBech32WithMemo() {
	// Test that packets with valid bech32 addresses and memo pass autopilot address validation

	validAddress := s.TestAccs[0].String()
	tokenPacketData := transfertypes.FungibleTokenPacketData{
		Denom:    "uatom",
		Amount:   "1000000",
		Sender:   "cosmos1sender",
		Receiver: validAddress,
		Memo:     `{"autopilot": {"receiver": "` + validAddress + `", "stakeibc": {"action": "LiquidStake"}}}`,
	}

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-1",
		Data:               transfertypes.ModuleCdc.MustMarshalJSON(&tokenPacketData),
	}

	relayer := sdk.AccAddress("relayer")

	ack := s.module.OnRecvPacket(s.Ctx, packet, relayer)

	if !ack.Success() {
		// Should not fail due to autopilot receiver address validation (error code 1506)
		// since we're using a valid bech32 address
		s.Require().NotContains(string(ack.Acknowledgement()), "ABCI code: 1506",
			"valid bech32 address should pass autopilot receiver address validation")
		// May fail for other reasons (missing host zone, invalid memo format, etc.)
		s.T().Logf("Packet failed due to other validation (may be expected): %s", string(ack.Acknowledgement()))
	}
}
