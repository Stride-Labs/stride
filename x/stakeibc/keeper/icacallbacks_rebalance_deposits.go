package keeper

import (
	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

func RebalanceTokenizedDepositsCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	rebalanceCallback := types.RebalanceTokenizedDepositsCallback{}
	if err := proto.Unmarshal(args, &rebalanceCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, "unable to unmarshal rebalance tokenized deposits callback: %s", err.Error())
	}
	chainId := rebalanceCallback.ChainId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_LSMRebalance, "Starting rebalance tokenized deposits callback"))

	// TODO [LSM]

	return nil
}
