package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const TypeMsgLiquidVote = "liquid_vote"

var _ sdk.Msg = &MsgLiquidVote{}

func NewMsgLiquidVote(creator string, host_zone_id string, proposal_id uint64, amount sdkmath.Int, vote_option govtypes.VoteOption) *MsgLiquidVote {
	return &MsgLiquidVote{
		Creator: creator,
		HostZoneId: host_zone_id,
		ProposalId: proposal_id,
		Amount: amount,
		VoteOption: vote_option,
	}
}

func (msg *MsgLiquidVote) Route() string {
	return RouterKey
}

func (msg *MsgLiquidVote) Type() string {
	return TypeMsgDepositVotingStake
}

func (msg *MsgLiquidVote) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgLiquidVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLiquidVote) ValidateBasic() error {
	return nil
}
