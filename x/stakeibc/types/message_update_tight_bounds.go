package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

const TypeMsgUpdateTightBounds = "update_tight_bounds"

var _ sdk.Msg = &MsgUpdateTightBounds{}

func NewMsgUpdateTightBounds(creator string, chainId string, minTightRedemptionRate sdk.Dec, maxTightRedemptionRate sdk.Dec) *MsgUpdateTightBounds {
	return &MsgUpdateTightBounds{
		Creator:                creator,
		ChainId:                chainId,
		MinTightRedemptionRate: minTightRedemptionRate,
		MaxTightRedemptionRate: maxTightRedemptionRate,
	}
}

func (msg *MsgUpdateTightBounds) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTightBounds) Type() string {
	return TypeMsgUpdateTightBounds
}

func (msg *MsgUpdateTightBounds) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateTightBounds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTightBounds) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
