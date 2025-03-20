package utils

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type BankKeeper interface {
	BlockedAddr(addr sdk.AccAddress) bool
	SendCoins(context context.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// SafeSendCoins transfers coins from one account to another with additional safety checks.
// It provides an extra layer of protection by optionally verifying if the recipient
// address is blocked from receiving funds before proceeding with the transfer.
//
// Parameters:
//   - checkBlockedAddr: if true, checks whether the recipient address is blocked (set to false if recipientAddr is a module)
//   - bankKeeper: the BankKeeper interface that handles coin transfers
//   - ctx: the SDK Context for the transaction
//   - senderAddr: the account address sending the coins
//   - recipientAddr: the account address receiving the coins
//   - amt: the coins to be transferred
//
// Returns:
//   - error: nil if the transfer is successful, otherwise an error
//   - returns sdkerrors.ErrUnauthorized if checkBlockedAddr is true and recipient is blocked
//   - returns any error from the underlying SendCoins operation
func SafeSendCoins(checkBlockedAddr bool, bankKeeper BankKeeper, ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if checkBlockedAddr && bankKeeper.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return bankKeeper.SendCoins(ctx, senderAddr, recipientAddr, amt)
}
