package keeper

import (
	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v9/x/records/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Callback after an LSM token is IBC tranferred to the host zone
//   If successful: mark the LSM Token status as DETOKENIZATION_QUEUE
//   If failure: mark the LSM Token status as FAILED
//   If timeout: re-submit the IBC transfer
func LSMTransferCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	transferCallback := types.TransferLSMTokenCallback{}
	if err := proto.Unmarshal(args, &transferCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal LSM transfer callback: %s", err.Error())
	}
	chainId := transferCallback.Deposit.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, IBCCallbacksID_LSMTransfer, "Starting LSM transfer callback"))

	// If timeout, retry the transfer
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		// TODO [LSM]
		return nil
	}

	// If the transfer failed, update the status to FAILED
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		k.UpdateLSMTokenDepositStatus(ctx, *transferCallback.Deposit, types.LSMTokenDeposit_TRANSFER_FAILED)
		return nil
	}

	// If the transfer was successful, update the status to DETOKENIZATION_QUEUE
	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, IBCCallbacksID_LSMTransfer,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	k.UpdateLSMTokenDepositStatus(ctx, *transferCallback.Deposit, types.LSMTokenDeposit_DETOKENIZATION_QUEUE)

	return nil
}
