package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateProposal = "update_proposal"

var _ sdk.Msg = &MsgUpdateProposal{}

func NewMsgUpdateProposal(creator string, host_zone_id string, proposal_id uint64) *MsgUpdateProposal {
	return &MsgUpdateProposal{
		Creator: creator,
		HostZoneId: host_zone_id,
		ProposalId: proposal_id,
	}
}

func (msg *MsgUpdateProposal) Route() string {
	return RouterKey
}

func (msg *MsgUpdateProposal) Type() string {
	return TypeMsgUpdateProposal
}

func (msg *MsgUpdateProposal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateProposal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateProposal) ValidateBasic() error {
	return nil
}
