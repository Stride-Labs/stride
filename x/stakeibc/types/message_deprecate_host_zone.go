package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v28/utils"
)

const TypeMsgDeprecateHostZone = "deprecate_host_zone"

var _ sdk.Msg = &MsgDeprecateHostZone{}

func NewMsgDeprecateHostZone(authority string, chainId string) *MsgDeprecateHostZone {
	return &MsgDeprecateHostZone{
		Authority: authority,
		ChainId:   chainId,
	}
}

func (msg *MsgDeprecateHostZone) Route() string {
	return RouterKey
}

func (msg *MsgDeprecateHostZone) Type() string {
	return TypeMsgDeprecateHostZone
}

func (msg *MsgDeprecateHostZone) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeprecateHostZone) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Authority); err != nil {
		return err
	}
	return nil
}
