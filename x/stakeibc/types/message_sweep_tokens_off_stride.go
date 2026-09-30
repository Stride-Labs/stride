package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSweepTokensOffStride = "sweep_tokens_off_stride"

var _ sdk.Msg = &MsgSweepTokensOffStride{}

func NewMsgSweepTokensOffStride(creator string, denoms []string, addresses []string) *MsgSweepTokensOffStride {
	return &MsgSweepTokensOffStride{
		Creator:   creator,
		Denoms:    denoms,
		Addresses: addresses,
	}
}

func (msg *MsgSweepTokensOffStride) Route() string {
	return RouterKey
}

func (msg *MsgSweepTokensOffStride) Type() string {
	return TypeMsgSweepTokensOffStride
}

func (msg *MsgSweepTokensOffStride) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic gates the sweep on the sweep operator (spec §4, §7) and bounds the batch.
// The operator var ships empty and is filled by the release gate; while it is empty the gate
// rejects every signer, so an unconfigured binary can never sweep.
func (msg *MsgSweepTokensOffStride) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if SweepOperatorAddress == "" {
		return ErrSweepOperatorNotConfigured
	}
	if msg.Creator != SweepOperatorAddress {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "creator %s is not the sweep operator", msg.Creator)
	}

	if len(msg.Denoms) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one denom is required")
	}
	seenDenoms := map[string]bool{}
	for _, denom := range msg.Denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "invalid denom %s: %s", denom, err)
		}
		if seenDenoms[denom] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate denom %s", denom)
		}
		seenDenoms[denom] = true
	}

	if len(msg.Addresses) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one address is required")
	}
	if len(msg.Addresses) > MaxSweepAddressesPerTx {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "%d addresses exceeds the batch bound of %d",
			len(msg.Addresses), MaxSweepAddressesPerTx)
	}
	seenAddresses := map[string]bool{}
	for _, address := range msg.Addresses {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address %s: %s", address, err)
		}
		if seenAddresses[address] {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate address %s", address)
		}
		seenAddresses[address] = true
	}
	return nil
}
