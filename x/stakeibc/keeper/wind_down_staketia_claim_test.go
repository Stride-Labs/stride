// x/stakeibc/keeper/wind_down_staketia_claim_test.go
package keeper_test

import (
	"time"

	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v11/testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

type staketiaClaimTestCase struct {
	claimAddress   sdk.AccAddress
	escrowAddress  sdk.AccAddress
	tiaDenom       string
	initialBalance sdkmath.Int
	strdBalance    sdkmath.Int
}

func (s *KeeperTestSuite) SetupTransferStaketiaClaimBalance() staketiaClaimTestCase {
	s.CreateTransferChannel(staketiatypes.CelestiaChainId)

	// The real voucher trace: registering it makes the hash equal the staketia constant
	tiaTrace := transfertypes.NewDenom("utia", transfertypes.NewHop(transfertypes.PortID, staketiatypes.StrideToCelestiaTransferChannelId))
	s.App.TransferKeeper.SetDenom(s.Ctx, tiaTrace)
	s.Require().Equal(staketiatypes.CelestiaNativeTokenIBCDenom, tiaTrace.IBCDenom(), "test trace must hash to the constant")

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, types.HostZone{
		ChainId:              staketiatypes.CelestiaChainId,
		TransferChannelId:    ibctesting.FirstChannelID,
		DelegationIcaAddress: "celestia_DELEGATION",
	})

	claimAddress := sdk.MustAccAddressFromBech32(staketiatypes.ClaimAddress)
	initialBalance := sdkmath.NewInt(1_000_000)
	strdBalance := sdkmath.NewInt(500)
	s.FundAccount(claimAddress, sdk.NewCoin(tiaTrace.IBCDenom(), initialBalance))
	s.FundAccount(claimAddress, sdk.NewCoin("ustrd", strdBalance))

	return staketiaClaimTestCase{
		claimAddress:   claimAddress,
		escrowAddress:  transfertypes.GetEscrowAddress(transfertypes.PortID, ibctesting.FirstChannelID),
		tiaDenom:       tiaTrace.IBCDenom(),
		initialBalance: initialBalance,
		strdBalance:    strdBalance,
	}
}

func (s *KeeperTestSuite) checkClaimTransferred(tc staketiaClaimTestCase, transferred sdkmath.Int) {
	// compared as strings: a zero Int from Sub and one from the store differ in big.Int representation
	s.Require().Equal(tc.initialBalance.Sub(transferred).String(), s.App.BankKeeper.GetBalance(s.Ctx, tc.claimAddress, tc.tiaDenom).Amount.String(), "claim address TIA")
	s.Require().Equal(transferred.String(), s.App.BankKeeper.GetBalance(s.Ctx, tc.escrowAddress, tc.tiaDenom).Amount.String(), "escrowed TIA")
	s.Require().Equal(tc.strdBalance, s.App.BankKeeper.GetBalance(s.Ctx, tc.claimAddress, "ustrd").Amount, "only TIA moves")
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_ZeroAmountSendsEverything() {
	tc := s.SetupTransferStaketiaClaimBalance()
	startSequence := s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)

	transferred, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, tc.initialBalance), transferred)

	s.Require().Equal(startSequence+1, s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID), "one transfer submitted")
	s.checkClaimTransferred(tc, tc.initialBalance)
	s.CheckEventValueEmitted(types.EventTypeTransferStaketiaClaimBalance, types.AttributeKeyAmount, transferred.String())
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_PositiveAmountSendsExactly() {
	tc := s.SetupTransferStaketiaClaimBalance()
	amount := sdkmath.NewInt(400)

	transferred, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", amount))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, amount), transferred)
	s.checkClaimTransferred(tc, amount)
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_Rejections() {
	tc := s.SetupTransferStaketiaClaimBalance()

	// above the balance
	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", tc.initialBalance.AddRaw(1)))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)
	s.checkClaimTransferred(tc, sdkmath.ZeroInt())

	// zero balance
	s.Require().NoError(s.App.BankKeeper.SendCoins(s.Ctx, tc.claimAddress, s.TestAccs[0], sdk.NewCoins(sdk.NewCoin(tc.tiaDenom, tc.initialBalance))))
	_, err = s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, sdkerrors.ErrInsufficientFunds)

	// no celestia host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	_, err = s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, types.ErrHostZoneNotFound)
}

func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_MissingDelegationIca() {
	s.SetupTransferStaketiaClaimBalance()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, staketiatypes.CelestiaChainId)
	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.ZeroInt()))
	s.Require().ErrorIs(err, types.ErrICAAccountNotFound)
}

func (s *KeeperTestSuite) TestBuildStaketiaClaimTransferMsg() {
	hostZone := types.HostZone{TransferChannelId: "channel-162", DelegationIcaAddress: "celestia_DELEGATION"}
	token := sdk.NewCoin(staketiatypes.CelestiaNativeTokenIBCDenom, sdkmath.NewInt(7))
	timeout := uint64(s.Ctx.BlockTime().Add(24 * time.Hour).UnixNano())

	msg := keeper.BuildStaketiaClaimTransferMsg(hostZone, token, timeout)
	s.Require().Equal(transfertypes.PortID, msg.SourcePort)
	s.Require().Equal("channel-162", msg.SourceChannel)
	s.Require().Equal(token, msg.Token)
	s.Require().Equal(staketiatypes.ClaimAddress, msg.Sender)
	s.Require().Equal("celestia_DELEGATION", msg.Receiver)
	s.Require().Equal(timeout, msg.TimeoutTimestamp)
	s.Require().Zero(msg.TimeoutHeight.RevisionHeight)
	s.Require().Empty(msg.Memo)
}

// A timed-out packet refunds the claim address through the normal ICS-20 path
func (s *KeeperTestSuite) TestTransferStaketiaClaimBalance_TimeoutRefundsClaimAddress() {
	tc := s.SetupTransferStaketiaClaimBalance()
	amount := sdkmath.NewInt(400)
	_, err := s.App.StakeibcKeeper.TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", amount))
	s.Require().NoError(err)
	s.checkClaimTransferred(tc, amount)

	// Rebuild the packet's token from the registered trace (the same lookup lsm.go uses)
	hash, err := transfertypes.ParseHexHash(tc.tiaDenom[len("ibc/"):])
	s.Require().NoError(err)
	denom, found := s.App.TransferKeeper.GetDenom(s.Ctx, hash)
	s.Require().True(found)
	packetData := transfertypes.NewInternalTransferRepresentation(
		transfertypes.Token{Denom: denom, Amount: amount.String()},
		staketiatypes.ClaimAddress, "celestia_DELEGATION", "",
	)
	err = s.App.TransferKeeper.OnTimeoutPacket(s.Ctx, transfertypes.PortID, ibctesting.FirstChannelID, packetData)
	s.Require().NoError(err)

	s.checkClaimTransferred(tc, sdkmath.ZeroInt())
}

func (s *KeeperTestSuite) TestMsgServer_TransferStaketiaClaimBalance() {
	tc := s.SetupTransferStaketiaClaimBalance()
	resp, err := s.GetMsgServer().TransferStaketiaClaimBalance(s.Ctx, types.NewMsgTransferStaketiaClaimBalance("admin", sdkmath.NewInt(10)))
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewCoin(tc.tiaDenom, sdkmath.NewInt(10)), resp.Transferred)
}
