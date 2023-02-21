package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Unmarshalls oracle callback arguments into a OracleCallback struct
func (k Keeper) UnmarshalUpdateOracleCallbackArgs(ctx sdk.Context, updateCallbackBz []byte) (*types.UpdateOracleCallback, error) {
	updateCallback := types.UpdateOracleCallback{}
	if err := proto.Unmarshal(updateCallbackBz, &updateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalUpdateOracleCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &updateCallback, nil
}

// Callback after an update oracle ICA
// Removes metric from pending store (regardless of ack status)
func UpdateOracleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// TODO
	return nil
}
