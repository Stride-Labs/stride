package keeper

import (
	"errors"
	"fmt"

	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
		cdc := k.cdc
		DelegateOnHost := k.DelegateOnHost

		var queryBalanceCB icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
			icqk.Logger(ctx).Info(fmt.Sprintf("\tdelegation staking callback on %s", zoneInfo.HostDenom))
			queryRes := bankTypes.QueryAllBalancesResponse{}
			err := cdc.Unmarshal(args, &queryRes)
			if err != nil {
				icqk.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
				return err
			}
			// Get denom dynamically
			balance := queryRes.Balances.AmountOf(zoneInfo.HostDenom)
			icqk.Logger(ctx).Info(fmt.Sprintf("\tBalance on %s is %s", zoneInfo.HostDenom, balance.String()))

			processAmount := balance.String() + zoneInfo.HostDenom
			amt, err := sdk.ParseCoinNormalized(processAmount)
			if err != nil {
				icqk.Logger(ctx).Error(fmt.Sprintf("Could not process coin %s: %s", zoneInfo.HostDenom, err))
				return err
			}
			err = DelegateOnHost(ctx, zoneInfo, amt)
			if err != nil {
				icqk.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s", processAmount, zoneInfo.ChainId))
				return sdkerrors.Wrapf(types.ErrInvalidHostZone, "Couldn't stake %s on %s", processAmount, zoneInfo.ChainId)
			} else {
				icqk.Logger(ctx).Info(fmt.Sprintf("Successfully staked %s on %s", processAmount, zoneInfo.ChainId))
			}

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute("hostZone", zoneInfo.ChainId),
					sdk.NewAttribute("newAmountStaked", balance.String()),
				),
			})

			// --- Update Undelegated Balance ---
			hz := zoneInfo

			da := hz.DelegationAccount
			da.Balance = balance.Int64()
			hz.DelegationAccount = da
			k.SetHostZone(ctx, hz)

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					// sdk.NewAttribute("totalUndelegatedBalance", balance.String()),
					sdk.NewAttribute("totalUndelegatedBalance", balance.String()),
				),
			})
			// ---------------------------------

			return nil
		}
		k.Logger(ctx).Info(fmt.Sprintf("\tQuerying balance for %s", zoneInfo.ChainId))
		k.InterchainQueryKeeper.QueryBalances(ctx, zoneInfo, queryBalanceCB, delegationIca.Address)
		return nil
	}

	// Iterate the zones and apply icaStake
	k.IterateHostZones(ctx, icaStake)
}
