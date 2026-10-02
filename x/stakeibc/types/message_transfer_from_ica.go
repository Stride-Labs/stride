// x/stakeibc/types/message_transfer_from_ica.go
package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v35/utils"
)

const TypeMsgTransferFromIca = "transfer_from_ica"

var _ sdk.Msg = &MsgTransferFromIca{}

// The four ICAs that hold anything worth moving (spec §3, §7)
var TransferableIcaTypes = map[ICAAccountType]bool{
	ICAAccountType_DELEGATION: true,
	ICAAccountType_WITHDRAWAL: true,
	ICAAccountType_FEE:        true,
	ICAAccountType_REDEMPTION: true,
}

func NewMsgTransferFromIca(creator, chainId string, icaType ICAAccountType, amount sdk.Coin) *MsgTransferFromIca {
	return &MsgTransferFromIca{
		Creator: creator,
		ChainId: chainId,
		IcaType: icaType,
		Amount:  amount,
	}
}

func (msg *MsgTransferFromIca) Route() string {
	return RouterKey
}

func (msg *MsgTransferFromIca) Type() string {
	return TypeMsgTransferFromIca
}

func (msg *MsgTransferFromIca) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferFromIca) ValidateBasic() error {
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
	if !TransferableIcaTypes[msg.IcaType] {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "ica type %s is not one of DELEGATION, WITHDRAWAL, FEE, REDEMPTION", msg.IcaType)
	}
	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount: %s", err)
	}
	if !msg.Amount.IsPositive() {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "amount must be greater than 0")
	}
	return nil
}
