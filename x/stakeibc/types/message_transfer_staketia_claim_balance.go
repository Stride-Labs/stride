// x/stakeibc/types/message_transfer_staketia_claim_balance.go
package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
)

const TypeMsgTransferStaketiaClaimBalance = "transfer_staketia_claim_balance"

var _ sdk.Msg = &MsgTransferStaketiaClaimBalance{}

// amount is in utia; zero means the whole balance
func NewMsgTransferStaketiaClaimBalance(creator string, amount sdkmath.Int) *MsgTransferStaketiaClaimBalance {
	return &MsgTransferStaketiaClaimBalance{
		Creator: creator,
		Amount:  amount,
	}
}

func (msg *MsgTransferStaketiaClaimBalance) Route() string {
	return RouterKey
}

func (msg *MsgTransferStaketiaClaimBalance) Type() string {
	return TypeMsgTransferStaketiaClaimBalance
}

func (msg *MsgTransferStaketiaClaimBalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferStaketiaClaimBalance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if !msg.Amount.IsNil() && msg.Amount.IsNegative() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount must not be negative")
	}
	return nil
}
