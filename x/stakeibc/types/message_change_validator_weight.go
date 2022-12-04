package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgChangeValidatorWeight = "change_validator_weight"

var _ sdk.Msg = &MsgChangeValidatorWeight{}

func NewMsgChangeValidatorWeight(creator string, hostZone string, address string, weight uint64) *MsgChangeValidatorWeight {
	return &MsgChangeValidatorWeight{
		Creator:  creator,
		HostZone: hostZone,
		ValAddr:  address,
		Weight:   weight,
	}
}

func (msg *MsgChangeValidatorWeight) Route() string {
	return RouterKey
}

func (msg *MsgChangeValidatorWeight) Type() string {
	return TypeMsgChangeValidatorWeight
}

func (msg *MsgChangeValidatorWeight) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgChangeValidatorWeight) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgChangeValidatorWeight) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s): invalid address", err.Error())
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
