package keeper

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SetDelegation set delegation in the store
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))
	b := k.cdc.MustMarshal(&delegation)
	store.Set([]byte{0}, b)
}

// GetDelegation returns delegation
func (k Keeper) GetDelegation(ctx sdk.Context) (val types.Delegation, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveDelegation removes delegation from the store
func (k Keeper) RemoveDelegation(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))
	store.Delete([]byte{0})
}

// ProcessDelegationStaking goes through each HostZone and stakes the delegation
func (k Keeper) ProcessDelegationStaking(ctx sdk.Context) {
	icaStake := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		// Verify the delegation ICA is registered
		k.Logger(ctx).Info(fmt.Sprintf("\tProcessing delegation %s", zoneInfo.ChainId))
		delegationIca := zoneInfo.GetDelegationAccount()
		if delegationIca == nil || delegationIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
			return errors.New("Zone is missing a delegation address!")
		}
		balance := zoneInfo.DelegationAccount.Balance
		processAmount := strconv.FormatInt(balance, 10) + zoneInfo.HostDenom
		amt, err := sdk.ParseCoinNormalized(processAmount)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not process coin %s: %s", zoneInfo.HostDenom, err))
			return err
		}
		err = k.DelegateOnHost(ctx, zoneInfo, amt)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s", processAmount, zoneInfo.ChainId))
			return sdkerrors.Wrapf(types.ErrInvalidHostZone, "Couldn't stake %s on %s", processAmount, zoneInfo.ChainId)
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Successfully staked %s on %s", processAmount, zoneInfo.ChainId))
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("hostZone", zoneInfo.ChainId),
				sdk.NewAttribute("newAmountStaked", strconv.FormatInt(balance, 10)),
			),
		})
		return nil
	}

	// Iterate the zones and apply icaStake
	k.IterateHostZones(ctx, icaStake)
}
