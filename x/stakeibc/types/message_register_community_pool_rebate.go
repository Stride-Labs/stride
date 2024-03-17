package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v19/utils"
)

const TypeMsgRegisterCommunityPoolRebate = "register_community_pool_rebate"

var _ sdk.Msg = &MsgRegisterCommunityPoolRebate{}

func NewMsgRegisterCommunityPoolRebate(
	creator string,
	chainId string,
	rebatePercentage sdk.Dec,
	liquidStakedAmount sdkmath.Int,
) *MsgRegisterCommunityPoolRebate {
	return &MsgRegisterCommunityPoolRebate{
		Creator:            creator,
		ChainId:            chainId,
		RebatePercentage:   rebatePercentage,
		LiquidStakedAmount: liquidStakedAmount,
	}
}

func (msg *MsgRegisterCommunityPoolRebate) Route() string {
	return RouterKey
}

func (msg *MsgRegisterCommunityPoolRebate) Type() string {
	return TypeMsgRegisterCommunityPoolRebate
}

func (msg *MsgRegisterCommunityPoolRebate) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRegisterCommunityPoolRebate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterCommunityPoolRebate) ValidateBasic() error {
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
	if msg.RebatePercentage.IsNil() || msg.RebatePercentage.LT(sdk.ZeroDec()) || msg.RebatePercentage.GT(sdk.OneDec()) {
		return errors.New("invalid rebate percentage, must be between 0 and 1 (inclusive)")
	}
	if msg.LiquidStakedAmount.IsNil() || msg.LiquidStakedAmount.LT(sdkmath.ZeroInt()) {
		return errors.New("invalid liquid stake amount, must be greater than or equal to zero")
	}

	return nil
}
