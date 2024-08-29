package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v24/utils"
)

const TypeMsgResumeHostZone = "resume_host_zone"

var _ sdk.Msg = &MsgResumeHostZone{}

func NewMsgResumeHostZone(creator string, chainId string) *MsgResumeHostZone {
	return &MsgResumeHostZone{
		Creator: creator,
		ChainId: chainId,
	}
}

func (msg *MsgResumeHostZone) Route() string {
	return RouterKey
}

func (msg *MsgResumeHostZone) Type() string {
	return TypeMsgResumeHostZone
}

func (msg *MsgResumeHostZone) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgResumeHostZone) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgResumeHostZone) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	return nil
}
