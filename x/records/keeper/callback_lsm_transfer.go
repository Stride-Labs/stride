package keeper

import (
	"github.com/Stride-Labs/stride/v9/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v9/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func LSMTransferCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	transferCallback := stakeibctypes.TransferLSMTokenCallback{}
	if err := proto.Unmarshal(args, &transferCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal LSM transfer callback: %s", err.Error())
	}
	chainId := transferCallback.Deposit.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, IBCCallbacksID_LSMTransfer, "Starting LSM transfer callback"))

	// TODO [LSM]

	return nil
}
