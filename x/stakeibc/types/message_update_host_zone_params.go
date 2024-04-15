package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	errorsmod "cosmossdk.io/errors"
)

const TypeMsgUpdateHostZoneParams = "update_host_zone_params"

var (
	_ sdk.Msg            = &MsgUpdateHostZoneParams{}
	_ legacytx.LegacyMsg = &MsgUpdateHostZoneParams{}
)

func (msg *MsgUpdateHostZoneParams) Type() string {
	return TypeMsgUpdateHostZoneParams
}

func (msg *MsgUpdateHostZoneParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateHostZoneParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateHostZoneParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
func (msg *MsgUpdateHostZoneParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}
	if msg.ChainId == "" {
		return errors.New("chain ID must be specified")
	}
	return nil
}
