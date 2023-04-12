package types

// TODO: Import type definition directly from LSM once LSM has upgraded to SDK 47

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgRedeemTokensforShares = "redeem_tokens_for_shares"
)

var (
	_ sdk.Msg = &MsgRedeemTokensforShares{}
)

// Type implements the sdk.Msg interface.
func (msg MsgRedeemTokensforShares) Type() string { return TypeMsgRedeemTokensforShares }

func (msg MsgRedeemTokensforShares) GetSigners() []sdk.AccAddress {
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{delegator}
}

func (msg MsgRedeemTokensforShares) GetSignBytes() []byte {
	bz := legacy.Cdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgRedeemTokensforShares) ValidateBasic() error {
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
