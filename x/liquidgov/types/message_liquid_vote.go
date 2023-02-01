package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

const (
	TypeMsgLiquidVote = "liquid_vote"
)

var (
	_ sdk.Msg = &MsgLiquidVote{}
)

func NewMsgLiquidVote(creator sdk.AccAddress, proposalID uint64, option govtypesv1.VoteOption, hostZoneChainId string) *MsgLiquidVote {
	return &MsgLiquidVote{creator.String(), proposalID, option, hostZoneChainId}
}

func (msg *MsgLiquidVote) Route() string {
	return RouterKey
}

func (msg *MsgLiquidVote) Type() string {
	return TypeMsgLiquidVote
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
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// validate vote option
	if !govtypesv1.ValidVoteOption(msg.VoteOption) {
		return sdkerrors.Wrap(govtypes.ErrInvalidVote, msg.VoteOption.String())
	}
	// validate host denom is not empty
	if msg.HostZoneChainId == "" {
		return sdkerrors.Wrapf(stakeibctypes.ErrRequiredFieldEmpty, "host zone chain id cannot be empty")
	}

	return nil
}
