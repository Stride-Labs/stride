package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v14/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	"github.com/Stride-Labs/stride/v14/x/records/types"
)

var (
	// Timeout for the IBC transfer of the LSM Token to the host zone
	LSMDepositTransferTimeout = time.Hour * 24 // 1 day
)

// Transfers native tokens, accumulated from normal liquid stakes, to the host zone
// This is invoked epochly
func (k Keeper) IBCTransferNativeTokens(ctx sdk.Context, msg *transfertypes.MsgTransfer, depositRecord types.DepositRecord) error {
	// Submit IBC transfer
	msgTransferResponse, err := k.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return err
	}

	// Build callback data
	transferCallback := types.TransferCallback{
		DepositRecordId: depositRecord.Id,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Marshalling TransferCallback args: %+v", transferCallback))
	marshalledCallbackArgs, err := k.MarshalTransferCallbackArgs(ctx, transferCallback)
	if err != nil {
		return err
	}

	// Store the callback data
	sequence := msgTransferResponse.Sequence
	callback := icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(msg.SourcePort, msg.SourceChannel, sequence),
		PortId:       msg.SourcePort,
		ChannelId:    msg.SourceChannel,
		Sequence:     sequence,
		CallbackId:   IBCCallbacksID_NativeTransfer,
		CallbackArgs: marshalledCallbackArgs,
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Storing callback data: %+v", callback))
	k.ICACallbacksKeeper.SetCallbackData(ctx, callback)

	// update the record state to TRANSFER_IN_PROGRESS
	depositRecord.Status = types.DepositRecord_TRANSFER_IN_PROGRESS
	k.SetDepositRecord(ctx, depositRecord)

	return nil
}

// Transfer's LSM Tokens to the host from LSMLiquidStakes
// This is invoked immediately after the LSMLiquidStake
func (k Keeper) IBCTransferLSMToken(
	ctx sdk.Context,
	lsmTokenDeposit types.LSMTokenDeposit,
	transferChannelID string,
	hostZoneDepositAddress string,
	hostZoneDelegationICAAddress string,
) error {
	// Build transfer message with a conservative timeout
	timeout := uint64(ctx.BlockTime().UnixNano() + (LSMDepositTransferTimeout).Nanoseconds())
	ibcToken := sdk.NewCoin(lsmTokenDeposit.IbcDenom, lsmTokenDeposit.Amount)
	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    transferChannelID,
		Token:            ibcToken,
		Sender:           hostZoneDepositAddress,
		Receiver:         hostZoneDelegationICAAddress,
		TimeoutTimestamp: timeout,
	}

	// Send LSM Token to host zone via IBC transfer
	msgTransferResponse, err := k.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), &transferMsg)
	if err != nil {
		return err
	}

	// Store transfer callback data
	callbackArgs := types.TransferLSMTokenCallback{
		Deposit: &lsmTokenDeposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "Unable to marshal transfer callback data for %+v", callbackArgs)
	}

	k.ICACallbacksKeeper.SetCallbackData(ctx, icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(transferMsg.SourcePort, transferMsg.SourceChannel, msgTransferResponse.Sequence),
		PortId:       transferMsg.SourcePort,
		ChannelId:    transferMsg.SourceChannel,
		Sequence:     msgTransferResponse.Sequence,
		CallbackId:   IBCCallbacksID_LSMTransfer,
		CallbackArgs: callbackArgsBz,
	})

	return nil
}
