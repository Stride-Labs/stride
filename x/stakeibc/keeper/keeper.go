package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/spf13/cast"

	icacallbackskeeper "github.com/Stride-Labs/stride/v18/x/icacallbacks/keeper"
	icqkeeper "github.com/Stride-Labs/stride/v18/x/interchainquery/keeper"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v18/x/records/keeper"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                   codec.BinaryCodec
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		paramstore            paramtypes.Subspace
		authority             string
		ICAControllerKeeper   icacontrollerkeeper.Keeper
		IBCKeeper             ibckeeper.Keeper
		bankKeeper            bankkeeper.Keeper
		AccountKeeper         types.AccountKeeper
		InterchainQueryKeeper icqkeeper.Keeper
		RecordsKeeper         recordsmodulekeeper.Keeper
		StakingKeeper         stakingkeeper.Keeper
		ICACallbacksKeeper    icacallbackskeeper.Keeper
		hooks                 types.StakeIBCHooks
		RatelimitKeeper       types.RatelimitKeeper
		ICAOracleKeeper       types.ICAOracleKeeper
		ConsumerKeeper        types.ConsumerKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	interchainQueryKeeper icqkeeper.Keeper,
	RecordsKeeper recordsmodulekeeper.Keeper,
	StakingKeeper stakingkeeper.Keeper,
	ICACallbacksKeeper icacallbackskeeper.Keeper,
	RatelimitKeeper types.RatelimitKeeper,
	icaOracleKeeper types.ICAOracleKeeper,
	ConsumerKeeper types.ConsumerKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:                   cdc,
		storeKey:              storeKey,
		memKey:                memKey,
		paramstore:            ps,
		authority:             authority,
		AccountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		ICAControllerKeeper:   icacontrollerkeeper,
		IBCKeeper:             ibcKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
		RecordsKeeper:         RecordsKeeper,
		StakingKeeper:         StakingKeeper,
		ICACallbacksKeeper:    ICACallbacksKeeper,
		RatelimitKeeper:       RatelimitKeeper,
		ICAOracleKeeper:       icaOracleKeeper,
		ConsumerKeeper:        ConsumerKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the hooks for ibc staking
func (k *Keeper) SetHooks(gh types.StakeIBCHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set ibc staking hooks twice")
	}

	k.hooks = gh

	return k
}

// GetAuthority returns the x/stakeibc module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Searches all interchain accounts and finds the connection ID that corresponds with a given port ID
func (k Keeper) GetConnectionIdFromICAPortId(ctx sdk.Context, portId string) (connectionId string, found bool) {
	icas := k.ICAControllerKeeper.GetAllInterchainAccounts(ctx)
	for _, ica := range icas {
		if ica.PortId == portId {
			return ica.ConnectionId, true
		}
	}
	return "", false
}

func (k Keeper) GetICATimeoutNanos(ctx sdk.Context, epochType string) (uint64, error) {
	epochTracker, found := k.GetEpochTracker(ctx, epochType)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get epoch tracker for %s", epochType))
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to get epoch tracker for %s", epochType)
	}
	// BUFFER by 5% of the epoch length
	bufferSizeParam := k.GetParam(ctx, types.KeyBufferSize)
	bufferSize := epochTracker.Duration / bufferSizeParam
	// buffer size should not be negative or longer than the epoch duration
	if bufferSize > epochTracker.Duration {
		k.Logger(ctx).Error(fmt.Sprintf("Invalid buffer size %d", bufferSize))
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Invalid buffer size %d", bufferSize)
	}
	timeoutNanos := epochTracker.NextEpochStartTime - bufferSize
	timeoutNanosUint64, err := cast.ToUint64E(timeoutNanos)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to convert timeoutNanos to uint64, error: %s", err.Error()))
		return 0, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to convert timeoutNanos to uint64, error: %s", err.Error())
	}
	return timeoutNanosUint64, nil
}

// safety check: ensure the redemption rate is NOT below our min safety threshold && NOT above our max safety threshold on host zone
func (k Keeper) IsRedemptionRateWithinSafetyBounds(ctx sdk.Context, zone types.HostZone) (bool, error) {
	// Get the wide bounds
	minSafetyThreshold, maxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	redemptionRate := zone.RedemptionRate

	if redemptionRate.LT(minSafetyThreshold) || redemptionRate.GT(maxSafetyThreshold) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed %s is outside safety bounds [%s, %s]", redemptionRate.String(), minSafetyThreshold.String(), maxSafetyThreshold.String())
		k.Logger(ctx).Error(errMsg)
		return false, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	// Verify the redemption rate is within the inner safety bounds
	// The inner safety bounds should always be within the safety bounds, but
	// the redundancy above is cheap.
	// There is also one scenario where the outer bounds go within the inner bounds - if they're updated as part of a param change proposal.
	minInnerSafetyThreshold, maxInnerSafetyThreshold := k.GetInnerSafetyBounds(ctx, zone)
	if redemptionRate.LT(minInnerSafetyThreshold) || redemptionRate.GT(maxInnerSafetyThreshold) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed %s is outside inner safety bounds [%s, %s]", redemptionRate.String(), minInnerSafetyThreshold.String(), maxInnerSafetyThreshold.String())
		k.Logger(ctx).Error(errMsg)
		return false, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}

	return true, nil
}

func (k Keeper) GetOuterSafetyBounds(ctx sdk.Context, zone types.HostZone) (sdk.Dec, sdk.Dec) {
	// Fetch the wide bounds
	minSafetyThresholdInt := k.GetParam(ctx, types.KeyDefaultMinRedemptionRateThreshold)
	minSafetyThreshold := sdk.NewDec(int64(minSafetyThresholdInt)).Quo(sdk.NewDec(100))

	if !zone.MinRedemptionRate.IsNil() && zone.MinRedemptionRate.IsPositive() {
		minSafetyThreshold = zone.MinRedemptionRate
	}

	maxSafetyThresholdInt := k.GetParam(ctx, types.KeyDefaultMaxRedemptionRateThreshold)
	maxSafetyThreshold := sdk.NewDec(int64(maxSafetyThresholdInt)).Quo(sdk.NewDec(100))

	if !zone.MaxRedemptionRate.IsNil() && zone.MaxRedemptionRate.IsPositive() {
		maxSafetyThreshold = zone.MaxRedemptionRate
	}

	return minSafetyThreshold, maxSafetyThreshold
}

func (k Keeper) GetInnerSafetyBounds(ctx sdk.Context, zone types.HostZone) (sdk.Dec, sdk.Dec) {
	// Fetch the inner bounds
	minSafetyThreshold, maxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	if !zone.MinInnerRedemptionRate.IsNil() && zone.MinInnerRedemptionRate.IsPositive() && zone.MinInnerRedemptionRate.GT(minSafetyThreshold) {
		minSafetyThreshold = zone.MinInnerRedemptionRate
	}
	if !zone.MaxInnerRedemptionRate.IsNil() && zone.MaxInnerRedemptionRate.IsPositive() && zone.MaxInnerRedemptionRate.LT(maxSafetyThreshold) {
		maxSafetyThreshold = zone.MaxInnerRedemptionRate
	}

	return minSafetyThreshold, maxSafetyThreshold
}
