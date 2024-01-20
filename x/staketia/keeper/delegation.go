package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/utils"
	stakeibctypes "github.com/Stride-Labs/stride/v17/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Liquid stakes native tokens and returns stTokens to the user
// The staker's native tokens (which exist as an IBC denom on stride) are escrowed
// in the deposit account
// StTokens are minted at the current redemption rate
func (k Keeper) LiquidStake(ctx sdk.Context, liquidStaker string, nativeAmount sdkmath.Int) (stToken sdk.Coin, err error) {
	// Get the host zone and verify it's unhalted
	hostZone, err := k.GetUnhaltedHostZone(ctx)
	if err != nil {
		return stToken, err
	}

	// Get user and deposit account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(liquidStaker)
	if err != nil {
		return stToken, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return stToken, errorsmod.Wrapf(err, "host zone deposit address is invalid")
	}

	// Check redemption rates are within safety bounds
	if err := k.CheckRedemptionRateExceedsBounds(ctx); err != nil {
		return stToken, err
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeToken := sdk.NewCoin(hostZone.NativeTokenIbcDenom, nativeAmount)
	if !utils.IsIBCToken(hostZone.NativeTokenIbcDenom) {
		return stToken, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken,
			"denom is not an IBC token (%s)", hostZone.NativeTokenIbcDenom)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdk.NewDecFromInt(nativeAmount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return stToken, errorsmod.Wrapf(stakeibctypes.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", nativeAmount.String(), hostZone.NativeTokenDenom)
	}

	// Transfer the native tokens from the user to module account
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "failed to send tokens from liquid staker %s to deposit address", liquidStaker)
	}

	// Mint the stTokens and transfer them to the user
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
	stToken = sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "Failed to mint stTokens")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stToken)); err != nil {
		return stToken, errorsmod.Wrapf(err, "Failed to send %s from deposit address to liquid staker", stToken.String())
	}

	// Emit liquid stake event with the same schema as stakeibc
	EmitSuccessfulLiquidStakeEvent(ctx, liquidStaker, hostZone, nativeAmount, stAmount)

	return stToken, nil
}

// TODO [sttia]
func (k Keeper) PrepareDelegation(ctx sdk.Context, epochNumber uint64) error {
	return nil
}
