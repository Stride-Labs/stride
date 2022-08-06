package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/utils"
)

const TypeMsgQueryDelegation = "query_delegation"

var _ sdk.Msg = &MsgQueryDelegation{}

func NewMsgQueryDelegation(creator string, hostdenom string, valoper string) *MsgQueryDelegation {
	return &MsgQueryDelegation{
		Creator:   creator,
		Hostdenom: hostdenom,
		Valoper:   valoper,
	}
}

func (msg *MsgQueryDelegation) Route() string {
	return RouterKey
}

func (msg *MsgQueryDelegation) Type() string {
	return TypeMsgQueryDelegation
}

func (msg *MsgQueryDelegation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgQueryDelegation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgQueryDelegation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// basic checks on host denom
	if len(msg.Hostdenom) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "host denom is required")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.Hostdenom); err != nil {
		return err
	}
	// basic checks on host zone
	if len(msg.Valoper) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "valoper is required")
	}
	if !strings.Contains(msg.Valoper, "valoper") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must contrain 'valoper'")
	}

	return nil
}
