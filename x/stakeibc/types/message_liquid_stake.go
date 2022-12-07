package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgLiquidStake = "liquid_stake"

var _ sdk.Msg = &MsgLiquidStake{}

func NewMsgLiquidStake(creator string, amount sdk.Int, hostDenom string) *MsgLiquidStake {
	return &MsgLiquidStake{
		Creator:   creator,
		Amount:    amount,
		HostDenom: hostDenom,
	}
}

// isIBCToken checks if the token came from the IBC module
// Each IBC token starts with an ibc/ denom, the check is rather simple
func IsIBCToken(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

func StAssetDenomFromHostZoneDenom(hostZoneDenom string) string {
	return "st" + hostZoneDenom
}

func (msg *MsgLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgLiquidStake) Type() string {
	return TypeMsgLiquidStake
}

func (msg *MsgLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLiquidStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLiquidStake) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate amount is positive nonzero
	if msg.Amount.LTE(sdk.ZeroInt()) {
		return sdkerrors.Wrapf(ErrInvalidAmount, "amount liquid staked must be positive and nonzero")
	}
	// validate host denom is not empty
	if msg.HostDenom == "" {
		return sdkerrors.Wrapf(ErrRequiredFieldEmpty, "host denom cannot be empty")
	}
	// host denom must be a valid asset denom
	if err := sdk.ValidateDenom(msg.HostDenom); err != nil {
		return err
	}
	return nil
}
