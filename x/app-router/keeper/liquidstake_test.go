package keeper_test

import (
	"fmt"
	"time"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v3/x/app-router/types"
	epochtypes "github.com/Stride-Labs/stride/v3/x/epochs/types"
	minttypes "github.com/Stride-Labs/stride/v3/x/mint/types"
	recordstypes "github.com/Stride-Labs/stride/v3/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

func (suite *KeeperTestSuite) TestTryLiquidStaking() {
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

	testCases := []struct {
		forwardingActive bool
		recvDenom        string
		packetData       transfertypes.FungibleTokenPacketData
		expNilAck        bool
	}{
		{ // params not enabled
			forwardingActive: false,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "",
				Receiver: "",
				Memo:     "",
			},
			recvDenom: atomIbcDenom,
			expNilAck: false,
		},
		{ // all okay
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    "uatom",
				Amount:   "1000000",
				Sender:   "",
				Receiver: "",
				Memo:     "",
			},
			recvDenom: atomIbcDenom,
			expNilAck: true,
		},
		{ // strd denom
			forwardingActive: true,
			packetData: transfertypes.FungibleTokenPacketData{
				Denom:    strdIbcDenom,
				Amount:   "1000000",
				Sender:   "",
				Receiver: "",
				Memo:     "",
			},
			recvDenom: "ustrd",
			expNilAck: false,
		},
	}

	for i, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %d", i), func() {
			suite.SetupTest() // reset
			ctx := suite.Ctx()

			suite.App.RouterKeeper.SetParams(ctx, types.Params{Active: tc.forwardingActive})

			addr1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

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

			ack := suite.App.RouterKeeper.TryLiquidStaking(
				ctx,
				packet,
				tc.packetData,
				&types.ParsedReceiver{
					ShouldLiquidStake: true,
					StrideAccAddress:  addr1,
				},
				nil,
			)
			if tc.expNilAck {
				suite.Require().Nil(ack)
			} else {
				suite.Require().NotNil(ack)
			}
		})
	}
}
