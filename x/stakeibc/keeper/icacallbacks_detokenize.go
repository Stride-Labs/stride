package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// ICACallback after an LSM token is detokenized into native stake
//   If successful: Remove the token deposit from the store and unbalanced delegation
//   If failure: flag the deposit as DETOKENIZATION_FAILED
//   If timeout: do nothing
//     - A timeout will force the channel closed, and once the channel is restored,
//       the ICA will get resubmitted
func DetokenizeCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	detokenizeCallback := types.DetokenizeSharesCallback{}
	if err := proto.Unmarshal(args, &detokenizeCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal detokenize callback: %s", err.Error())
	}
	chainId := detokenizeCallback.Deposit.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Detokenize, "Starting detokenize callback"))

	// No action is necessary on a timeout
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Detokenize,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// If the ICA failed, update the deposit status
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Detokenize,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, *detokenizeCallback.Deposit, recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED)
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Detokenize,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// If the ICA succeeded, remove the token deposit
	deposit := detokenizeCallback.Deposit
	k.RecordsKeeper.RemoveLSMTokenDeposit(ctx, deposit.ChainId, deposit.Denom)

	// Update unbalanced delegation on the host zone and validator
	// TODO [LSM] - Use helper function

	return nil
}
