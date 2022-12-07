package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	recordsmodule "github.com/Stride-Labs/stride/v3/x/records"

	sdk "github.com/cosmos/cosmos-sdk/types"

	router "github.com/Stride-Labs/stride/v3/x/app-router"
	"github.com/Stride-Labs/stride/v3/x/app-router/types"
	epochtypes "github.com/Stride-Labs/stride/v3/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v3/x/mint/types"
	recordstypes "github.com/Stride-Labs/stride/v3/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

func (suite *KeeperTestSuite) TestOnRecvPacket() {
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

	addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	testCases := []struct {
		forwardingActive bool
		recvDenom        string
		packetData       transfertypes.FungibleTokenPacketData
		expSuccess       bool
		expLiquidStake   bool
	}{
		{ // params not enabled
			forwardingActive: false,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: fmt.Sprintf("%s|stakeibc/LiquidStake", addr1.String()),
				Memo:     "",
			},
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
				Receiver: fmt.Sprintf("%s|stakeibc/LiquidStake", addr1.String()),
				Memo:     "",
			},
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
				Receiver: fmt.Sprintf("%s|stakeibc/LiquidStake", addr1.String()),
				Memo:     "",
			},
			recvDenom:      atomIbcDenom,
			expSuccess:     true,
			expLiquidStake: true,
		},
		{ // all okay with memo liquidstaking since ibc-go v5.1.0
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: fmt.Sprintf("%s", addr1.String()),
				Memo:     "stakeibc/LiquidStake",
			},
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
				Receiver: fmt.Sprintf("%s", addr1.String()),
				Memo:     "",
			},
			recvDenom:      atomIbcDenom,
			expSuccess:     true,
			expLiquidStake: false,
		},
		{ // invalid receiver
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: "xxx|stakeibc/LiquidStake",
				Memo:     "",
			},
			recvDenom:      atomIbcDenom,
			expSuccess:     false,
			expLiquidStake: false,
		},
		{ // invalid receiver liquid staking
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k",
				Receiver: "xxx|stakeibc/LiquidStake",
				Memo:     "",
			},
			recvDenom:      atomIbcDenom,
			expSuccess:     false,
			expLiquidStake: false,
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			packet.Data = transfertypes.ModuleCdc.MustMarshalJSON(&tc.packetData)

			suite.SetupTest() // reset
			ctx := suite.Ctx()

			suite.App.RouterKeeper.SetParams(ctx, types.Params{Active: tc.forwardingActive})

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
				Amount:             100,
				Denom:              atomIbcDenom,
				HostZoneId:         "hub-1",
				Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
				DepositEpochNumber: 1,
				Source:             recordstypes.DepositRecord_STRIDE,
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
				IBCDenom:              atomIbcDenom,
				HostDenom:             atomHostDenom,
				RedemptionRate:        sdk.NewDec(1),
				Address:               addr1.String(),
			})

			// mint coins to be spent on liquid staking
			coins := sdk.Coins{sdk.NewInt64Coin(tc.recvDenom, 1000000)}
			err := suite.App.BankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			suite.Require().NoError(err)
			err = suite.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr1, coins)
			suite.Require().NoError(err)

			transferIBCModule := transfer.NewIBCModule(suite.App.TransferKeeper)
			recordsStack := recordsmodule.NewIBCModule(suite.App.RecordsKeeper, transferIBCModule)
			routerIBCModule := router.NewIBCModule(suite.App.RouterKeeper, recordsStack)
			ack := routerIBCModule.OnRecvPacket(
				ctx,
				packet,
				addr1,
			)
			if tc.expSuccess {
				suite.Require().True(ack.Success(), string(ack.Acknowledgement()))

				// check minted balance for liquid staking
				allBalance := suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				liquidBalance := suite.App.BankKeeper.GetBalance(ctx, addr1, "stuatom")
				if tc.expLiquidStake {
					suite.Require().True(liquidBalance.Amount.IsPositive(), allBalance.String())
				} else {
					suite.Require().True(liquidBalance.Amount.IsZero(), allBalance.String())
				}
			} else {
				suite.Require().False(ack.Success(), string(ack.Acknowledgement()))
			}
		})
	}
}
