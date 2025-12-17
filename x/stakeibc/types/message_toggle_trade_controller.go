package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v31/utils"
)

const TypeMsgToggleTradeController = "toggle_trade_controller"

var _ sdk.Msg = &MsgToggleTradeController{}

func NewMsgToggleTradeController(creator, chainId string, permissionChange AuthzPermissionChange, address string, legacy bool) *MsgToggleTradeController {
	return &MsgToggleTradeController{
		Creator:          creator,
		ChainId:          chainId,
		PermissionChange: permissionChange,
		Address:          address,
		Legacy:           legacy,
	}
}

func (msg *MsgToggleTradeController) Route() string {
	return RouterKey
}

func (msg *MsgToggleTradeController) Type() string {
	return TypeMsgToggleTradeController
}

func (msg *MsgToggleTradeController) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgToggleTradeController) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if msg.ChainId == "" {
		return errors.New("chain ID must be specified")
	}
	if msg.Address == "" {
		return errors.New("trade controller address must be specified")
	}
	if _, ok := AuthzPermissionChange_name[int32(msg.PermissionChange)]; !ok {
		return errors.New("invalid permission change enum value")
	}

	return nil
}
