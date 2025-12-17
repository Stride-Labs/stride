package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/Stride-Labs/stride/v31/x/autopilot"
	"github.com/Stride-Labs/stride/v31/x/autopilot/types"
	epochtypes "github.com/Stride-Labs/stride/v31/x/epochs/types"
	recordsmodule "github.com/Stride-Labs/stride/v31/x/records"
	recordstypes "github.com/Stride-Labs/stride/v31/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

var (
	// Arbitrary channel ID on the non-stride zone
	SourceChannelOnHost = "channel-1000"

	// Building a mapping to of base denom to expected denom traces for the transfer packet data
	// This is all assuming the packet has been sent to stride (the FungibleTokenPacketData has
	// a denom-trace for the Denom field, instead of an IBC hash)
	ReceivePacketDenomTraces = map[string]string{
		// For host zone tokens, since stride is the first hop, there's no port/channel in the denom trace path
		Atom: Atom,
		Osmo: Osmo,
		// For strd, the other zone's channel ID is appended to the denom trace
		Strd: transfertypes.GetPrefixedDenom(transfertypes.PortID, SourceChannelOnHost, Strd),
	}
)

// Helper function to create the autopilot JSON payload for a liquid stake
func getLiquidStakePacketMetadata(receiver, ibcReceiver, transferChannelId string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "action": "LiquidStake", "ibc_receiver": "%[2]s", "transfer_channel": "%[3]s" } 
			}
		}`, receiver, ibcReceiver, transferChannelId)
}

// Helper function to mock out all the state needed to test autopilot liquid stake
// A transfer channel-0 is created, and the state is mocked out with an atom host zone
//
// Note: The testing framework is limited to one transfer channel per test, which is channel-0.
// If there's an outbound transfer, it must be on channel-0. So when testing a transfer along
// a non-host-zone channel (e.g. a transfer of statom to Osmosis), a different `strideToHostChannelId`
// channel ID must be passed to this function
//
// Returns the ibc denom of the native token
func (s *KeeperTestSuite) SetupAutopilotLiquidStake(
	featureEnabled bool,
	strideToHostChannelId string,
	depositAddress sdk.AccAddress,
	liquidStaker sdk.AccAddress,
) (nativeTokenIBCDenom string) {
	// Create a transfer channel on channel-0 for the outbound transfer
	// Note: We pass a dummy chain ID cause all that matters here is
	// that channel-0 exists, it does not have to line up with the host zone
	s.CreateTransferChannel("chain-0")

	// Set whether the feature is active
	params := s.App.AutopilotKeeper.GetParams(s.Ctx)
	params.StakeibcActive = featureEnabled
	s.App.AutopilotKeeper.SetParams(s.Ctx, params)

	// Set the epoch tracker to lookup the deposit record
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.STRIDE_EPOCH,
		EpochNumber:     1,
	})

	// Set deposit record to store the new liquid stake
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordstypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		Amount:             sdkmath.ZeroInt(),
		HostZoneId:         HostChainId,
		Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
	})

	// Set the host zone - this should have the actual IBC denom
	prefixedDenom := transfertypes.GetPrefixedDenom(transfertypes.PortID, strideToHostChannelId, HostDenom)
	nativeTokenIBCDenom = transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         HostDenom,
		RedemptionRate:    sdkmath.LegacyNewDec(1), // used to determine the stAmount
		DepositAddress:    depositAddress.String(),
		IbcDenom:          nativeTokenIBCDenom,
		TransferChannelId: strideToHostChannelId,
	})

	return nativeTokenIBCDenom
}

func (s *KeeperTestSuite) CheckLiquidStakeSucceeded(
	liquidStakeAmount sdkmath.Int,
	liquidStakerAddress sdk.AccAddress,
	depositAddress sdk.AccAddress,
	nativeTokenIBCDenom string,
	expectedForwardChannelId string,
) {
	// If there was a forwarding step, the stTokens will end up in the escrow account
	// Otherwise, they'll be in the liquid staker's account
	stTokenRecipient := liquidStakerAddress
	if expectedForwardChannelId != "" {
		escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, expectedForwardChannelId)
		stTokenRecipient = escrowAddress
	}

	// Confirm the liquid staker has lost his native tokens
	stakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, liquidStakerAddress, nativeTokenIBCDenom)
	s.Require().Zero(stakerBalance.Amount.Int64(), "liquid staker should have lost host tokens")

	// Confirm the deposit address now has the native tokens
	depositBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, nativeTokenIBCDenom)
	s.Require().Equal(liquidStakeAmount.Int64(), depositBalance.Amount.Int64(), "deposit address should have gained host tokens")

	// Confirm the stToken's were minted and sent to the recipient
	recipientBalance := s.App.BankKeeper.GetBalance(s.Ctx, stTokenRecipient, "st"+HostDenom)
	s.Require().Equal(liquidStakeAmount.Int64(), recipientBalance.Amount.Int64(), "st token recipient balance")

	// If there was a forwarding step, confirm the fallback address was stored
	// The fallback address in all these tests is the same as the liquid staker
	if expectedForwardChannelId != "" {
		address, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, expectedForwardChannelId, 1)
		s.Require().True(found, "fallback address should have been found")
		s.Require().Equal(liquidStakerAddress.String(), address, "fallback address")
	}
}

// Tests TryLiquidStake directly - beginning after the inbound autopilot transfer has passed down the stack
func (s *KeeperTestSuite) TestTryLiquidStake() {
	liquidStakerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	forwardRecipientOnHost := HostAddress

	stakeAmount := sdkmath.NewInt(1000000)

	testCases := []struct {
		name                      string
		enabled                   bool
		liquidStakeDenom          string
		liquidStakeAmount         string
		autopilotMetadata         types.StakeibcPacketMetadata
		hostZoneChannelID         string // defaults to channel-0 if not specified
		inboundTransferChannnelId string // defaults to channel-0 if not specified
		expectedForwardChannelId  string // defaults to empty (no forwarding)
		expectedError             string
	}{
		{
			// Normal autopilot liquid stake with no transfer
			name:              "successful liquid stake with atom",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.String(),
		},
		{
			// Liquid stake and forward, using the default host channel ID
			name:              "successful liquid stake and forward atom to the hub",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.String(),
			autopilotMetadata: types.StakeibcPacketMetadata{
				StrideAddress: liquidStakerOnStride.String(), // fallback address
				IbcReceiver:   forwardRecipientOnHost,
			},
			expectedForwardChannelId: ibctesting.FirstChannelID, // default for host zone
		},
		{
			// Liquid stake and forward, using a custom channel ID
			// Host Zone Channel: channel-1, Outbound Transfer Channel: channel-0
			name:              "successful liquid stake and forward atom to osmo",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.String(),
			autopilotMetadata: types.StakeibcPacketMetadata{
				StrideAddress:   liquidStakerOnStride.String(), // fallback address
				IbcReceiver:     forwardRecipientOnHost,
				TransferChannel: "channel-0", // custom channel (different than host channel below)
			},
			inboundTransferChannnelId: "channel-1",
			hostZoneChannelID:         "channel-1",
			expectedForwardChannelId:  "channel-0",
		},
		{
			// Error caused by autopilot disabled
			name:              "autopilot disabled",
			enabled:           false,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "autopilot stakeibc routing is inactive",
		},
		{
			// Error caused an invalid amount in the packet
			name:              "invalid token amount",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: "",
			expectedError:     "not a parsable amount field",
		},
		{
			// Error caused by the transfer of a non-native token
			// (i.e. a token that originated on stride)
			name:              "unable to liquid stake native token",
			enabled:           true,
			liquidStakeDenom:  Strd,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "native token is not supported for liquid staking",
		},
		{
			// Error caused by the transfer of non-host zone token
			name:              "unable to liquid stake non-host zone token",
			enabled:           true,
			liquidStakeDenom:  Osmo,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "No HostZone for uosmo denom found",
		},
		{
			// Error caused by a mismatched IBC denom
			// Invoked by specifiying a different host zone channel ID
			name:                      "ibc denom does not match host zone",
			enabled:                   true,
			liquidStakeDenom:          Atom,
			liquidStakeAmount:         stakeAmount.String(),
			hostZoneChannelID:         "channel-0",
			inboundTransferChannnelId: "channel-1", // Different than host zone
			expectedError:             "is not equal to host zone ibc denom",
		},
		{
			// Error caused by a failed validate basic before the liquid stake
			// Invoked by passing a negative amount
			name:              "failed liquid stake validate basic",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: "-10000",
			expectedError:     "amount liquid staked must be positive and nonzero",
		},
		{
			// Error caused by a failed liquid stake
			// Invoked by trying to liquid stake more tokens than the staker has available
			name:              "failed to liquid stake",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.Add(sdkmath.NewInt(100000)).String(), // greater than balance
			expectedError:     "failed to liquid stake",
		},
		{
			// Failed to send transfer during forwarding step
			// Invoked by specifying a non-existent channel ID
			name:              "failed to forward transfer",
			enabled:           true,
			liquidStakeDenom:  Atom,
			liquidStakeAmount: stakeAmount.String(),
			autopilotMetadata: types.StakeibcPacketMetadata{
				IbcReceiver:     forwardRecipientOnHost,
				TransferChannel: "channel-100", // does not exist
			},
			expectedError: "failed to submit transfer during autopilot liquid stake and forward",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Fill in the default channel ID's if they weren't specified
			if tc.hostZoneChannelID == "" {
				tc.hostZoneChannelID = ibctesting.FirstChannelID
			}
			if tc.inboundTransferChannnelId == "" {
				tc.inboundTransferChannnelId = ibctesting.FirstChannelID
			}

			transferMetadata := transfertypes.FungibleTokenPacketData{
				Denom:    ReceivePacketDenomTraces[tc.liquidStakeDenom],
				Amount:   tc.liquidStakeAmount,
				Receiver: liquidStakerOnStride.String(),
			}
			packet := channeltypes.Packet{
				SourcePort:         transfertypes.PortID,
				SourceChannel:      SourceChannelOnHost,
				DestinationPort:    transfertypes.PortID,
				DestinationChannel: tc.inboundTransferChannnelId,
			}

			s.SetupTest()
			nativeTokenIBCDenom := s.SetupAutopilotLiquidStake(tc.enabled, tc.hostZoneChannelID, depositAddress, liquidStakerOnStride)

			// Since this tested function is normally downstream of the inbound IBC transfer,
			// we have to fund the staker with ibc/atom before calling this function so
			// they can liquid stake
			s.FundAccount(liquidStakerOnStride, sdk.NewCoin(nativeTokenIBCDenom, stakeAmount))

			err := s.App.AutopilotKeeper.TryLiquidStaking(s.Ctx, packet, transferMetadata, tc.autopilotMetadata)

			if tc.expectedError == "" {
				s.Require().NoError(err, "%s - no error expected when attempting liquid stake", tc.name)
				s.CheckLiquidStakeSucceeded(
					stakeAmount,
					liquidStakerOnStride,
					depositAddress,
					nativeTokenIBCDenom,
					tc.expectedForwardChannelId,
				)
			} else {
				s.Require().ErrorContains(err, tc.expectedError, tc.name)
			}
		})
	}
}

// Tests the full OnRecvPacket callback, with liquid staking specific test cases
func (s *KeeperTestSuite) TestOnRecvPacket_LiquidStake() {
	liquidStakerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	differentAddress := s.TestAccs[2].String()
	forwardRecipientOnHost := HostAddress

	stakeAmount := sdkmath.NewInt(1000000)

	testCases := []struct {
		name                      string
		enabled                   bool
		liquidStakeDenom          string
		transferReceiver          string
		transferMemo              string
		hostZoneChannelID         string // defaults to channel-0 if not specified
		inboundTransferChannnelId string // defaults to channel-0 if not specified
		expectedForwardChannelId  string // defaults to empty (no forwarding)
		expectedSuccess           bool
		expectedLiquidStake       bool
	}{
		{
			name:                "successful liquid stake",
			enabled:             true,
			liquidStakeDenom:    Atom,
			transferReceiver:    liquidStakerOnStride.String(),
			transferMemo:        getLiquidStakePacketMetadata(liquidStakerOnStride.String(), "", ""),
			expectedSuccess:     true,
			expectedLiquidStake: true,
		},
		{
			name:                     "successful liquid stake and forward to default host",
			enabled:                  true,
			liquidStakeDenom:         Atom,
			transferReceiver:         liquidStakerOnStride.String(),
			transferMemo:             getLiquidStakePacketMetadata(liquidStakerOnStride.String(), forwardRecipientOnHost, ""),
			expectedForwardChannelId: ibctesting.FirstChannelID,
			expectedSuccess:          true,
			expectedLiquidStake:      true,
		},
		{
			name:                      "successful liquid stake and forward to custom transfer channel",
			enabled:                   true,
			liquidStakeDenom:          Atom,
			transferReceiver:          liquidStakerOnStride.String(),
			transferMemo:              getLiquidStakePacketMetadata(liquidStakerOnStride.String(), forwardRecipientOnHost, "channel-0"),
			hostZoneChannelID:         "channel-1",
			inboundTransferChannnelId: "channel-1",
			expectedForwardChannelId:  "channel-0", // different than host zone, specified in memo
			expectedSuccess:           true,
			expectedLiquidStake:       true,
		},
		{
			name:                "normal transfer with no liquid stake",
			enabled:             true,
			liquidStakeDenom:    Atom,
			transferReceiver:    liquidStakerOnStride.String(),
			transferMemo:        "",
			expectedSuccess:     true,
			expectedLiquidStake: false,
		},
		{
			name:             "autopilot disabled",
			enabled:          false,
			liquidStakeDenom: Atom,
			transferReceiver: liquidStakerOnStride.String(),
			transferMemo:     getLiquidStakePacketMetadata(liquidStakerOnStride.String(), "", ""),
			expectedSuccess:  false,
		},
		{
			name:             "invalid stride address (receiver)",
			enabled:          true,
			liquidStakeDenom: Osmo,
			transferReceiver: getLiquidStakePacketMetadata("XXX", "", ""),
			transferMemo:     "",
			expectedSuccess:  false,
		},
		{
			name:             "invalid stride address (memo)",
			enabled:          true,
			liquidStakeDenom: Osmo,
			transferReceiver: liquidStakerOnStride.String(),
			transferMemo:     getLiquidStakePacketMetadata("XXX", "", ""),
			expectedSuccess:  false,
		},
		{
			name:             "memo and transfer address mismatch",
			enabled:          true,
			liquidStakeDenom: Osmo,
			transferReceiver: liquidStakerOnStride.String(),
			transferMemo:     getLiquidStakePacketMetadata(differentAddress, "", ""),
			expectedSuccess:  false,
		},
		{
			name:             "not host denom",
			enabled:          true,
			liquidStakeDenom: Osmo,
			transferReceiver: liquidStakerOnStride.String(),
			transferMemo:     getLiquidStakePacketMetadata(liquidStakerOnStride.String(), "", ""),
			expectedSuccess:  false,
		},
		{
			name:             "failed to outbound transfer",
			enabled:          true,
			liquidStakeDenom: Atom,
			transferReceiver: liquidStakerOnStride.String(),
			transferMemo:     getLiquidStakePacketMetadata(liquidStakerOnStride.String(), forwardRecipientOnHost, "channel-999"), // channel DNE
			expectedSuccess:  false,
		},
		{
			name:                      "valid uatom token from invalid channel",
			enabled:                   true,
			liquidStakeDenom:          Atom,
			transferReceiver:          liquidStakerOnStride.String(),
			transferMemo:              getLiquidStakePacketMetadata(liquidStakerOnStride.String(), "", ""),
			hostZoneChannelID:         "channel-0",
			inboundTransferChannnelId: "channel-999", // channel DNE
			expectedSuccess:           false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // reset

			// Fill in the default channel ID's if they weren't specified
			if tc.hostZoneChannelID == "" {
				tc.hostZoneChannelID = ibctesting.FirstChannelID
			}
			if tc.inboundTransferChannnelId == "" {
				tc.inboundTransferChannnelId = ibctesting.FirstChannelID
			}

			transferMetadata := transfertypes.FungibleTokenPacketData{
				Sender:   HostAddress,
				Receiver: tc.transferReceiver,
				Denom:    ReceivePacketDenomTraces[tc.liquidStakeDenom],
				Amount:   stakeAmount.String(),
				Memo:     tc.transferMemo,
			}
			packet := channeltypes.Packet{
				SourcePort:         transfertypes.PortID,
				SourceChannel:      SourceChannelOnHost,
				DestinationPort:    transfertypes.PortID,
				DestinationChannel: tc.inboundTransferChannnelId,
				Data:               transfertypes.ModuleCdc.MustMarshalJSON(&transferMetadata),
			}

			nativeTokenIBCDenom := s.SetupAutopilotLiquidStake(tc.enabled, tc.hostZoneChannelID, depositAddress, liquidStakerOnStride)

			transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(s.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := autopilot.NewIBCModule(s.App.AutopilotKeeper, recordsStack)
			ack := routerIBCModule.OnRecvPacket(
				s.Ctx,
				packet,
				s.TestAccs[2], // arbitrary relayer address - not actually used
			)

			if tc.expectedSuccess {
				s.Require().True(ack.Success(), "ack should be successful - ack: %+v", string(ack.Acknowledgement()))

				if tc.expectedLiquidStake {
					s.CheckLiquidStakeSucceeded(
						stakeAmount,
						liquidStakerOnStride,
						depositAddress,
						nativeTokenIBCDenom,
						tc.expectedForwardChannelId,
					)
				}
			} else {
				s.Require().False(ack.Success(), "ack should have failed - ack: %+v", string(ack.Acknowledgement()))
			}
		})
	}
}
