package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	recordsmodule "github.com/Stride-Labs/stride/v9/x/records"

	sdk "github.com/cosmos/cosmos-sdk/types"

	router "github.com/Stride-Labs/stride/v9/x/autopilot"
	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v9/x/mint/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func getRedeemStakeStakeibcPacketMetadata(address, ibcReceiver, transferChannel string) string {
	return fmt.Sprintf(`
		{
			"autopilot": {
				"receiver": "%[1]s",
				"stakeibc": { "stride_address": "%[1]s", "action": "RedeemStake", "ibc_receiver": "%[2]s", "transfer_channel": "%[3]s" } 
			}
		}`, address, ibcReceiver, transferChannel)
}

func (suite *KeeperTestSuite) TestOnRecvPacket_RedeemStake() {
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
				ChainId:               "hub-1",
				ConnectionId:          "connection-0",
				Bech32Prefix:          "cosmos",
				TransferChannelId:     "channel-0",
				Validators:            []*stakeibctypes.Validator{},
				BlacklistedValidators: []*stakeibctypes.Validator{},
				WithdrawalAccount:     nil,
				FeeAccount:            nil,
				DelegationAccount:     nil,
				RedemptionAccount:     nil,
				IbcDenom:              atomIbcDenom,
				HostDenom:             atomHostDenom,
				RedemptionRate:        sdk.NewDec(1),
				Address:               addr1.String(),
				StakedBal:             sdk.NewInt(1000000),
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
