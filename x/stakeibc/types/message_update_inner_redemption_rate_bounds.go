package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v25/utils"
)

const TypeMsgUpdateRedemptionRateBounds = "update_redemption_rate_bounds"

var _ sdk.Msg = &MsgUpdateRedemptionRateBounds{}

func NewMsgUpdateRedemptionRateBounds(creator string, chainId string, minInnerRedemptionRate sdk.Dec, maxInnerRedemptionRate sdk.Dec) *MsgUpdateRedemptionRateBounds {
	return &MsgUpdateRedemptionRateBounds{
		Creator:                creator,
		ChainId:                chainId,
		MinInnerRedemptionRate: minInnerRedemptionRate,
		MaxInnerRedemptionRate: maxInnerRedemptionRate,
	}
}

func (msg *MsgUpdateRedemptionRateBounds) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRedemptionRateBounds) Type() string {
	return TypeMsgUpdateRedemptionRateBounds
}

func (msg *MsgUpdateRedemptionRateBounds) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRedemptionRateBounds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRedemptionRateBounds) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// Confirm the max is greater than the min
	if msg.MaxInnerRedemptionRate.LTE(msg.MinInnerRedemptionRate) {
		return errorsmod.Wrapf(ErrInvalidBounds, "Inner max safety threshold (%s) is less than inner min safety threshold (%s)", msg.MaxInnerRedemptionRate, msg.MinInnerRedemptionRate)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
