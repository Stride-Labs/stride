package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRedeemStake = "redeem_stake"

var _ sdk.Msg = &MsgRedeemStake{}

func NewMsgRedeemStake(creator string, amount int64, denom string) *MsgRedeemStake {
	return &MsgRedeemStake{
		Creator: creator,
		Amount:  amount,
		Denom:   denom,
	}
}

// isStAsset checks if the denom of the asset matches our stAsset prefix
// Note, this is not definitive (other f'st{something}' tokens could be minted)
func IsStAsset(denom string) bool {
	return strings.HasPrefix(denom, "st")
}

func (msg *MsgRedeemStake) Route() string {
	return RouterKey
}

func (msg *MsgRedeemStake) Type() string {
	return TypeMsgRedeemStake
}

func (msg *MsgRedeemStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRedeemStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRedeemStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	isStAsset := IsStAsset(msg.Denom)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	} else if !isStAsset {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid token denom (requires stAsset) (%s)", msg.Denom)
	}
	return nil
}
