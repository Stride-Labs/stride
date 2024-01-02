package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	recordsmodule "github.com/Stride-Labs/stride/v16/x/records"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v16/x/autopilot"
	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v16/x/mint/types"
	recordstypes "github.com/Stride-Labs/stride/v16/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func getStakeibcPacketMetadata(address, action string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "action": "%[2]s" } 
			}
		}`, address, action)
}

// Helper function to mock out all the state needed to test autopilot liquid stake
// A transfer channel-0 is created, and the state is mocked out with an atom host zone
//
// Note: The testing framework is limited to one transfer channel per test, which is channel-0.
// If there's an outbound transfer, it must be on channel-0. So when testing a transfer along
// a non-host-zone channel, a different channel ID must be passed to this function
func (s *KeeperTestSuite) SetupAutopilotLiquidStake(
	featureEnabled bool,
	stakeAmount sdkmath.Int,
	strideToHostChannelId string,
	depositAddress sdk.AccAddress,
	liquidStaker sdk.AccAddress,
) {
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
		Amount:             sdk.ZeroInt(),
		HostZoneId:         HostChainId,
		Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
	})

	// Set the host zone - this should have the actual IBC denom
	prefixedDenom := transfertypes.GetPrefixedDenom(transfertypes.PortID, strideToHostChannelId, HostDenom)
	ibcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:           HostChainId,
		HostDenom:         HostDenom,
		RedemptionRate:    sdk.NewDec(1), // used to determine the stAmount
		DepositAddress:    depositAddress.String(),
		IbcDenom:          ibcDenom,
		TransferChannelId: strideToHostChannelId,
	})

	// Fund the staker with ibc/atom so they can liquid stake
	// (the function being tested is invoked downstream of the IBC transfer)
	s.FundAccount(liquidStaker, sdk.NewCoin(ibcDenom, stakeAmount))
}

func (s *KeeperTestSuite) CheckLiquidStakeSucceeded(
	liquidStakeAmount sdkmath.Int,
	nativeDenom string,
	staker sdk.AccAddress,
	depositAddress sdk.AccAddress,
	strideToHostChannelId string,
	expectedForwardChannelId string,
	originalReceiver string,
) {
	// If there was a forwarding step, the stTokens will end up in the escrow account
	// Otherwise, they'll be in the liquid staker's account
	stTokenRecipient := staker
	if expectedForwardChannelId != "" {
		escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, expectedForwardChannelId)
		stTokenRecipient = escrowAddress
	}

	prefixedNativeDenom := transfertypes.GetPrefixedDenom(transfertypes.PortID, strideToHostChannelId, HostDenom)
	nativeIBCDenom := transfertypes.ParseDenomTrace(prefixedNativeDenom).IBCDenom()

	// Confirm the liquid staker has lost his native tokens
	stakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, staker, nativeIBCDenom)
	s.Require().Zero(stakerBalance.Amount.Int64(), "liquid staker should have lost host tokens")

	// Confirm the deposit address now has the native tokens
	depositBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, nativeIBCDenom)
	s.Require().Equal(liquidStakeAmount.Int64(), depositBalance.Amount.Int64(), "deposit address should have gained host tokens")

	// Confirm the stToken's were minted and sent to the recipient
	recipientBalance := s.App.BankKeeper.GetBalance(s.Ctx, stTokenRecipient, "st"+nativeDenom)
	s.Require().Equal(liquidStakeAmount.Int64(), recipientBalance.Amount.Int64(), "st token recipient balance")

	// If there was a forwarding step, confirm the fallback address was stored
	if expectedForwardChannelId != "" {
		expectedFallbackAddress := originalReceiver
		address, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, expectedForwardChannelId, 1)
		s.Require().True(found, "fallback address should have been found")
		s.Require().Equal(expectedFallbackAddress, address, "fallback address")
	}
}

func (s *KeeperTestSuite) TestTryLiquidStake() {
	s.CreateTransferChannel(HostChainId)

	liquidStakerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	forwardRecipientOnHost := HostAddress

	stakeAmount := sdk.NewInt(1000000)

	// Building expected denom traces for the transfer packet data - this is all assuming the packet has been sent to stride
	// (the FungibleTokenPacketData has an denom trace for the Denom field, instead of an IBC hash)
	atom := "uatom"
	strd := "ustrd"
	osmo := "uosmo"
	denomTraces := map[string]string{
		// For host zone tokens, since stride is the first hop, there's no port/channel in the denom trace path
		atom: atom,
		osmo: osmo,
		// For strd, the hub's channel ID would have been appended to the denom trace
		strd: transfertypes.GetPrefixedDenom(transfertypes.PortID, ibctesting.FirstChannelID, strd),
	}

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
			liquidStakeDenom:  atom,
			liquidStakeAmount: stakeAmount.String(),
		},
		{
			// Liquid stake and forward, using the default host channel ID
			name:              "successful liquid stake and forward atom to the hub",
			enabled:           true,
			liquidStakeDenom:  atom,
			liquidStakeAmount: stakeAmount.String(),
			autopilotMetadata: types.StakeibcPacketMetadata{
				IbcReceiver: forwardRecipientOnHost,
			},
			expectedForwardChannelId: ibctesting.FirstChannelID, // default for host zone
		},
		{
			// Liquid stake and forward, using a custom channel ID
			// Host Zone Channel: channel-1, Outbound Transfer Channel: channel-0
			name:              "successful liquid stake and forward atom to osmo",
			enabled:           true,
			liquidStakeDenom:  atom,
			liquidStakeAmount: stakeAmount.String(),
			autopilotMetadata: types.StakeibcPacketMetadata{
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
			liquidStakeDenom:  atom,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "autopilot stakeibc routing is inactive",
		},
		{
			// Error caused an invalid amount in the packet
			name:              "invalid token amount",
			enabled:           true,
			liquidStakeDenom:  atom,
			liquidStakeAmount: "",
			expectedError:     "not a parsable amount field",
		},
		{
			// Error caused by the transfer of a non-native token
			// (i.e. a token that originated on stride)
			name:              "unable to liquid stake native token",
			enabled:           true,
			liquidStakeDenom:  strd,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "native token is not supported for liquid staking",
		},
		{
			// Error caused by the transfer of non-host zone token
			name:              "unable to liquid stake non-host zone token",
			enabled:           true,
			liquidStakeDenom:  osmo,
			liquidStakeAmount: stakeAmount.String(),
			expectedError:     "No HostZone for uosmo denom found",
		},
		{
			// Error caused by a mismatched IBC denom
			// Invoked by specifiying a different host zone channel ID
			name:                      "ibc denom does not match host zone",
			enabled:                   true,
			liquidStakeDenom:          atom,
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
			liquidStakeDenom:  atom,
			liquidStakeAmount: "-10000",
			expectedError:     "amount liquid staked must be positive and nonzero",
		},
		{
			// Error caused by a failed liquid stake
			// Invoked by trying to liquid stake more tokens than the staker has available
			name:              "failed to liquid stake",
			enabled:           true,
			liquidStakeDenom:  atom,
			liquidStakeAmount: stakeAmount.Add(sdkmath.NewInt(100000)).String(), // greater than balance
			expectedError:     "failed to liquid stake",
		},
		{
			// Failed to send transfer during forwarding step
			// Invoked by specifying a non-existent channel ID
			name:              "failed to forward transfer",
			enabled:           true,
			liquidStakeDenom:  atom,
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
				Denom:    denomTraces[tc.liquidStakeDenom],
				Amount:   tc.liquidStakeAmount,
				Receiver: liquidStakerOnStride.String(),
			}
			packet := channeltypes.Packet{
				SourcePort:         transfertypes.PortID,
				SourceChannel:      ibctesting.FirstChannelID,
				DestinationPort:    transfertypes.PortID,
				DestinationChannel: tc.inboundTransferChannnelId,
			}

			s.SetupTest()
			s.SetupAutopilotLiquidStake(tc.enabled, stakeAmount, tc.hostZoneChannelID, depositAddress, liquidStakerOnStride)

			err := s.App.AutopilotKeeper.TryLiquidStaking(s.Ctx, packet, transferMetadata, tc.autopilotMetadata)

			if tc.expectedError == "" {
				s.Require().NoError(err, "%s - no error expected when attempting liquid stake", tc.name)
				s.CheckLiquidStakeSucceeded(
					stakeAmount,
					tc.liquidStakeDenom,
					liquidStakerOnStride,
					depositAddress,
					tc.hostZoneChannelID,
					tc.expectedForwardChannelId,
					transferMetadata.Receiver,
				)
			} else {
				s.Require().ErrorContains(err, tc.expectedError, tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestOnRecvPacket_LiquidStake() {
	now := time.Now()

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

	atomHostDenom := "uatom"
	prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), atomHostDenom)
	atomIbcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
	prefixedDenom2 := transfertypes.GetPrefixedDenom(packet.GetDestPort(), "channel-1000", atomHostDenom)
	atomIbcDenom2 := transfertypes.ParseDenomTrace(prefixedDenom2).IBCDenom()

	strdDenom := "ustrd"
	prefixedDenom = transfertypes.GetPrefixedDenom(packet.GetSourcePort(), packet.GetSourceChannel(), strdDenom)
	strdIbcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	testCases := []struct {
		forwardingActive bool
		recvDenom        string
		packetData       transfertypes.FungibleTokenPacketData
		destChannel      string
		expSuccess       bool
		expLiquidStake   bool
	}{
		// successful liquid stake with metadata in receiver
		// successful liquid stake with metadata in the memo
		// successful liquid stake and forward to default host
		// successful liquid stake and forward to custom transfer channel

		// failed because param not enabled
		// failed because invalid stride address
		// failed because not a host denom
		// failed because transfer channel does not exist

		{ // params not enabled
			forwardingActive: false,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getStakeibcPacketMetadata(addr1.String(), "LiquidStake"),
				Memo:     "",
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     false,
			expLiquidStake: false,
		},
		{ // strd denom
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    strdIbcDenom,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getStakeibcPacketMetadata(addr1.String(), "LiquidStake"),
				Memo:     "",
			},
			destChannel:    "channel-0",
			recvDenom:      "ustrd",
			expSuccess:     false,
			expLiquidStake: false,
		},
		{ // all okay
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getStakeibcPacketMetadata(addr1.String(), "LiquidStake"),
				Memo:     "",
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     true,
			expLiquidStake: true,
		},
		{ // ibc denom uatom from different channel
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getStakeibcPacketMetadata(addr1.String(), "LiquidStake"),
				Memo:     "",
			},
			destChannel:    "channel-1000",
			recvDenom:      atomIbcDenom2,
			expSuccess:     false,
			expLiquidStake: false,
		},
		{ // all okay with memo liquidstaking since ibc-go v5.1.0
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: addr1.String(),
				Memo:     getStakeibcPacketMetadata(addr1.String(), "LiquidStake"),
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     true,
			expLiquidStake: true,
		},
		{ // all okay with no functional part
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: addr1.String(),
				Memo:     "",
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     true,
			expLiquidStake: false,
		},
		{ // invalid stride address (receiver)
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getStakeibcPacketMetadata("invalid_address", "LiquidStake"),
				Memo:     "",
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     false,
			expLiquidStake: false,
		},
		{ // invalid stride address (memo)
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: addr1.String(),
				Memo:     getStakeibcPacketMetadata("invalid_address", "LiquidStake"),
			},
			destChannel:    "channel-0",
			recvDenom:      atomIbcDenom,
			expSuccess:     false,
			expLiquidStake: false,
		},
	}

	for i, tc := range testCases {
		s.Run(fmt.Sprintf("Case %d", i), func() {
			packet.DestinationChannel = tc.destChannel
			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&tc.packetData)

			s.SetupTest() // reset
			ctx := s.Ctx

			s.App.AutopilotKeeper.SetParams(ctx, types.Params{StakeibcActive: tc.forwardingActive})

			// set epoch tracker for env
			s.App.StakeibcKeeper.SetEpochTracker(ctx, stakeibctypes.EpochTracker{
				EpochIdentifier:    epochtypes.STRIDE_EPOCH,
				EpochNumber:        1,
				NextEpochStartTime: uint64(now.Unix()),
				Duration:           43200,
			})
			// set deposit record for env
			s.App.RecordsKeeper.SetDepositRecord(ctx, recordstypes.DepositRecord{
				Id:                 1,
				Amount:             sdk.NewInt(100),
				Denom:              atomIbcDenom,
				HostZoneId:         "hub-1",
				Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
				DepositEpochNumber: 1,
				Source:             recordstypes.DepositRecord_STRIDE,
			})
			// set host zone for env
			s.App.StakeibcKeeper.SetHostZone(ctx, stakeibctypes.HostZone{
				ChainId:           "hub-1",
				ConnectionId:      "connection-0",
				Bech32Prefix:      "cosmos",
				TransferChannelId: "channel-0",
				IbcDenom:          atomIbcDenom,
				HostDenom:         atomHostDenom,
				RedemptionRate:    sdk.NewDec(1),
				DepositAddress:    addr1.String(),
			})

			// mint coins to be spent on liquid staking
			coins := sdk.Coins{sdk.NewInt64Coin(tc.recvDenom, 1000000)}
			err := s.App.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			s.Require().NoError(err)
			err = s.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, coins)
			s.Require().NoError(err)

			transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(s.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := autopilot.NewIBCModule(s.App.AutopilotKeeper, recordsStack)
			ack := routerIBCModule.OnRecvPacket(
				ctx,
				packet,
				addr1,
			)
			if tc.expSuccess {
				s.Require().True(ack.Success(), "ack should be successful - ack: %+v", string(ack.Acknowledgement()))

				// Check funds were transferred
				coin := s.App.BankKeeper.GetBalance(s.Ctx, addr1, tc.recvDenom)
				s.Require().Equal("2000000", coin.Amount.String(), "balance should have updated after successful transfer")

				// check minted balance for liquid staking
				allBalance := s.App.BankKeeper.GetAllBalances(ctx, addr1)
				liquidBalance := s.App.BankKeeper.GetBalance(ctx, addr1, "stuatom")
				if tc.expLiquidStake {
					s.Require().True(liquidBalance.Amount.IsPositive(), "liquid balance should be positive but was %s", allBalance.String())
				} else {
					s.Require().True(liquidBalance.Amount.IsZero(), "liquid balance should be zero but was %s", allBalance.String())
				}
			} else {
				s.Require().False(ack.Success(), "ack should have failed - ack: %+v", string(ack.Acknowledgement()))
			}
		})
	}
}
