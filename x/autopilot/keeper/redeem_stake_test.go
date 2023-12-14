package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	recordsmodule "github.com/Stride-Labs/stride/v16/x/records"

	sdk "github.com/cosmos/cosmos-sdk/types"

	router "github.com/Stride-Labs/stride/v16/x/autopilot"
	"github.com/Stride-Labs/stride/v16/x/autopilot/types"
	epochtypes "github.com/Stride-Labs/stride/v16/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v16/x/mint/types"
	recordstypes "github.com/Stride-Labs/stride/v16/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v16/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func getRedeemStakeStakeibcPacketMetadata(address, ibcReceiver, transferChannel string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "action": "RedeemStake", "ibc_receiver": "%[2]s", "transfer_channel": "%[3]s" } 
			}
		}`, address, ibcReceiver, transferChannel)
}

// Helper function to mock out all the state needed to test redeem stake
// The state is mocked out with an atom host zone
func (s *KeeperTestSuite) SetupAutopilotRedeemStake(featureEnabled bool, redeemAmount sdkmath.Int, depositAddress, redeemerOnStride sdk.AccAddress) {
	// Set whether the feature is active
	params := s.App.AutopilotKeeper.GetParams(s.Ctx)
	params.StakeibcActive = featureEnabled
	s.App.AutopilotKeeper.SetParams(s.Ctx, params)

	// set epoch tracker to look up epoch unbonding record
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     1,
	})

	// set epoch unbonding record which will store the new user redemption record
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordstypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordstypes.HostZoneUnbonding{
			{
				HostZoneId:            HostChainId,
				UserRedemptionRecords: []string{},
				NativeTokenAmount:     sdk.NewInt(1000000),
			},
		},
	})

	// store the host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          HostChainId,
		Bech32Prefix:     HostBechPrefix, // required to validate claim receiver
		HostDenom:        HostDenom,
		RedemptionRate:   sdk.NewDec(1), // used to determine native token amount
		DepositAddress:   depositAddress.String(),
		TotalDelegations: redeemAmount, // there must be enough stake to cover the redemption
	})

	// fund the user with sttokens so they can redeem
	// (the function being tested is invoked downstream of the IBC transfer)
	s.FundAccount(redeemerOnStride, sdk.NewCoin("st"+HostDenom, redeemAmount))
}

func (s *KeeperTestSuite) TestTryRedeemStake() {
	redeemerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	redeemerOnHost := HostAddress

	redeemAmount := sdk.NewInt(1000000)

	strideToHubChannel := "channel-0"
	hubToStrideChannel := "channel-1"

	packet := channeltypes.Packet{
		SourcePort:         transfertypes.PortID,
		SourceChannel:      hubToStrideChannel,
		DestinationPort:    transfertypes.PortID,
		DestinationChannel: strideToHubChannel,
	}

	// Building on expected denom's in the packet data below - this is all assuming the packet has been sent to stride
	// For host zone tokens, since stride is the first hop, there's no port/channel in the denom trace path
	atom := "uatom"
	atomTrace := atom

	// For stride, the hub's channel ID would have been appended to the denom trace
	strd := "ustrd"
	strdTrace := transfertypes.GetPrefixedDenom(transfertypes.PortID, hubToStrideChannel, strd)

	// Similarly for stTokens, the hub's channel ID would be appended
	stAtom := "stuatom"
	stAtomTrace := transfertypes.GetPrefixedDenom(transfertypes.PortID, hubToStrideChannel, stAtom)

	// StOsmo will have a valid denom but no host zone
	stOsmo := "stuosmo"
	stOsmoTrace := transfertypes.GetPrefixedDenom(transfertypes.PortID, hubToStrideChannel, stOsmo)

	testCases := []struct {
		name           string
		enabled        bool
		redeemDenom    string
		packetData     transfertypes.FungibleTokenPacketData
		packetMetadata types.StakeibcPacketMetadata
		expectedError  string
	}{
		{
			name:        "successful redemption with stuatom",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
		},
		{
			name:        "forwarding inactive",
			enabled:     false,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "packet forwarding param is not active",
		},
		{
			name:        "failed redemption with atom",
			enabled:     true,
			redeemDenom: atom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    atomTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "the ibc token uatom is not supported for redeem stake",
		},
		{
			name:        "failed redemption with ustrd",
			enabled:     true,
			redeemDenom: strd,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    strdTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "not a liquid staking token",
		},
		{
			name:        "failed to parse amount",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   "XXX",
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "not a parsable amount field",
		},
		{
			name:        "negative amount",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   "-1000",
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "not a parsable amount field",
		},
		{
			name:        "not a host zone denom",
			enabled:     true,
			redeemDenom: stOsmo,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stOsmoTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "No HostZone for uosmo found",
		},
		{
			name:        "invalid stride address",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Receiver: "",
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "invalid creator address",
		},
		{
			name:        "invalid claim receiver",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: "",
			},
			expectedError: "receiver cannot be empty",
		},
		{
			name:        "redeem msg failed",
			enabled:     true,
			redeemDenom: stAtom,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   "100000000000000", // amount is too large - causes failure
				Receiver: redeemerOnStride.String(),
			},
			packetMetadata: types.StakeibcPacketMetadata{
				Action:      types.RedeemStake,
				IbcReceiver: redeemerOnHost,
			},
			expectedError: "redeem stake failed",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupAutopilotRedeemStake(tc.enabled, redeemAmount, depositAddress, redeemerOnStride)

			err := s.App.AutopilotKeeper.TryRedeemStake(s.Ctx, packet, tc.packetData, tc.packetMetadata)
			if tc.expectedError == "" {
				s.Require().NoError(err, "%s - no error expected when attempting redeem stake", tc.name)

				// check if redeem record is created
				hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, 1, HostChainId)
				s.Require().True(found)
				s.Require().True(len(hostZoneUnbonding.UserRedemptionRecords) > 0,
					"%s - user redemption record should have been created", tc.name)

				// check that tokens were escrowed
				escrowBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, tc.redeemDenom)
				s.Require().Equal(redeemAmount.Int64(), escrowBalance.Amount.Int64(), "%s - tokens should have been escrowed", tc.name)
			} else {
				s.Require().ErrorContains(err, tc.expectedError, tc.name)
			}
		})
	}
}

// TODO: Move to ibc_test.go when OnRecvPacket is moved
func (suite *KeeperTestSuite) TestOnRecvPacket_RedeemStake() {
	now := time.Now()

	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         "transfer",
		SourceChannel:      "channel-0",
		DestinationPort:    "transfer",
		DestinationChannel: "channel-0",
	}

	atomHostDenom := "uatom"
	prefixedDenom := transfertypes.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), atomHostDenom)
	atomIbcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	strdDenom := "ustrd"
	prefixedDenom = transfertypes.GetPrefixedDenom(packet.GetSourcePort(), packet.GetSourceChannel(), strdDenom)
	strdIbcDenom := transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	stAtomDenom := "stuatom"
	prefixedDenom = transfertypes.GetPrefixedDenom(packet.GetSourcePort(), packet.GetSourceChannel(), stAtomDenom)
	stAtomFullDenomPath := transfertypes.ParseDenomTrace(prefixedDenom).GetFullDenomPath()

	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	testCases := []struct {
		forwardingActive bool
		recvDenom        string
		packetData       transfertypes.FungibleTokenPacketData
		expSuccess       bool
		expRedeemStake   bool
	}{
		{ // params not enabled
			forwardingActive: false,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomFullDenomPath,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata(addr1.String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", ""),
				Memo:     "",
			},
			recvDenom:      stAtomDenom,
			expSuccess:     false,
			expRedeemStake: false,
		},
		{ // strd denom
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    strdIbcDenom,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata(addr1.String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", ""),
				Memo:     "",
			},
			recvDenom:      "ustrd",
			expSuccess:     false,
			expRedeemStake: false,
		},
		{ // all okay
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomFullDenomPath,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata(addr1.String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", ""),
				Memo:     "",
			},
			recvDenom:      stAtomDenom,
			expSuccess:     true,
			expRedeemStake: true,
		},
		{ // all okay with memo liquidstaking since ibc-go v5.1.0
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomFullDenomPath,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: addr1.String(),
				Memo:     getRedeemStakeStakeibcPacketMetadata(addr1.String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", ""),
			},
			recvDenom:      stAtomDenom,
			expSuccess:     true,
			expRedeemStake: true,
		},
		{ // invalid receiver
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomFullDenomPath,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata("xxx", "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", ""),
				Memo:     "",
			},
			recvDenom:      stAtomDenom,
			expSuccess:     false,
			expRedeemStake: false,
		},
		{ // invalid redeem receiver
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomFullDenomPath,
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata(addr1.String(), "xxx", ""),
				Memo:     "",
			},
			recvDenom:      stAtomDenom,
			expSuccess:     false,
			expRedeemStake: false,
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
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
			suite.App.StakeibcKeeper.SetEpochTracker(ctx, stakeibctypes.EpochTracker{
				EpochIdentifier:    "day",
				EpochNumber:        1,
				NextEpochStartTime: uint64(now.Unix()),
				Duration:           86400,
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

			suite.App.RecordsKeeper.SetEpochUnbondingRecord(ctx, recordstypes.EpochUnbondingRecord{
				EpochNumber: 1,
				HostZoneUnbondings: []*recordstypes.HostZoneUnbonding{
					{
						HostZoneId:            "hub-1",
						Status:                recordstypes.HostZoneUnbonding_CLAIMABLE,
						UserRedemptionRecords: []string{},
						NativeTokenAmount:     sdk.NewInt(1000000),
					},
				},
			})

			// set host zone for env
			suite.App.StakeibcKeeper.SetHostZone(ctx, stakeibctypes.HostZone{
				ChainId:              "hub-1",
				ConnectionId:         "connection-0",
				Bech32Prefix:         "cosmos",
				TransferChannelId:    "channel-0",
				Validators:           []*stakeibctypes.Validator{},
				WithdrawalIcaAddress: "",
				FeeIcaAddress:        "",
				DelegationIcaAddress: "",
				RedemptionIcaAddress: "",
				IbcDenom:             atomIbcDenom,
				HostDenom:            atomHostDenom,
				RedemptionRate:       sdk.NewDec(1),
				DepositAddress:       addr1.String(),
				TotalDelegations:     sdk.NewInt(1000000),
			})

			// mint coins to be spent on liquid staking
			coins := sdk.Coins{sdk.NewInt64Coin(atomIbcDenom, 1000000)}
			err := suite.App.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			suite.Require().NoError(err)
			err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, coins)
			suite.Require().NoError(err)

			// issue liquid-stake tokens
			msgServer := stakeibckeeper.NewMsgServerImpl(suite.App.StakeibcKeeper)
			msg := stakeibctypes.NewMsgLiquidStake(addr1.String(), sdk.NewInt(1000000), atomHostDenom)
			_, err = msgServer.LiquidStake(sdk.WrapSDKContext(suite.Ctx), msg)
			suite.Require().NoError(err)

			// send tokens to ibc transfer channel escrow address
			escrowAddr := transfertypes.GetEscrowAddress(packet.DestinationPort, packet.DestinationChannel)
			err = suite.App.BankKeeper.SendCoins(suite.Ctx, addr1, escrowAddr, sdk.Coins{sdk.NewInt64Coin(stAtomDenom, 1000000)})
			suite.Require().NoError(err)
			suite.App.TransferKeeper.SetTotalEscrowForDenom(suite.Ctx, sdk.NewInt64Coin(stAtomDenom, 1000000))

			transferIBCModule := transfer.NewIBCModule(suite.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(suite.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := router.NewIBCModule(suite.App.AutopilotKeeper, recordsStack)
			ack := routerIBCModule.OnRecvPacket(
				ctx,
				packet,
				addr1,
			)
			if tc.expSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))

				// check if redeem record is created
				hostZoneUnbonding, found := suite.App.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, 1, "hub-1")
				suite.Require().True(found)
				suite.Require().True(len(hostZoneUnbonding.UserRedemptionRecords) > 0)
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}
		})
	}
}
