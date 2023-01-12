package keeper

import (
	"fmt"

	"github.com/spf13/cast"
	"github.com/tendermint/tendermint/libs/log"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Stride-Labs/stride/v4/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper of the mint store.
type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         sdk.StoreKey
	paramSpace       paramtypes.Subspace
	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	epochKeeper      types.EpochKeeper
	hooks            types.MintHooks
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper, dk types.DistrKeeper, epochKeeper types.EpochKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace,
		accountKeeper:    ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		epochKeeper:      epochKeeper,
		feeCollectorName: feeCollectorName,
	}
}

// _____________________________________________________________________

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Set the mint hooks.
func (k *Keeper) SetHooks(h types.MintHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set mint hooks twice")
	}

	k.hooks = h

	return k
}

// GetLastReductionEpochNum returns last Reduction epoch number.
func (k Keeper) GetLastReductionEpochNum(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastReductionEpochKey)
	if b == nil {
		return 0
	}

	return cast.ToInt64(sdk.BigEndianToUint64(b))
}

// SetLastReductionEpochNum set last Reduction epoch number.
func (k Keeper) SetLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastReductionEpochKey, sdk.Uint64ToBigEndian(cast.ToUint64(epochNum)))
}

// get the minter.
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter.
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}

// _____________________________________________________________________

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// _____________________________________________________________________

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// GetProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the `AllocationRatio`.
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) sdk.Coin {
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt())
}

const (
	// strategic reserve address F0
	StrategicReserveAddress = "stride1alnn79kh0xka0r5h4h82uuaqfhpdmph6rvpf5f"
)

// DistributeMintedCoins implements distribution of minted coins from mint to external modules.
func (k Keeper) DistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	k.Logger(ctx).Info(fmt.Sprintf("Distributing minted x/mint rewards: %d coins...", mintedCoin.Amount.Int64()))

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module
	stakingIncentivesProportions := k.GetProportions(ctx, mintedCoin, proportions.Staking)
	stakingIncentivesCoins := sdk.NewCoins(stakingIncentivesProportions)
	k.Logger(ctx).Info(fmt.Sprintf("\t\t\t...staking rewards: %d to %s", stakingIncentivesProportions.Amount.Int64(), k.feeCollectorName))

	// allocate pool allocation ratio to strategic reserve
	strategicReserveAddress, err := sdk.AccAddressFromBech32(StrategicReserveAddress)
	if err != nil {
		errMsg := fmt.Sprintf("invalid strategic reserve address: %s", StrategicReserveAddress)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, errMsg)
	}
	strategicReserveProportion := k.GetProportions(ctx, mintedCoin, proportions.StrategicReserve)
	strategicReserveCoins := sdk.NewCoins(strategicReserveProportion)
	k.Logger(ctx).Info(fmt.Sprintf("\t\t\t...strategic reserve: %d to %s", strategicReserveProportion.Amount.Int64(), strategicReserveAddress))

	// allocate pool allocation ratio to community growth pool
	communityPoolGrowthAddress := k.GetSubmoduleAddress(types.CommunityGrowthSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	communityPoolGrowthProportion := k.GetProportions(ctx, mintedCoin, proportions.CommunityPoolGrowth)
	communityPoolGrowthCoins := sdk.NewCoins(communityPoolGrowthProportion)
	k.Logger(ctx).Info(fmt.Sprintf("\t\t\t...community growth: %d to %s", communityPoolGrowthProportion.Amount.Int64(), communityPoolGrowthAddress))

	// allocate pool allocation ratio to security budget pool
	communityPoolSecurityBudgetAddress := k.GetSubmoduleAddress(types.CommunitySecurityBudgetSubmoduleName, types.SubmoduleCommunityNamespaceKey)
	communityPoolSecurityBudgetProportion := k.GetProportions(ctx, mintedCoin, proportions.CommunityPoolSecurityBudget)
	communityPoolSecurityBudgetCoins := sdk.NewCoins(communityPoolSecurityBudgetProportion)
	k.Logger(ctx).Info(fmt.Sprintf("\t\t\t...community security budget: %d to %s", communityPoolSecurityBudgetProportion.Amount.Int64(), communityPoolSecurityBudgetAddress))

	// remaining tokens to the community growth pool (this should NEVER happen, barring rounding imprecision)
	remainingCoins := sdk.NewCoins(mintedCoin).
		Sub(stakingIncentivesCoins).
		Sub(strategicReserveCoins).
		Sub(communityPoolGrowthCoins).
		Sub(communityPoolSecurityBudgetCoins)

	// check: remaining coins should be less than 5% of minted coins
	remainingBal := remainingCoins.AmountOf(sdk.DefaultBondDenom)
	thresh := sdk.NewDec(5).Quo(sdk.NewDec(100))
	if remainingBal.ToDec().Quo(mintedCoin.Amount.ToDec()).GT(thresh) {
		errMsg := fmt.Sprintf("Failed to divvy up mint module rewards fully -- remaining coins should be LT 5pct of total, instead are %#v/%#v", remainingCoins, remainingBal)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, errMsg)
	}

	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, stakingIncentivesCoins)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, strategicReserveAddress, strategicReserveCoins)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, communityPoolGrowthAddress, communityPoolGrowthCoins)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, communityPoolSecurityBudgetAddress, communityPoolSecurityBudgetCoins)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, communityPoolGrowthAddress, remainingCoins)
	if err != nil {
		return err
	}

	// call a hook after the minting and distribution of new coins
	// see osmosis' pool incentives hooks.go for an example
	// k.hooks.AfterDistributeMintedCoin(ctx, mintedCoin)

	return nil
}

// set up a new module account address
func (k Keeper) SetupNewModuleAccount(ctx sdk.Context, submoduleName string, submoduleNamespace string) {
	// create and save the module account to the account keeper
	acctAddress := k.GetSubmoduleAddress(submoduleName, submoduleNamespace)
	acc := k.accountKeeper.NewAccount(
		ctx,
		authtypes.NewModuleAccount(
			authtypes.NewBaseAccountWithAddress(acctAddress),
			acctAddress.String(),
		),
	)
	k.Logger(ctx).Info(fmt.Sprintf("Created new %s.%s module account %s", types.ModuleName, submoduleName, acc.GetAddress().String()))
	k.accountKeeper.SetAccount(ctx, acc)
}

// helper: get the address of a submodule
func (k Keeper) GetSubmoduleAddress(submoduleName string, submoduleNamespace string) sdk.AccAddress {
	key := append([]byte(submoduleNamespace), []byte(submoduleName)...)
	return address.Module(types.ModuleName, key)
}
