package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v5/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck
)

// Marshalls oracle callback arguments
func (k Keeper) MarshalInstantiateOracleCallbackArgs(ctx sdk.Context, instantiateCallback types.InstantiateOracleCallback) ([]byte, error) {
	out, err := proto.Marshal(&instantiateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("MarshalInstantiateOracleCallbackArgs %v", err.Error()))
		return nil, err
	}
	return out, nil
}

// Unmarshalls oracle callback arguments into a OracleCallback struct
func (k Keeper) UnmarshalInstantiateOracleCallbackArgs(ctx sdk.Context, oracleCallbackBz []byte) (*types.InstantiateOracleCallback, error) {
	instantiateCallback := types.InstantiateOracleCallback{}
	if err := proto.Unmarshal(oracleCallbackBz, &instantiateCallback); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalInstantiateOracleCallbackArgs %v", err.Error()))
		return nil, err
	}
	return &instantiateCallback, nil
}

// Callback after an update oracle ICA
//     If successful: Stores the cosmwasm contract address on the oracle object
//     If timeout/failure: Does nothing
func InstantiateOracleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// TODO
	return nil
}
