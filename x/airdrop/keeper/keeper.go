package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,

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

func (k Keeper) SetAirdropRecords(ctx sdk.Context, airdrop_record []types.AirdropRecord) error {
	return nil
}

func (k Keeper) SetAllocationRecords(ctx sdk.Context, allocation_record []types.AllocationRecord) error {
	return nil
}
