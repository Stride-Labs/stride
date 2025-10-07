package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	claimvestingtypes "github.com/Stride-Labs/stride/v28/x/claim/vesting/types"
)

// Downgrades a vesting account to a base account
// This downgrades the following:
//   - PeriodicVestingAccount
//   - ContinuousVestingAccount
//   - StridePeriodicVestingAccount
//
// Although, on mainnet it should not be called for StridePeriodicVestingAccount
// since all those accounts are already vested
func (k Keeper) DowngradeVestingAccount(ctx sdk.Context, address sdk.AccAddress) error {
	account := k.accountKeeper.GetAccount(ctx, address)

	if periodicVestingAccount, ok := account.(*vestingtypes.PeriodicVestingAccount); ok {
		k.accountKeeper.SetAccount(ctx, periodicVestingAccount.BaseVestingAccount.BaseAccount)
		return nil
	}

	if continuousVestingAccount, ok := account.(*vestingtypes.ContinuousVestingAccount); ok {
		k.accountKeeper.SetAccount(ctx, continuousVestingAccount.BaseVestingAccount.BaseAccount)
		return nil
	}

	if strideVestingAccount, ok := account.(*claimvestingtypes.StridePeriodicVestingAccount); ok {
		k.accountKeeper.SetAccount(ctx, strideVestingAccount.BaseVestingAccount.BaseAccount)
		return nil
	}

	return errors.New("unable to downgrade vesiting account, account not recognized")
}
