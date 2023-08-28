package types

import (
	"regexp"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

const TypeMsgInstantiateOracle = "instantiate_oracle"

var _ sdk.Msg = &MsgInstantiateOracle{}

func NewMsgInstantiateOracle(creator string, chainId string, contractCodeId uint64, transferChannelId string) *MsgInstantiateOracle {
	return &MsgInstantiateOracle{
		Creator:                 creator,
		OracleChainId:           chainId,
		ContractCodeId:          contractCodeId,
		TransferChannelOnOracle: transferChannelId,
	}
}

func (msg MsgInstantiateOracle) Type() string {
	return TypeMsgInstantiateOracle
}

func (msg *MsgInstantiateOracle) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgInstantiateOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgInstantiateOracle) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if err := utils.ValidateAdminAddress(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.OracleChainId == "" {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "oracle-chain-id is required")
	}

	if msg.ContractCodeId == 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract code-id cannot be 0")
	}

	matched, err := regexp.MatchString(`^channel-\d+$`, msg.TransferChannelOnOracle)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to verify channel-id (%s)", msg.TransferChannelOnOracle)
	}
	if !matched {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid channel-id (%s), must be of the format 'channel-{N}'", msg.TransferChannelOnOracle)
	}

	return nil
}
