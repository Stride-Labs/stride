package types

import (
	fmt "fmt"
	"github.com/Stride-Labs/stride/utils"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	proto "github.com/gogo/protobuf/proto"
	"strings"
)

const TypeMsgSubmitTx = "submit_tx"

var _ sdk.Msg = &MsgSubmitTx{}

// NewMsgSubmitTx creates and returns a new MsgSubmitTx instance
func NewMsgSubmitTx(owner string, connectionID string, sdkMsg sdk.Msg) (*MsgSubmitTx, error) {
	any, err := PackTxMsgAny(sdkMsg)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitTx{
		ConnectionId: connectionID,
		Owner:        owner,
		Msg:          any,
	}, nil
}

func (msg *MsgSubmitTx) Route() string {
	return RouterKey
}

func (msg *MsgSubmitTx) Type() string {
	return TypeMsgSubmitTx
}

// PackTxMsgAny marshals the sdk.Msg payload to a protobuf Any type
func PackTxMsgAny(sdkMsg sdk.Msg) (*codectypes.Any, error) {
	msg, ok := sdkMsg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", sdkMsg)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return any, nil
}

// UnpackInterfaces implements codectypes.UnpackInterfacesMessage
func (msg *MsgSubmitTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var (
		sdkMsg sdk.Msg
	)

	return unpacker.UnpackAny(msg.Msg, &sdkMsg)
}

// GetTxMsg fetches the cached any message
func (msg *MsgSubmitTx) GetTxMsg() sdk.Msg {
	sdkMsg, ok := msg.Msg.GetCachedValue().(sdk.Msg)
	if !ok {
		return nil
	}

	return sdkMsg
}

// GetSigners implements sdk.Msg
func (msg *MsgSubmitTx) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{accAddr}
}

func (msg *MsgSubmitTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements sdk.Msg
func (msg *MsgSubmitTx) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid owner address")
	}
	// connectionid should not be empty and should begin with "connection"
	if msg.ConnectionId == "" || !strings.HasPrefix(msg.ConnectionId, "connection") {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid connection id (can't be empty and must begin with connection)")
	}
	if err := utils.ValidateAdminAddress(msg.Owner); err != nil {
		return err
	}
	return nil
}
