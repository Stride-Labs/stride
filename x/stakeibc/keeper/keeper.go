package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	icqkeeper "github.com/Stride-Labs/stride/v15/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v15/x/stakeibc/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	epochstypes "github.com/Stride-Labs/stride/v15/x/epochs/types"
	icacallbackskeeper "github.com/Stride-Labs/stride/v15/x/icacallbacks/keeper"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v15/x/records/keeper"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                   codec.BinaryCodec
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		paramstore            paramtypes.Subspace
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

func (k Keeper) GetChainID(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		errMsg := fmt.Sprintf("client id %s not found for connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		errMsg := fmt.Sprintf("invalid client state for client %s on connection %s", conn.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	return client.ChainId, nil
}

func (k Keeper) GetCounterpartyChainId(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		errMsg := fmt.Sprintf("invalid connection id, %s not found", connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	counterPartyClientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.Counterparty.ClientId)
	if !found {
		errMsg := fmt.Sprintf("counterparty client id %s not found for connection %s", conn.Counterparty.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}
	counterpartyClient, ok := counterPartyClientState.(*ibctmtypes.ClientState)
	if !ok {
		errMsg := fmt.Sprintf("invalid client state for client %s on connection %s", conn.Counterparty.ClientId, connectionID)
		k.Logger(ctx).Error(errMsg)
		return "", fmt.Errorf(errMsg)
	}

	return counterpartyClient.ChainId, nil
}

func (k Keeper) GetConnectionId(ctx sdk.Context, portId string) (string, error) {
	icas := k.ICAControllerKeeper.GetAllInterchainAccounts(ctx)
	for _, ica := range icas {
		if ica.PortId == portId {
			return ica.ConnectionId, nil
		}
	}
	errMsg := fmt.Sprintf("portId %s has no associated connectionId", portId)
	k.Logger(ctx).Error(errMsg)
	return "", fmt.Errorf(errMsg)
}

// helper to get what share of the curr epoch we're through
func (k Keeper) GetStrideEpochElapsedShare(ctx sdk.Context) (sdk.Dec, error) {
	// Get the current stride epoch
	epochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		errMsg := fmt.Sprintf("Failed to get epoch tracker for %s", epochstypes.STRIDE_EPOCH)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(sdkerrors.ErrNotFound, errMsg)
	}

	// Get epoch start time, end time, and duration
	epochDuration, err := cast.ToInt64E(epochTracker.Duration)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert epoch duration to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrIntCast, errMsg)
	}
	epochEndTime, err := cast.ToInt64E(epochTracker.NextEpochStartTime)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert next epoch start time to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrIntCast, errMsg)
	}
	epochStartTime := epochEndTime - epochDuration

	// Confirm the current block time is inside the current epoch's start and end times
	currBlockTime := ctx.BlockTime().UnixNano()
	if currBlockTime < epochStartTime || currBlockTime > epochEndTime {
		errMsg := fmt.Sprintf("current block time %d is not within current epoch (ending at %d)", currBlockTime, epochTracker.NextEpochStartTime)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrInvalidEpoch, errMsg)
	}

	// Get elapsed share
	elapsedTime := currBlockTime - epochStartTime
	elapsedShare := sdk.NewDec(elapsedTime).Quo(sdk.NewDec(epochDuration))
	if elapsedShare.LT(sdk.ZeroDec()) || elapsedShare.GT(sdk.OneDec()) {
		errMsg := fmt.Sprintf("elapsed share (%s) for epoch is not between 0 and 1", elapsedShare)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), errorsmod.Wrapf(types.ErrInvalidEpoch, errMsg)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Epoch elapsed share: %v (Block Time: %d, Epoch End Time: %d)", elapsedShare, currBlockTime, epochEndTime))
	return elapsedShare, nil
}

// helper to check whether ICQs are valid in this portion of the epoch
func (k Keeper) IsWithinBufferWindow(ctx sdk.Context) (bool, error) {
	elapsedShareOfEpoch, err := k.GetStrideEpochElapsedShare(ctx)
	if err != nil {
		return false, err
	}
	bufferSize, err := cast.ToInt64E(k.GetParam(ctx, types.KeyBufferSize))
	if err != nil {
		return false, err
	}
	epochShareThresh := sdk.NewDec(1).Sub(sdk.NewDec(1).Quo(sdk.NewDec(bufferSize)))

	inWindow := elapsedShareOfEpoch.GT(epochShareThresh)
	if !inWindow {
		k.Logger(ctx).Error(fmt.Sprintf("Outside ICQ Callback Window. We're %d pct through the epoch, ICQ cutoff is %d", elapsedShareOfEpoch, epochShareThresh))
	}
	return inWindow, nil
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
