package keeper

import (
	"github.com/Stride-Labs/stride/v8/utils"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func DetokenizeCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	detokenizeCallback := types.DetokenizeSharesCallback{}
	if err := proto.Unmarshal(args, &detokenizeCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal detokenize callback: %s", err.Error())
	}
	chainId := detokenizeCallback.Deposit.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Detokenize, "Starting detokenize callback"))

	// TODO [LSM]

	return nil
}
