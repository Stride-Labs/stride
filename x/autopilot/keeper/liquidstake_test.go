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

// Helper function to mock out all the state needed to test liquid stake
// The state is mocked out with an atom host zone
func (s *KeeperTestSuite) SetupAutopilotLiquidStake(
	featureEnabled bool,
	stakeAmount sdkmath.Int,
	strideToHostChannelId string,
	depositAddress sdk.AccAddress,
	liquidStaker sdk.AccAddress,
) {
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

func (s *KeeperTestSuite) TestTryLiquidStake() {
	s.CreateTransferChannel(HostChainId)

	liquidStakerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	forwardRecipientOnHost := HostAddress

	stakeAmount := sdk.NewInt(1000000)

	strideToHubChannel := "channel-0"
	hubToStrideChannel := "channel-1"
	strideToOsmoChannel := "channel-2"

	packet := channeltypes.Packet{
		SourcePort:         transfertypes.PortID,
		SourceChannel:      hubToStrideChannel,
		DestinationPort:    transfertypes.PortID,
		DestinationChannel: strideToHubChannel,
	}

	// Building expected denom traces for the transfer packet data - this is all assuming the packet has been sent to stride
	// (the FungibleTokenPacketData has an denom trace for the Denom field, instead of an IBC hash)

	// For host zone tokens, since stride is the first hop, there's no port/channel in the denom trace path
	atom := "uatom"
	atomTrace := atom

	// // For strd, the hub's channel ID would have been appended to the denom trace
	strd := "ustrd"
	strdTrace := transfertypes.GetPrefixedDenom(transfertypes.PortID, hubToStrideChannel, strd)

	// Osmo will have a valid denom but no host zone
	osmo := "uosmo"
	osmoTrace := osmo

	testCases := []struct {
		name              string
		enabled           bool
		liquidStakeDenom  string
		transferMetadata  transfertypes.FungibleTokenPacketData
		autopilotMetadata types.StakeibcPacketMetadata
		forwardChannelID  string // also used to dictate whether there's a forwarding step
		expectedError     string
	}{
		{
			// Normal autopilot liquid stake with no transfer
			name:             "successful liquid stake with atom",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
		},
		{
			// Liquid stake and forward, using the default host channel ID
			name:             "successful liquid stake and forward atom to the hub",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			autopilotMetadata: types.StakeibcPacketMetadata{
				IbcReceiver: forwardRecipientOnHost,
			},
			forwardChannelID: strideToHubChannel, // default for atom
		},
		{
			// Liquid stake and forward, using a custom channel ID
			name:             "successful liquid stake and forward atom to osmo",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			autopilotMetadata: types.StakeibcPacketMetadata{
				Action:          types.LiquidStake,
				IbcReceiver:     forwardRecipientOnHost,
				TransferChannel: strideToOsmoChannel,
			},
			forwardChannelID: strideToOsmoChannel, // determined by autopilot metadata
		},
		{
			// Error caused by autopilot disabled
			name:             "autopilot disabled",
			enabled:          false,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "autopilot stakeibc routing is inactive",
		},
		{
			// Error caused an invalid amount in the packet
			name:             "invalid token amount",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   "",
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "not a parsable amount field",
		},
		{
			// Error caused by the transfer of a non-native token
			// (i.e. a token that originated on stride)
			name:             "unable to liquid stake native token",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    strdTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "native token is not supported for liquid staking",
		},
		{
			// Error caused by the transfer of non-host zone token
			name:             "unable to liquid stake non-host zone token",
			enabled:          true,
			liquidStakeDenom: osmo,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    osmoTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "No HostZone for uosmo denom found",
		},
		{
			// Error caused by a mismatched IBC denom
			name:             "ibc denom does not match host zone",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "is not equal to host zone ibc denom",
		},
		{
			// Error caused by a failed validate basic before the liquid stake
			// Invoked by passing a negative amount
			name:             "failed liquid stake validate basic",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   "-10000",
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "amount liquid staked must be positive and nonzero",
		},
		{
			// Error caused by a failed liquid stake
			// Invoked by trying to liquid stake more tokens than the staker has available
			name:             "failed to liquid stake",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.Add(sdkmath.NewInt(100000)).String(), // greater than balance
				Receiver: liquidStakerOnStride.String(),
			},
			expectedError: "failed to liquid stake",
		},
		{
			// Failed to send transfer during forwarding step
			// Invoked by specifying a non-existent channel ID
			name:             "failed to forward transfer",
			enabled:          true,
			liquidStakeDenom: atom,
			transferMetadata: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   stakeAmount.String(),
				Receiver: liquidStakerOnStride.String(),
			},
			autopilotMetadata: types.StakeibcPacketMetadata{
				IbcReceiver:     forwardRecipientOnHost,
				TransferChannel: "channel-100", // does not exist
			},
			expectedError: "failed to submit transfer during autopilot liquid stake and forward",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.CreateTransferChannel(HostChainId)
			s.SetupAutopilotLiquidStake(tc.enabled, stakeAmount, strideToHubChannel, depositAddress, liquidStakerOnStride)

			err := s.App.AutopilotKeeper.TryLiquidStaking(s.Ctx, packet, tc.transferMetadata, tc.autopilotMetadata)
			if tc.expectedError == "" {
				s.Require().NoError(err, "%s - no error expected when attempting liquid stake", tc.name)

				// If there was a forwarding step, the stTokens will end up in the escrow account
				// Otherwise, they'll be in the liquid staker's account
				stTokenRecipient := liquidStakerOnStride
				if tc.forwardChannelID != "" {
					escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, tc.forwardChannelID)
					stTokenRecipient = escrowAddress
				}

				// Confirm the liquid staker has lost his native tokens
				liquidStakerBalance := s.App.BankKeeper.GetBalance(s.Ctx, liquidStakerOnStride, tc.liquidStakeDenom)
				s.Require().Zero(liquidStakerBalance.Amount.Int64(), "%s - liquid staker should have lost host tokens", tc.name)

				// Confirm the stToken's were minted and sent to the recipient
				recipientBalance := s.App.BankKeeper.GetBalance(s.Ctx, stTokenRecipient, "st"+tc.liquidStakeDenom)
				s.Require().Equal(stakeAmount.Int64(), recipientBalance.Amount.Int64(), "%s - st token recipient balance", tc.name)

				// If there was a forwarding step, confirm the fallback address was stored
				if tc.forwardChannelID != "" {
					expectedFallbackAddress := tc.transferMetadata.Receiver
					address, found := s.App.AutopilotKeeper.GetTransferFallbackAddress(s.Ctx, packet.DestinationChannel, 1)
					s.Require().True(found, "%s - fallback address should have been found", tc.name)
					s.Require().Equal(expectedFallbackAddress, address, "%s - fallback address", tc.name)
				}
			} else {
				s.Require().ErrorContains(err, tc.expectedError, tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLiquidStakeOnRecvPacket() {
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
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			packet.DestinationChannel = tc.destChannel
			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&tc.packetData)

			suite.SetupTest() // reset
			ctx := suite.Ctx

			suite.App.AutopilotKeeper.SetParams(ctx, types.Params{StakeibcActive: tc.forwardingActive})

			// set epoch tracker for env
			suite.App.StakeibcKeeper.SetEpochTracker(ctx, stakeibctypes.EpochTracker{
				EpochIdentifier:    epochtypes.STRIDE_EPOCH,
				EpochNumber:        1,
				NextEpochStartTime: uint64(now.Unix()),
				Duration:           43200,
			})
			// set deposit record for env
			suite.App.RecordsKeeper.SetDepositRecord(ctx, recordstypes.DepositRecord{
				Id:                 1,
				Amount:             sdk.NewInt(100),
				Denom:              atomIbcDenom,
				HostZoneId:         "hub-1",
				Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
				DepositEpochNumber: 1,
				Source:             recordstypes.DepositRecord_STRIDE,
			})
			// set host zone for env
			suite.App.StakeibcKeeper.SetHostZone(ctx, stakeibctypes.HostZone{
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
			err := suite.App.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			suite.Require().NoError(err)
			err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, coins)
			suite.Require().NoError(err)

			transferIBCModule := transfer.NewIBCModule(suite.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(suite.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := autopilot.NewIBCModule(suite.App.AutopilotKeeper, recordsStack)
			ack := routerIBCModule.OnRecvPacket(
				ctx,
				packet,
				addr1,
			)
			if tc.expSuccess {
				suite.Require().True(ack.Success(), "ack should be successful - ack: %+v", string(ack.Acknowledgement()))

				// Check funds were transferred
				coin := suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, tc.recvDenom)
				suite.Require().Equal("2000000", coin.Amount.String(), "balance should have updated after successful transfer")

				// check minted balance for liquid staking
				allBalance := suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				liquidBalance := suite.App.BankKeeper.GetBalance(ctx, addr1, "stuatom")
				if tc.expLiquidStake {
					suite.Require().True(liquidBalance.Amount.IsPositive(), "liquid balance should be positive but was %s", allBalance.String())
				} else {
					suite.Require().True(liquidBalance.Amount.IsZero(), "liquid balance should be zero but was %s", allBalance.String())
				}
			} else {
				suite.Require().False(ack.Success(), "ack should have failed - ack: %+v", string(ack.Acknowledgement()))
			}
		})
	}
}
