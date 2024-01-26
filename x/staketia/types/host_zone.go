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

	// Validate the redemption rate bounds are set properly
	if !h.RedemptionRate.IsPositive() {
		return ErrInvalidHostZone.Wrap("redemption rate must be positive")
	}
	if err := h.ValidateRedemptionRateBoundsInitalized(); err != nil {
		return err
	}

	// Validate unbonding period is set
	if h.UnbondingPeriodSeconds == 0 {
		return ErrInvalidHostZone.Wrap("unbonding period must be set")
	}

	return nil
}

// Verify the redemption rate bounds are set properly on the host zone
func (h HostZone) ValidateRedemptionRateBoundsInitalized() error {
	// Validate outer bounds are set
	if h.MinRedemptionRate.IsNil() || !h.MinRedemptionRate.IsPositive() {
		return ErrInvalidRedemptionRateBounds.Wrapf("min outer redemption rate bound not set")
	}
	if h.MaxRedemptionRate.IsNil() || !h.MaxRedemptionRate.IsPositive() {
		return ErrInvalidRedemptionRateBounds.Wrapf("max outer redemption rate bound not set")
	}

	// Validate inner bounds set
	if h.MinInnerRedemptionRate.IsNil() || !h.MinInnerRedemptionRate.IsPositive() {
		return ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate bound not set")
	}
	if h.MaxInnerRedemptionRate.IsNil() || !h.MaxInnerRedemptionRate.IsPositive() {
		return ErrInvalidRedemptionRateBounds.Wrapf("max inner redemption rate bound not set")
	}

	// Validate inner bounds are within outer bounds
	if h.MinInnerRedemptionRate.LT(h.MinRedemptionRate) {
		return ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate bound outside of min outer bound")
	}
	if h.MaxInnerRedemptionRate.GT(h.MaxRedemptionRate) {
		return ErrInvalidRedemptionRateBounds.Wrapf("max inner redemption rate bound outside of max outer bound")
	}
	if h.MinInnerRedemptionRate.GT(h.MaxInnerRedemptionRate) {
		return ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate greater than max inner bound")
	}

	return nil
}
