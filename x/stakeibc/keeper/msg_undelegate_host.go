package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const isUndelegateHostPreventedKey = "is-undelegate-host-prevented"

func (k msgServer) UndelegateHost(goCtx context.Context, msg *types.MsgUndelegateHost) (*types.MsgUndelegateHostResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// undelegateHost is callable only if it has not yet been called and succeeded
	if k.IsUndelegateHostPrevented(ctx) {
		return nil, errorsmod.Wrapf(types.ErrUndelegateHostNotCallable, "")
	}

	// Get host zone unbonding message by summing up the unbonding records
	if err := k.UndelegateHostEvmos(ctx, msg.Amount); err != nil {
		return nil, fmt.Errorf("Error initiating host zone unbondings for UndelegateHostEvmos %s", err.Error())
	}

	// log: issuing an undelegation to Evmos
	k.Logger(ctx).Info(fmt.Sprintf("Issuing an undelegation to Evmos"))

	return &types.MsgUndelegateHostResponse{}, nil
}

func (k Keeper) SetUndelegateHostPrevented(ctx sdk.Context) error {

	store := ctx.KVStore(k.storeKey)

	// set the key to 1 if it's not set
	if !k.IsUndelegateHostPrevented(ctx) {
		store.Set([]byte(isUndelegateHostPreventedKey), []byte{1})
	}
	return nil
}

func (k Keeper) IsUndelegateHostPrevented(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(isUndelegateHostPreventedKey)) {
		return false
	}

	value := store.Get([]byte(isUndelegateHostPreventedKey))
	return len(value) == 1 && value[0] == 1
}
