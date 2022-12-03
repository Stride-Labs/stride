package keeper

import (
	"fmt"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	"github.com/Stride-Labs/stride/v4/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	epochIdentifier := epochInfo.Identifier
	epochNumber := epochInfo.CurrentEpoch
	params := k.GetParams(ctx)

	if epochIdentifier == params.EpochIdentifier {
		// not distribute rewards if it's not time yet for rewards distribution
		if epochNumber < params.MintingRewardsDistributionStartEpoch {
			return
		} else if epochNumber == params.MintingRewardsDistributionStartEpoch {
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}
		// fetch stored minter & params
		minter := k.GetMinter(ctx)
		params := k.GetParams(ctx)

		// Check if we have hit an epoch where we update the inflation parameter.
		// Since epochs only update based on BFT time data, it is safe to store the "reductioning period time"
		// in terms of the number of epochs that have transpired.
		if epochNumber >= k.GetParams(ctx).ReductionPeriodInEpochs+k.GetLastReductionEpochNum(ctx) {
			// reduction the reward per reduction period
			minter.EpochProvisions = minter.NextEpochProvisions(params)
			k.SetMinter(ctx, minter)
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}

		// mint coins, update supply
		mintedCoin := minter.EpochProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to their respective module accounts (e.g. staking rewards to the feecollector)
		err = k.DistributeMintedCoin(ctx, mintedCoin)
		if err != nil {
			panic(err)
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
				sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}
