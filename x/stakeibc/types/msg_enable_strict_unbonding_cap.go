package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

const TypeMsgEnableStrictUnbondingCap = "enable_strict_unbonding_cap"

var _ sdk.Msg = &MsgEnableStrictUnbondingCap{}

func NewMsgEnableStrictUnbondingCap(creator string) *MsgEnableStrictUnbondingCap {
	return &MsgEnableStrictUnbondingCap{
		Creator: creator,
	}
}

func (msg *MsgEnableStrictUnbondingCap) Route() string {
	return RouterKey
}

func (msg *MsgEnableStrictUnbondingCap) Type() string {
	return TypeMsgEnableStrictUnbondingCap
}

func (msg *MsgEnableStrictUnbondingCap) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgEnableStrictUnbondingCap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgEnableStrictUnbondingCap) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
