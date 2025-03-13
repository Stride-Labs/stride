package types

import (
	"errors"
	time "time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Validates the airdrop creation or update message
func AirdropConfigValidateBasic(
	airdropId string,
	rewardDenom string,
	distributionStartDate *time.Time,
	distributionEndDate *time.Time,
	clawbackDate *time.Time,
	claimTypeDeadlineDate *time.Time,
	earlyClaimPenalty sdk.Dec,
	distributorAddress string,
	allocatorAddress string,
	linkerAddress string,
) error {
	if airdropId == "" {
		return errors.New("airdrop-id must be specified")
	}
	if rewardDenom == "" {
		return errors.New("reward denom must be specified")
	}

	if distributionStartDate == nil || *distributionStartDate == (time.Time{}) {
		return errors.New("distribution start date must be specified")
	}
	if distributionEndDate == nil || *distributionEndDate == (time.Time{}) {
		return errors.New("distribution end date must be specified")
	}
	if clawbackDate == nil || *clawbackDate == (time.Time{}) {
		return errors.New("clawback date must be specified")
	}
	if claimTypeDeadlineDate == nil || *claimTypeDeadlineDate == (time.Time{}) {
		return errors.New("claim type deadline date must be specified")
	}

	if !distributionEndDate.After(*distributionStartDate) {
		return errors.New("distribution end date must be after the start date")
	}
	if !claimTypeDeadlineDate.After(*distributionStartDate) {
		return errors.New("claim type deadline date must be after the distribution start date")
	}
	if !claimTypeDeadlineDate.Before(*distributionEndDate) {
		return errors.New("claim type deadline date must be before the distribution end date")
	}
	if !clawbackDate.After(*distributionEndDate) {
		return errors.New("clawback date must be after the distribution end date")
	}

	if earlyClaimPenalty.IsNil() {
		return errors.New("early claim penalty must be specified")
	}
	if earlyClaimPenalty.LT(sdkmath.LegacyZeroDec()) || earlyClaimPenalty.GT(sdkmath.LegacyOneDec()) {
		return errors.New("early claim penalty must be between 0 and 1")
	}

	if _, err := sdk.AccAddressFromBech32(distributorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid distributor address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(allocatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid allocator address (%s)", err)
	}
	if _, err := sdk.AccAddressFromBech32(linkerAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid linker address (%s)", err)
	}

	return nil
}
