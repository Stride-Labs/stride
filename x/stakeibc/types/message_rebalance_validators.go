package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgRebalanceValidators = "rebalance_validators"

var _ sdk.Msg = &MsgRebalanceValidators{}

func NewMsgRebalanceValidators(creator string, hostZone string, numValidators uint64) *MsgRebalanceValidators {
	return &MsgRebalanceValidators{
		Creator:      creator,
		HostZone:     hostZone,
		NumRebalance: numValidators,
	}
}

func (msg *MsgRebalanceValidators) Route() string {
	return RouterKey
}

func (msg *MsgRebalanceValidators) Type() string {
	return TypeMsgRebalanceValidators
}

func (msg *MsgRebalanceValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRebalanceValidators) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRebalanceValidators) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if (msg.NumRebalance < 1) || (msg.NumRebalance > 10) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, fmt.Sprintf("invalid number of validators to rebalance (%d)", msg.NumRebalance))
	}
	return nil
}
