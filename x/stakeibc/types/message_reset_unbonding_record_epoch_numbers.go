package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"
)

const TypeMsgResetUnbondingRecordEpochNumbers = "reset_unbonding_record_epoch_numbers"

var _ sdk.Msg = &MsgResetUnbondingRecordEpochNumbers{}

func NewMsgResetUnbondingRecordEpochNumbers(creator string) *MsgResetUnbondingRecordEpochNumbers {
	return &MsgResetUnbondingRecordEpochNumbers{
		Creator: creator,
	}
}

func (msg *MsgResetUnbondingRecordEpochNumbers) Route() string {
	return RouterKey
}

func (msg *MsgResetUnbondingRecordEpochNumbers) Type() string {
	return TypeMsgResetUnbondingRecordEpochNumbers
}

func (msg *MsgResetUnbondingRecordEpochNumbers) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgResetUnbondingRecordEpochNumbers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResetUnbondingRecordEpochNumbers) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}

	return nil
}
