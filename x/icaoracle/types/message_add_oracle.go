package types

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v5/utils"
)

const TypeMsgAddOracle = "add_oracle"

var _ sdk.Msg = &MsgAddOracle{}

func NewMsgAddOracle(creator string, moniker string, connectionId string, contractCodeId uint64) *MsgAddOracle {
	return &MsgAddOracle{
		Creator:        creator,
		Moniker:        moniker,
		ConnectionId:   connectionId,
		ContractCodeId: contractCodeId,
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

	if msg.Moniker == "" {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "moniker is required")
	}
	if strings.Contains(msg.Moniker, " ") {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "moniker cannot contain any spaces")
	}

	matched, err := regexp.MatchString(`^connection-\d+$`, msg.ConnectionId)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify connnection-id (%s)", msg.ConnectionId)
	}
	if !matched {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid connection-id (%s), must be of the format 'connection-{N}'", msg.ConnectionId)
	}

	if msg.ContractCodeId == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract code-id cannot be 0")
	}

	return nil
}
