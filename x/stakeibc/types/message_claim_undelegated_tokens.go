package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v9/utils"
)

const TypeMsgClaimUndelegatedTokens = "claim_undelegated_tokens"

var _ sdk.Msg = &MsgClaimUndelegatedTokens{}

func NewMsgClaimUndelegatedTokens(creator string, hostZone string, epoch uint64, sender string) *MsgClaimUndelegatedTokens {
	return &MsgClaimUndelegatedTokens{
		Creator:    creator,
		HostZoneId: hostZone,
		Epoch:      epoch,
		Sender:     sender,
	}
}

func (msg *MsgClaimUndelegatedTokens) Route() string {
	return RouterKey
}

func (msg *MsgClaimUndelegatedTokens) Type() string {
	return TypeMsgClaimUndelegatedTokens
}

func (msg *MsgClaimUndelegatedTokens) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgClaimUndelegatedTokens) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClaimUndelegatedTokens) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// sender must be a valid stride address
	_, err = utils.AccAddressFromBech32(msg.Sender, "stride")
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address (%s)", err)
	}
	// validate host denom is not empty
	if msg.HostZoneId == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "host zone id cannot be empty")
	}
	if !(msg.Epoch < (1<<63 - 1)) {
		return errorsmod.Wrapf(ErrInvalidAmount, "epoch must be less than math.MaxInt64 %d", 1<<63-1)
	}
	return nil
}
