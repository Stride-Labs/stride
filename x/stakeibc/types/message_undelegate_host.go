package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

const TypeMsgUndelegateHost = "undelegate_host"

var _ sdk.Msg = &MsgUndelegateHost{}

func NewMsgUndelegateHost(creator string, amount sdkmath.Int) *MsgUndelegateHost {
	return &MsgUndelegateHost{
		Creator: creator,
		Amount:  amount,
	}
}

func (msg *MsgUndelegateHost) Route() string {
	return RouterKey
}

func (msg *MsgUndelegateHost) Type() string {
	return TypeMsgUndelegateHost
}

func (msg *MsgUndelegateHost) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUndelegateHost) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUndelegateHost) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if msg.Amount.IsZero() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}
	return nil
}
