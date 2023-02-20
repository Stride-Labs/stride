package types

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"
)

const TypeMsgAddOracle = "add_oracle"

var _ sdk.Msg = &MsgAddOracle{}

func NewMsgAddOracle(creator string, connectionId string) *MsgAddOracle {
	return &MsgAddOracle{
		Creator:      creator,
		ConnectionId: connectionId,
	}
}

func (msg *MsgAddOracle) Route() string {
	return RouterKey
}

func (msg *MsgAddOracle) Type() string {
	return TypeMsgAddOracle
}

func (msg *MsgAddOracle) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddOracle) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	matched, err := regexp.MatchString(`^connection-\d+$`, msg.ConnectionId)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify connnection-id (%s)", msg.ConnectionId)
	}
	if !matched {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid connection-id (%s), must be of the format 'connection-{N}'", msg.ConnectionId)
	}

	return nil
}
