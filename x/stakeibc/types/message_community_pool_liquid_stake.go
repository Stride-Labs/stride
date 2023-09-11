package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCommunityPoolLiquidStake = "community_pool_liquid_stake"

var _ sdk.Msg = &MsgCommunityPoolLiquidStake{}

func NewMsgCommunityPoolLiquidStake(creator string, communityPoolHostZoneId string, amount sdkmath.Int, tokenDenom string) *MsgCommunityPoolLiquidStake {
	return &MsgCommunityPoolLiquidStake{
		Creator: 		creator,	
		ChainId:		communityPoolHostZoneId,
		Amount:        	amount,
		TokenDenom: 	tokenDenom,
	}
}

func (msg *MsgCommunityPoolLiquidStake) Route() string {
	return RouterKey
}

func (msg *MsgCommunityPoolLiquidStake) Type() string {
	return TypeMsgCommunityPoolLiquidStake
}

func (msg *MsgCommunityPoolLiquidStake) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCommunityPoolLiquidStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCommunityPoolLiquidStake) ValidateBasic() error {
	// ensure amount is a nonzero positive integer, maybe add threshold check in future to avoid ddos
	if msg.Amount.LTE(sdkmath.ZeroInt()) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount (%v)", msg.Amount)
	}
	// validate chain id is not empty
	if msg.ChainId == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "chain id cannot be empty")
	}	
	// validate host denom is not empty
	if msg.TokenDenom == "" {
		return errorsmod.Wrapf(ErrRequiredFieldEmpty, "token denom cannot be empty")
	}
	// token denom must be a valid asset denom matching regex
	if err := sdk.ValidateDenom(msg.TokenDenom); err != nil {
		return errorsmod.Wrapf(err, "invalid token denom")
	}
	return nil
}
