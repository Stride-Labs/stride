package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateValSharesExchRate = "update_val_shares_exch_rate"

var _ sdk.Msg = &MsgUpdateValSharesExchRate{}

func NewMsgUpdateValSharesExchRate(creator string, chainid string, valoper string) *MsgUpdateValSharesExchRate {
	return &MsgUpdateValSharesExchRate{
		Creator: creator,
		ChainId: chainid,
		Valoper: valoper,
	}
}

func (msg *MsgUpdateValSharesExchRate) Route() string {
	return RouterKey
}

func (msg *MsgUpdateValSharesExchRate) Type() string {
	return TypeMsgUpdateValSharesExchRate
}

func (msg *MsgUpdateValSharesExchRate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateValSharesExchRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateValSharesExchRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// basic checks on host denom
	if len(msg.ChainId) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chainid is required")
	}
	// basic checks on host zone
	if len(msg.Valoper) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "valoper is required")
	}
	if !strings.Contains(msg.Valoper, "valoper") {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must contrain 'valoper'")
	}

	return nil
}
