package keeper

import (
	"github.com/Stride-Labs/stride/v14/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v14/x/records/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// Callback after an LSM token is IBC tranferred to the host zone
//
//	If successful: mark the LSM Token status as DETOKENIZATION_QUEUE
//	If failure: mark the LSM Token status as FAILED
//	If timeout: revert the LSM Token status back to TRANSFER_QUEUE so it gets resubmitted
func (k Keeper) LSMTransferCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	transferCallback := types.TransferLSMTokenCallback{}
	if err := proto.Unmarshal(args, &transferCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal LSM transfer callback: %s", err.Error())
	}
	deposit := *transferCallback.Deposit
	chainId := deposit.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, IBCCallbacksID_LSMTransfer, "Starting LSM transfer callback"))

	// If timeout, update the status to TRANSFER_QUEUE so that it gets resubmitted
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		k.Logger(ctx).Error(utils.LogICACallbackWithHostZone(chainId, IBCCallbacksID_LSMTransfer, "Retrying transfer"))

		k.UpdateLSMTokenDepositStatus(ctx, deposit, types.LSMTokenDeposit_TRANSFER_QUEUE)
		return nil
	}

	// If the transfer failed, update the status to FAILED
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		k.UpdateLSMTokenDepositStatus(ctx, deposit, types.LSMTokenDeposit_TRANSFER_FAILED)
		return nil
	}

	// If the transfer was successful, update the status to DETOKENIZATION_QUEUE
	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	k.UpdateLSMTokenDepositStatus(ctx, deposit, types.LSMTokenDeposit_DETOKENIZATION_QUEUE)

	return nil
}
