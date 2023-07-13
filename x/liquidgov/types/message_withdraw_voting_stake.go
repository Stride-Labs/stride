package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdrawVotingStake = "withdraw_voting_stake"

var _ sdk.Msg = &MsgWithdrawVotingStake{}

func NewMsgWithdrawVotingStake(creator string, amount sdkmath.Int, denom string) *MsgWithdrawVotingStake {
	return &MsgWithdrawVotingStake{
		Creator: creator,
		Amount: amount,
		Denom: denom,
	}
}

func (msg *MsgWithdrawVotingStake) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawVotingStake) Type() string {
	return TypeMsgWithdrawVotingStake
}

func (msg *MsgWithdrawVotingStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdrawVotingStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdrawVotingStake) ValidateBasic() error {
	return nil
}
