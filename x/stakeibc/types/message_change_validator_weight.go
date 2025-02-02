package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v25/utils"
)

const TypeMsgChangeValidatorWeights = "change_validator_weight"

var _ sdk.Msg = &MsgChangeValidatorWeights{}

func NewMsgChangeValidatorWeights(creator, hostZone string, weights []*ValidatorWeight) *MsgChangeValidatorWeights {
	return &MsgChangeValidatorWeights{
		Creator:          creator,
		HostZone:         hostZone,
		ValidatorWeights: weights,
	}
}

func (msg *MsgChangeValidatorWeights) Route() string {
	return RouterKey
}

func (msg *MsgChangeValidatorWeights) Type() string {
	return TypeMsgChangeValidatorWeights
}

func (msg *MsgChangeValidatorWeights) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgChangeValidatorWeights) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgChangeValidatorWeights) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if msg.HostZone == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone must be specified")
	}
	if len(msg.ValidatorWeights) < 1 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "at least one validator must be specified")
	}
	for _, weightUpdate := range msg.ValidatorWeights {
		if weightUpdate.Address == "" {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address must be specified")
		}
	}
	return nil
}
