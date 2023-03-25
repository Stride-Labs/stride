package keeper_test

import (
	"fmt"

	"github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/utils"
	"github.com/Stride-Labs/stride/v8/x/autopilot"
	"github.com/Stride-Labs/stride/v8/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v8/x/claim/types"
)

func getClaimPacketMetadata(address, airdropId string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"claim": { "stride_address": "%[1]s", "airdrop_id": "%[2]s" } 
			}
		}`, address, airdropId)
}

func (s *KeeperTestSuite) TestAirdropOnRecvPacket() {
	evmosAirdropId := "evmos"
	evmosDenom := "aevmos"

	// The evmos addresses represent the airdrop recipient
	evmosAddress := "evmos1wg6vh689gw93umxqquhe3yaqf0h9wt9d4q7550"

	// Each evmos address has a serialized mapping that was used to store the claim record
	// This is in the form of an "incorrect" stride address and was stored during the upgrade
	evmosAddressKeyString := utils.ConvertAddressToStrideAddress(evmosAddress)
	evmosAddressKey := sdk.MustAccAddressFromBech32(evmosAddressKeyString)

	// For each evmos address, there is a corresponding stride address that will specified
	// in the transfer packet - so for the sake of this test, we'll use arbitrary stride addresses
	strideAccAddress := s.TestAccs[0]
	strideAddress := strideAccAddress.String()

	// Build the template for the transfer packet (the data and channel fields will get updated from each unit test)
	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-0",
		Data:               []byte{},
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   0,
	}

	testCases := []struct {
		name                  string
		forwardingActive      bool
		packetData            transfertypes.FungibleTokenPacketData
		transferShouldSucceed bool
		airdropShouldUpdate   bool
	}{
		{
			name:             "successful airdrop update from receiver",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: getClaimPacketMetadata(strideAddress, evmosAirdropId),
				Memo:     "",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   true,
		},
		{
			name:             "successful airdrop update from memo",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     getClaimPacketMetadata(strideAddress, evmosAirdropId),
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   true,
		},
		{
			name:             "airdrop does not exist",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     getClaimPacketMetadata(strideAddress, "fake_airdrop"),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:             "invalid stride address",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uevmos",
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     getClaimPacketMetadata("invalid_address", evmosAirdropId),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:             "normal transfer packet - no memo",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     "",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   false,
		},
		{
			name:             "normal transfer packet - empty JSON memo",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     "{}",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   false,
		},
		{
			name:             "invalid autopilot JSON - no receiver",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     `{ "autopilot": {} }`,
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:             "invalid autopilot JSON - no routing module",
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    evmosDenom,
				Amount:   "1000000",
				Sender:   evmosAddress,
				Receiver: strideAddress,
				Memo:     fmt.Sprintf(`{ "autopilot": { "receiver": "%s" } }`, strideAddress),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
	}

	for i, tc := range testCases {
		s.Run(fmt.Sprintf("Case %d", i), func() {
			s.SetupTest()

			// Update the autopilot active flag
			s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{ClaimActive: tc.forwardingActive})

			// Set evmos airdrop
			airdrops := claimtypes.Params{
				Airdrops: []*claimtypes.Airdrop{{AirdropIdentifier: evmosAirdropId}},
			}
			err := s.App.ClaimKeeper.SetParams(s.Ctx, airdrops)
			s.Require().NoError(err, "no error expected when setting airdrop params")

			// Set claim records using key'd address
			oldClaimRecord := claimtypes.ClaimRecord{
				AirdropIdentifier: evmosAirdropId,
				Address:           evmosAddressKeyString,
				Weight:            sdk.NewDec(10),
				ActionCompleted:   []bool{false, false, false},
			}
			err = s.App.ClaimKeeper.SetClaimRecord(s.Ctx, oldClaimRecord)
			s.Require().NoError(err, "no error expected when setting claim record")

			// Store the expected new cliam record which should have the address changed
			expectedNewClaimRecord := oldClaimRecord
			expectedNewClaimRecord.Address = strideAddress

			// Replicate middleware stack
			transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
			autopilotStack := autopilot.NewIBCModule(s.App.AutopilotKeeper, transferIBCModule)

			// Call OnRecvPacket for autopilot
			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&tc.packetData)
			ack := autopilotStack.OnRecvPacket(
				s.Ctx,
				packet,
				sdk.AccAddress{},
			)

			if tc.transferShouldSucceed {
				s.Require().True(ack.Success(), "ack should be successful - ack: %+v", string(ack.Acknowledgement()))

				// Check funds were transferred

				if tc.airdropShouldUpdate {
					// Check that we have a new record for the user
					actualNewClaimRecord, err := s.App.ClaimKeeper.GetClaimRecord(s.Ctx, strideAccAddress, evmosAirdropId)
					s.Require().NoError(err, "no error expected when getting new claim record")
					s.Require().Equal(expectedNewClaimRecord, actualNewClaimRecord)

					// Check that the old record was removed (GetClaimRecord returns a zero-struct if not found)
					oldClaimRecord, _ := s.App.ClaimKeeper.GetClaimRecord(s.Ctx, evmosAddressKey, evmosAirdropId)
					s.Require().Equal("", oldClaimRecord.Address)
				} else {
					// If the airdrop code was never called, check that the old record claim record is still there
					oldClaimRecordAfterTransfer, err := s.App.ClaimKeeper.GetClaimRecord(s.Ctx, evmosAddressKey, evmosAirdropId)
					s.Require().NoError(err, "no error expected when getting old claim record")
					s.Require().Equal(oldClaimRecord, oldClaimRecordAfterTransfer)
				}
			} else {
				s.Require().False(ack.Success(), "ack should have failed - ack: %+v", string(ack.Acknowledgement()))
			}
		})
	}
}
