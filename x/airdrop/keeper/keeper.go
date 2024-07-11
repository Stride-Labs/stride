package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		logger   log.Logger

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	logger log.Logger,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		logger:   logger,

		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Claim(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}

func (k Keeper) ClaimAndStake(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}

func (k Keeper) ClaimEarly(ctx sdk.Context, claimer string) error {
	claimerAccount, err := sdk.AccAddressFromBech32(claimer)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", claimer)
	}

	// TODO implement logic

	return nil
}

func (k Keeper) GetAirdropRecords(ctx sdk.Context) []types.AirdropRecord {
	// TODO add pagination?
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allAirdrops := []types.AirdropRecord{}
	for ; iterator.Valid(); iterator.Next() {

		airdrop := types.AirdropRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &airdrop)
		allAirdrops = append(allAirdrops, airdrop)
	}

	return allAirdrops
}

func (k Keeper) SetAirdropRecords(ctx sdk.Context, airdropRecords []types.AirdropRecord) {
	for _, airdrop := range airdropRecords {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropRecordsKeyPrefix)

		key := types.AirdropRecordKeyPrefix(airdrop.Id)
		value := k.cdc.MustMarshal(&airdrop)

		store.Set(key, value)
	}
}

func (k Keeper) GetAllocationRecords(ctx sdk.Context) []types.AirdropRecord {
	// TODO add pagination?
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AllocationRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allAllocations := []types.AirdropRecord{}
	for ; iterator.Valid(); iterator.Next() {

		airdrop := types.AirdropRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &airdrop)
		allAllocations = append(allAllocations, airdrop)
	}

	return allAllocations
}

func (k Keeper) SetAllocationRecords(ctx sdk.Context, allocationRecord []types.AllocationRecord) {
	for _, allocation := range allocationRecord {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AllocationRecordsKeyPrefix)

		key := types.AllocationRecordKeyPrefix(allocation.AirdropId, allocation.UserAddress)
		value := k.cdc.MustMarshal(&allocation)

		store.Set(key, value)
	}
}
