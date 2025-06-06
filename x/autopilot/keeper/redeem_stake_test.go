package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	recordsmodule "github.com/Stride-Labs/stride/v27/x/records"

	sdk "github.com/cosmos/cosmos-sdk/types"

	router "github.com/Stride-Labs/stride/v27/x/autopilot"
	"github.com/Stride-Labs/stride/v27/x/autopilot/types"
	epochtypes "github.com/Stride-Labs/stride/v27/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func getRedeemStakeStakeibcPacketMetadata(strideAddress, redemptionReceiver string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "action": "RedeemStake", "ibc_receiver": "%[2]s" } 
			}
		}`, strideAddress, redemptionReceiver)
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
		ChainId:            HostChainId,
		Bech32Prefix:       HostBechPrefix, // required to validate claim receiver
		HostDenom:          HostDenom,
		RedemptionRate:     sdk.NewDec(1), // used to determine native token amount
		DepositAddress:     depositAddress.String(),
		TotalDelegations:   redeemAmount, // there must be enough stake to cover the redemption
		RedemptionsEnabled: true,
	})

	// fund the user with sttokens so they can redeem
	// (the function being tested is invoked downstream of the IBC transfer)
	s.FundAccount(redeemerOnStride, sdk.NewCoin("st"+HostDenom, redeemAmount))
}

// Helper function to confirm that an autopilot redemption succeeded by confirming a redemption
// record was created and tokens were escrowed in the deposit account
func (s *KeeperTestSuite) CheckRedeemStakeSucceeded(redeemAmount sdkmath.Int, redeemDenom string, depositAddress sdk.AccAddress) {
	// check if redeem record is created
	hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, 1, HostChainId)
	s.Require().True(found)
	s.Require().True(len(hostZoneUnbonding.UserRedemptionRecords) > 0,
		"user redemption record should have been created")

	// check that tokens were escrowed
	escrowBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAddress, redeemDenom)
	s.Require().Equal(redeemAmount.Int64(), escrowBalance.Amount.Int64(), "tokens should have been escrowed")
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

	// For strd, the hub's channel ID would have been appended to the denom trace
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
			name:        "failed because param not enabled",
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
			expectedError: "invalid amount",
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
			expectedError: "not a liquid staking token",
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

				s.CheckRedeemStakeSucceeded(redeemAmount, tc.redeemDenom, depositAddress)
			} else {
				s.Require().ErrorContains(err, tc.expectedError, tc.name)
			}
		})
	}
}

// TODO: Move to ibc_test.go when OnRecvPacket is moved
func (s *KeeperTestSuite) TestOnRecvPacket_RedeemStake() {
	redeemerOnStride := s.TestAccs[0]
	depositAddress := s.TestAccs[1]
	differentAddress := s.TestAccs[2].String()
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

	// For stTokens, the hub's channel ID would have been appended to the denom trace
	stAtom := "stuatom"
	stAtomTrace := transfertypes.GetPrefixedDenom(transfertypes.PortID, hubToStrideChannel, stAtom)

	testCases := []struct {
		name       string
		enabled    bool
		packetData transfertypes.FungibleTokenPacketData
		expSuccess bool
	}{
		{
			name:    "successful redemption",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   redeemerOnHost,
				Receiver: redeemerOnStride.String(),
				Memo:     getRedeemStakeStakeibcPacketMetadata(redeemerOnStride.String(), redeemerOnHost),
			},
			expSuccess: true,
		},
		{
			name:    "failed because param not enabled",
			enabled: false,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   redeemerOnHost,
				Receiver: redeemerOnStride.String(),
				Memo:     getRedeemStakeStakeibcPacketMetadata(redeemerOnStride.String(), redeemerOnHost),
			},
			expSuccess: false,
		},
		{
			name:    "failed because invalid stride address in memo",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: redeemerOnStride.String(),
				Memo:     getRedeemStakeStakeibcPacketMetadata("XXX", redeemerOnHost),
			},
			expSuccess: false,
		},
		{
			name:    "failed because invalid stride address in reciever",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: "XXX",
				Memo:     getRedeemStakeStakeibcPacketMetadata(redeemerOnStride.String(), redeemerOnHost),
			},
			expSuccess: false,
		},
		{
			name:    "failed because transfer receiver address does not match memo receiver",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: differentAddress,
				Memo:     getRedeemStakeStakeibcPacketMetadata(redeemerOnStride.String(), redeemerOnHost),
			},
			expSuccess: false,
		},
		{
			name:    "failed because not stride address",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata("osmo15440wjgs208zm6dz8wvk23z5lmcx9hyxk0ew3c", redeemerOnHost),
				Memo:     "",
			},
			expSuccess: false,
		},
		{
			name:    "failed because invalid redeem address",
			enabled: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    stAtomTrace,
				Amount:   redeemAmount.String(),
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: getRedeemStakeStakeibcPacketMetadata(redeemerOnStride.String(), "XXX"),
				Memo:     "",
			},
			expSuccess: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // required since testing full ibc module

			s.SetupAutopilotRedeemStake(tc.enabled, redeemAmount, depositAddress, redeemerOnStride)

			// send tokens to ibc transfer channel escrow address
			stAtomCoin := sdk.NewInt64Coin(stAtom, 1000000)
			escrowAddr := transfertypes.GetEscrowAddress(packet.DestinationPort, packet.DestinationChannel)
			s.FundAccount(escrowAddr, stAtomCoin)
			s.App.TransferKeeper.SetTotalEscrowForDenom(s.Ctx, stAtomCoin)

			transferIBCModule := transfer.NewIBCModule(s.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(s.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := router.NewIBCModule(s.App.AutopilotKeeper, recordsStack)

			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&tc.packetData)
			ack := routerIBCModule.OnRecvPacket(
				s.Ctx,
				packet,
				s.TestAccs[2],
			)

			if tc.expSuccess {
				s.Require().True(ack.Success(), string(ack.Acknowledgement()))

				s.CheckRedeemStakeSucceeded(redeemAmount, stAtom, depositAddress)
			} else {
				s.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}
		})
	}
}
