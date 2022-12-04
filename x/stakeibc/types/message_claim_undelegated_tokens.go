package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/utils"
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
		return fmt.Errorf("invalid creator address (%s): invalid address", err.Error())
	}
	// sender must be a valid stride address
	_, err = utils.AccAddressFromBech32(msg.Sender, "stride")
	if err != nil {
		return fmt.Errorf("invalid sender address (%s): invalid address", err.Error())
	}
	// validate host denom is not empty
	if msg.HostZoneId == "" {
		return fmt.Errorf("host zone id cannot be empty: %s", ErrRequiredFieldEmpty.Error())
	}
	if !(msg.Epoch < (1<<63 - 1)) {
		return fmt.Errorf("epoch must be less than math.MaxInt64 %d: invalid amount", 1<<63-1)
	}
	return nil
}
