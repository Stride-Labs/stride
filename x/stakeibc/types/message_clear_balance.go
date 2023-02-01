package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	"github.com/Stride-Labs/stride/v5/utils"
)

const TypeMsgClearBalance = "clear_balance"

var _ sdk.Msg = &MsgClearBalance{}

func NewMsgClearBalance(creator string, chainId string, amount sdkmath.Int, channelId string) *MsgClearBalance {
	return &MsgClearBalance{
		Creator: creator,
		ChainId: chainId,
		Amount:  amount,
		Channel: channelId,
	}
}

func (msg *MsgClearBalance) Route() string {
	return RouterKey
}

func (msg *MsgClearBalance) Type() string {
	return TypeMsgClearBalance
}

func (msg *MsgClearBalance) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgClearBalance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClearBalance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address (%s): %s", err, ErrInvalidAddress.Error())
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	// basic checks on host denom
	if len(msg.ChainId) == 0 {
		return fmt.Errorf("chainid is required: %s", ErrInvalidRequest.Error())
	}

	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return fmt.Errorf("amount must be greater than 0: %s", ErrInvalidRequest.Error())
	}
	if isValid := channeltypes.IsValidChannelID(msg.Channel); !isValid {
		return fmt.Errorf("channel is invalid: %s", ErrInvalidRequest.Error())
	}
	return nil
}
