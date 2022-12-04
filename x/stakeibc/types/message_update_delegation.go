package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/utils"
)

const TypeMsgUpdateValidatorSharesExchRate = "update_validator_shares_exch_rate"

var _ sdk.Msg = &MsgUpdateValidatorSharesExchRate{}

func NewMsgUpdateValidatorSharesExchRate(creator string, chainid string, valoper string) *MsgUpdateValidatorSharesExchRate {
	return &MsgUpdateValidatorSharesExchRate{
		Creator: creator,
		ChainId: chainid,
		Valoper: valoper,
	}
}

func (msg *MsgUpdateValidatorSharesExchRate) Route() string {
	return RouterKey
}

func (msg *MsgUpdateValidatorSharesExchRate) Type() string {
	return TypeMsgUpdateValidatorSharesExchRate
}

func (msg *MsgUpdateValidatorSharesExchRate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateValidatorSharesExchRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateValidatorSharesExchRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s): invalid address", err.Error())
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// basic checks on host denom
	if len(msg.ChainId) == 0 {
		return fmt.Errorf("chainid is required: invalid address")
	}
	// basic checks on host zone
	if len(msg.Valoper) == 0 {
		return fmt.Errorf("valoper is required: invalid address")
	}
	if !strings.Contains(msg.Valoper, "valoper") {
		return fmt.Errorf("validator operator address must contrain 'valoper': invalid address")
	}

	return nil
}
