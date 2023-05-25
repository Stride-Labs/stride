package keeper_test

import (
	"fmt"
	"strings"

	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/autopilot"
	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
	claimtypes "github.com/Stride-Labs/stride/v9/x/claim/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// TODO: Separate out tests cases that are not necessarily Claim or Stakeibc related,
// but more just test the parsing that occurs in OnRecvPacket
// Move them to a different test file

var EvmosChainId = "evmos-1"

func getClaimPacketMetadata(address string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"claim": { "stride_address": "%[1]s" } 
			}
		}`, address)
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
	packetTemplate := channeltypes.Packet{
		Sequence:         1,
		SourcePort:       transfertypes.PortID,
		SourceChannel:    ibctesting.FirstChannelID,
		DestinationPort:  transfertypes.PortID,
		Data:             []byte{},
		TimeoutHeight:    clienttypes.Height{},
		TimeoutTimestamp: 0,
	}
	packetDataTemplate := transfertypes.FungibleTokenPacketData{
		Denom:  evmosDenom,
		Amount: "1000000",
		Sender: evmosAddress,
	}

	// To test the case where the packet has a valid channel but for a host zone without an airdrop
	channelIdForDifferentHostZone := "channel-1"

	testCases := []struct {
		name                         string
		autopilotClaimActive         bool
		autopilotClaimEnabledForHost bool
		packetData                   transfertypes.FungibleTokenPacketData
		destinationChannelID         string
		destinationPortID            string
		transferShouldSucceed        bool
		airdropShouldUpdate          bool
	}{
		{
			name:                         "successful airdrop update from receiver",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   true,
		},
		{
			name:                         "successful airdrop update from memo",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     getClaimPacketMetadata(strideAddress),
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   true,
		},
		{
			name:                         "memo receiver overrides original receiver field",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: "address-will-get-overriden",
				Memo:     getClaimPacketMetadata(strideAddress),
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   true,
		},
		{
			name:                         "valid receiver routing schema, but routing inactive",
			autopilotClaimActive:         false,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "valid receiver routing schema, but routing inactive for airdrop",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: false,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "valid memo routing schema, but routing inactive",
			autopilotClaimActive:         false,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "invalid stride address",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     getClaimPacketMetadata("invalid_address"),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "normal transfer packet - no memo",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     "",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "normal transfer packet - empty JSON memo",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     "{}",
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "normal transfer packet - different middleware",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     `{ "other_module": { } }`,
			},
			transferShouldSucceed: true,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "invalid autopilot JSON - no receiver",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     `{ "autopilot": {} }`,
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "invalid autopilot JSON - no routing module",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     fmt.Sprintf(`{ "autopilot": { "receiver": "%s" } }`, strideAddress),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "memo too long",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strideAddress,
				Memo:     strings.Repeat("X", 300),
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "receiver too long",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: strings.Repeat("X", 300),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "not transfer channel",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         ibctesting.FirstChannelID,
			destinationPortID:            "invalid_port",
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "transfer channel from a different host zone",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         channelIdForDifferentHostZone,
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
		{
			name:                         "transfer channel does not exist",
			autopilotClaimActive:         true,
			autopilotClaimEnabledForHost: true,
			destinationChannelID:         "channel-XXX",
			destinationPortID:            transfertypes.PortID,
			packetData: transfertypes.FungibleTokenPacketData{
				Receiver: getClaimPacketMetadata(strideAddress),
				Memo:     "",
			},
			transferShouldSucceed: false,
			airdropShouldUpdate:   false,
		},
	}

	for i, tc := range testCases {
		s.Run(fmt.Sprintf("Case %d_%s", i, tc.name), func() {
			s.SetupTest()

			// Update the autopilot active flag
			s.App.AutopilotKeeper.SetParams(s.Ctx, types.Params{
				ClaimActive: tc.autopilotClaimActive,
			})

			// Set evmos airdrop
			airdrops := claimtypes.Params{
				Airdrops: []*claimtypes.Airdrop{{
					AirdropIdentifier: evmosAirdropId,
					ChainId:           EvmosChainId,
					AutopilotEnabled:  tc.autopilotClaimEnabledForHost,
				}},
			}
			err := s.App.ClaimKeeper.SetParams(s.Ctx, airdrops)
			s.Require().NoError(err, "no error expected when setting airdrop params")

			// Store the host zone so that we can verify the channel
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
				ChainId:           EvmosChainId,
				TransferChannelId: ibctesting.FirstChannelID,
			})

			// Store a second host zone that does not have an airdrop
			s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
				ChainId:           "differnet_host_zone",
				TransferChannelId: channelIdForDifferentHostZone,
			})

			// Set claim records using key'd address
			oldClaimRecord := claimtypes.ClaimRecord{
				AirdropIdentifier: evmosAirdropId,
				Address:           evmosAddressKeyString,
				Weight:            sdk.NewDec(10),
				ActionCompleted:   []bool{false, false, false},
			}
			err = s.App.ClaimKeeper.SetClaimRecord(s.Ctx, oldClaimRecord)
			s.Require().NoError(err, "no error expected when setting claim record")

			// Store the expected new claim record which should have the address changed
			expectedNewClaimRecord := oldClaimRecord
			expectedNewClaimRecord.Address = strideAddress

			// Replicate middleware stack
			transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
			autopilotStack := autopilot.NewIBCModule(s.App.AutopilotKeeper, transferIBCModule)

			// Update packet and packet data
			packetData := packetDataTemplate
			packetData.Memo = tc.packetData.Memo
			packetData.Receiver = tc.packetData.Receiver

			packet := packetTemplate
			packet.DestinationChannel = tc.destinationChannelID
			packet.DestinationPort = tc.destinationPortID
			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&packetData)

			// Build the evmos denom from the packet
			prefixedDenom := transfertypes.GetPrefixedDenom(packet.DestinationPort, packet.DestinationChannel, evmosDenom)
			evmosIbcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

			// Call OnRecvPacket for autopilot
			ack := autopilotStack.OnRecvPacket(
				s.Ctx,
				packet,
				sdk.AccAddress{},
			)

			if tc.transferShouldSucceed {
				s.Require().True(ack.Success(), "ack should be successful - ack: %+v", string(ack.Acknowledgement()))

				// Check funds were transferred
				coin := s.App.BankKeeper.GetBalance(s.Ctx, sdk.MustAccAddressFromBech32(strideAddress), evmosIbcDenom)
				s.Require().Equal(packetDataTemplate.Amount, coin.Amount.String(), "balance should have updated after successful transfer")

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
