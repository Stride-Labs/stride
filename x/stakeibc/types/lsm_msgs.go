// #nosec G101
package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// staking message types
const (
	TypeMsgRedeemTokensForShares = "redeem_tokens_for_shares" // #nosec G101
)

var _ sdk.Msg = &MsgRedeemTokensForShares{}

// NewMsgRedeemTokensForShares creates a new MsgRedeemTokensForShares instance.
//
//nolint:interfacer
func NewMsgRedeemTokensForShares(delAddr sdk.AccAddress, amount sdk.Coin) *MsgRedeemTokensForShares {
	return &MsgRedeemTokensForShares{
		DelegatorAddress: delAddr.String(),
		Amount:           amount,
	}
}

// Route implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) Type() string { return TypeMsgRedeemTokensForShares }

// GetSigners implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgRedeemTokensForShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"invalid shares amount",
		)
	}

	return nil
}
