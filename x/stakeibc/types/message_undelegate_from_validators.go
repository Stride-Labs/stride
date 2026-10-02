// x/stakeibc/types/message_undelegate_from_validators.go
package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgUndelegateFromValidators = "undelegate_from_validators"

var _ sdk.Msg = &MsgUndelegateFromValidators{}

func NewMsgUndelegateFromValidators(creator, chainId string, validators []ValidatorUndelegation) *MsgUndelegateFromValidators {
	return &MsgUndelegateFromValidators{
		Creator:    creator,
		ChainId:    chainId,
		Validators: validators,
	}
}

func (msg *MsgUndelegateFromValidators) Route() string {
	return RouterKey
}

func (msg *MsgUndelegateFromValidators) Type() string {
	return TypeMsgUndelegateFromValidators
}

func (msg *MsgUndelegateFromValidators) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUndelegateFromValidators) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if len(msg.ChainId) == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "chain-id is required")
	}

	// An empty list is the full drain; a non-empty list must name distinct valopers with
	// non-negative offsets (a nil offset is the zero value from JSON and means zero)
	seen := map[string]bool{}
	for _, validator := range msg.Validators {
		if len(validator.Address) == 0 {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address is required")
		}
		if !strings.Contains(validator.Address, "valoper") {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator address %s must contain 'valoper'", validator.Address)
		}
		if seen[validator.Address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s listed twice", validator.Address)
		}
		seen[validator.Address] = true
		if !validator.Offset.IsNil() && validator.Offset.IsNegative() {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "offset for %s must not be negative", validator.Address)
		}
	}
	return nil
}
