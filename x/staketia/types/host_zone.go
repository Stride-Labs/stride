package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// Helper fucntion to validate a host zone is properly initialized
// during genesis
func (h HostZone) ValidateGenesis() error {
	// Validate that the chain ID is provided
	if h.ChainId == "" {
		return ErrInvalidHostZone.Wrap("chain-id must be specified")
	}
	if h.TransferChannelId == "" {
		return ErrInvalidHostZone.Wrap("transfer channel-id must be specified")
	}

	// Validate that the token denom's are provided and that the IBC denom matches
	// the hash built from the denom and transfer channel
	if h.NativeTokenDenom == "" {
		return ErrInvalidHostZone.Wrap("native token denom must be specified")
	}
	if h.NativeTokenIbcDenom == "" {
		return ErrInvalidHostZone.Wrap("native token ibc denom must be specified")
	}

	ibcDenomTracePrefix := transfertypes.GetDenomPrefix(transfertypes.PortID, h.TransferChannelId)
	expectedIbcDenom := transfertypes.ParseDenomTrace(ibcDenomTracePrefix + h.NativeTokenDenom).IBCDenom()
	if h.NativeTokenIbcDenom != expectedIbcDenom {
		return ErrInvalidHostZone.Wrapf(
			"native token ibc denom did not match hash generated"+
				"from channel (%s) and denom (%s). Provided: %s, Expected: %s",
			h.TransferChannelId, h.NativeTokenDenom, h.NativeTokenIbcDenom, expectedIbcDenom,
		)
	}

	// Validate all the addresses are provided
	if h.DelegationAddress == "" {
		return ErrInvalidHostZone.Wrap("delegation address must be specified")
	}
	if h.RewardAddress == "" {
		return ErrInvalidHostZone.Wrap("reward address must be specified")
	}
	if h.DepositAddress == "" {
		return ErrInvalidHostZone.Wrap("deposit address must be specified")
	}
	if h.RedemptionAddress == "" {
		return ErrInvalidHostZone.Wrap("redemption address must be specified")
	}
	if h.ClaimAddress == "" {
		return ErrInvalidHostZone.Wrap("claim address must be specified")
	}
	if h.OperatorAddressOnStride == "" {
		return ErrInvalidHostZone.Wrap("operator address must be specified")
	}
	if h.SafeAddressOnStride == "" {
		return ErrInvalidHostZone.Wrap("safe address must be specified")
	}

	// Validate all the stride addresses are valid bech32 addresses
	if _, err := sdk.AccAddressFromBech32(h.DepositAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid deposit address")
	}
	if _, err := sdk.AccAddressFromBech32(h.RedemptionAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid redemption address")
	}
	if _, err := sdk.AccAddressFromBech32(h.ClaimAddress); err != nil {
		return errorsmod.Wrapf(err, "invalid claim address")
	}
	if _, err := sdk.AccAddressFromBech32(h.OperatorAddressOnStride); err != nil {
		return errorsmod.Wrapf(err, "invalid operator address")
	}
	if _, err := sdk.AccAddressFromBech32(h.SafeAddressOnStride); err != nil {
		return errorsmod.Wrapf(err, "invalid safe address")
	}

	// Validate unbonding period is set
	if h.UnbondingPeriodSeconds == 0 {
		return ErrInvalidHostZone.Wrap("unbonding period must be set")
	}

	return nil
}
