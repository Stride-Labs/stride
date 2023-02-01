package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/utils"
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
		return fmt.Errorf("invalid creator address (%s): %s", err, ErrRequiredFieldEmpty.Error())
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if (msg.NumRebalance < 1) || (msg.NumRebalance > 10) {
		return fmt.Errorf("invalid number of validators to rebalance (%d): %s", msg.NumRebalance, ErrInvalidRequest.Error())
	}
	return nil
}
