package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgDepositVotingStake = "deposit_voting_stake"

var _ sdk.Msg = &MsgDepositVotingStake{}

func NewMsgDepositVotingStake(creator string, amount sdkmath.Int, denom string) *MsgDepositVotingStake {
	return &MsgDepositVotingStake{
		Creator: creator,
		Amount: amount,
		Denom: denom,
	}
}

// Helper to get original backing Token type from a given stToken
func HostZoneDenomFromStAssetDenom(denom string) string {
	prefix := denom[0:1]
	if prefix == "st" {
		return denom[2:]
	}
	return denom
}

func (msg *MsgDepositVotingStake) Route() string {
	return RouterKey
}

func (msg *MsgDepositVotingStake) Type() string {
	return TypeMsgDepositVotingStake
}

func (msg *MsgDepositVotingStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDepositVotingStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositVotingStake) ValidateBasic() error {
	return nil
}
