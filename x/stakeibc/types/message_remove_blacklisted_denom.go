package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v23/utils"
)

const TypeMsgRemoveBlacklistedDenom = "remove_blacklisted_denom"

var _ sdk.Msg = &MsgRemoveBlacklistedDenom{}

func (msg *MsgRemoveBlacklistedDenom) Route() string {
	return RouterKey
}

func (msg *MsgRemoveBlacklistedDenom) Type() string {
	return TypeMsgRemoveBlacklistedDenom
}

func (msg *MsgRemoveBlacklistedDenom) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveBlacklistedDenom) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveBlacklistedDenom) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	return nil
}
