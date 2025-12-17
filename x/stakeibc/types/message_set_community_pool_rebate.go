package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v31/utils"
)

const TypeMsgSetCommunityPoolRebate = "set_community_pool_rebate"

var _ sdk.Msg = &MsgSetCommunityPoolRebate{}

func NewMsgSetCommunityPoolRebate(
	creator string,
	chainId string,
	rebateRate sdkmath.LegacyDec,
	liquidStakedStTokenAmount sdkmath.Int,
) *MsgSetCommunityPoolRebate {
	return &MsgSetCommunityPoolRebate{
		Creator:                   creator,
		ChainId:                   chainId,
		RebateRate:                rebateRate,
		LiquidStakedStTokenAmount: liquidStakedStTokenAmount,
	}
}

func (msg *MsgSetCommunityPoolRebate) Route() string {
	return RouterKey
}

func (msg *MsgSetCommunityPoolRebate) Type() string {
	return TypeMsgSetCommunityPoolRebate
}

func (msg *MsgSetCommunityPoolRebate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetCommunityPoolRebate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return err
	}
	if msg.ChainId == "" {
		return errors.New("chain ID must be specified")
	}
	if msg.RebateRate.IsNil() || msg.RebateRate.LT(sdkmath.LegacyZeroDec()) || msg.RebateRate.GT(sdkmath.LegacyOneDec()) {
		return errors.New("invalid rebate rate, must be a decimal between 0 and 1 (inclusive)")
	}
	if msg.LiquidStakedStTokenAmount.IsNil() || msg.LiquidStakedStTokenAmount.LT(sdkmath.ZeroInt()) {
		return errors.New("invalid liquid stake amount, must be greater than or equal to zero")
	}

	return nil
}
