package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cast"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	icqkeeper "github.com/Stride-Labs/stride/v4/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	icacallbackskeeper "github.com/Stride-Labs/stride/v4/x/icacallbacks/keeper"
	recordsmodulekeeper "github.com/Stride-Labs/stride/v4/x/records/keeper"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                   codec.BinaryCodec
		storeKey              sdk.StoreKey
		memKey                sdk.StoreKey
		paramstore            paramtypes.Subspace
		ICAControllerKeeper   icacontrollerkeeper.Keeper
		IBCKeeper             ibckeeper.Keeper
		scopedKeeper          capabilitykeeper.ScopedKeeper
		bankKeeper            bankkeeper.Keeper
		InterchainQueryKeeper icqkeeper.Keeper
		RecordsKeeper         recordsmodulekeeper.Keeper
		StakingKeeper         stakingkeeper.Keeper
		ICACallbacksKeeper    icacallbackskeeper.Keeper
		hooks                 types.StakeIBCHooks
		accountKeeper         types.AccountKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	// channelKeeper cosmosibckeeper.ChannelKeeper,
	// portKeeper cosmosibckeeper.PortKeeper,
	// scopedKeeper cosmosibckeeper.ScopedKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	interchainQueryKeeper icqkeeper.Keeper,
	RecordsKeeper recordsmodulekeeper.Keeper,
	StakingKeeper stakingkeeper.Keeper,
	ICACallbacksKeeper icacallbackskeeper.Keeper,
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
		accountKeeper:         accountKeeper,
		bankKeeper:            bankKeeper,
		ICAControllerKeeper:   icacontrollerkeeper,
		IBCKeeper:             ibcKeeper,
		scopedKeeper:          scopedKeeper,
		InterchainQueryKeeper: interchainQueryKeeper,
		RecordsKeeper:         RecordsKeeper,
		StakingKeeper:         StakingKeeper,
		ICACallbacksKeeper:    ICACallbacksKeeper,
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

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
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
		return sdk.ZeroDec(), sdkerrors.Wrapf(sdkerrors.ErrNotFound, errMsg)
	}

	// Get epoch start time, end time, and duration
	epochDuration, err := cast.ToInt64E(epochTracker.Duration)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert epoch duration to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}
	epochEndTime, err := cast.ToInt64E(epochTracker.NextEpochStartTime)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert next epoch start time to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}
	epochStartTime := epochEndTime - epochDuration

	// Confirm the current block time is inside the current epoch's start and end times
	currBlockTime := ctx.BlockTime().UnixNano()
	if currBlockTime < epochStartTime || currBlockTime > epochEndTime {
		errMsg := fmt.Sprintf("current block time %d is not within current epoch (ending at %d)", currBlockTime, epochTracker.NextEpochStartTime)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrInvalidEpoch, errMsg)
	}

	// Get elapsed share
	elapsedTime := currBlockTime - epochStartTime
	elapsedShare := sdk.NewDec(elapsedTime).Quo(sdk.NewDec(epochDuration))
	if elapsedShare.LT(sdk.ZeroDec()) || elapsedShare.GT(sdk.OneDec()) {
		errMsg := fmt.Sprintf("elapsed share (%s) for epoch is not between 0 and 1", elapsedShare)
		k.Logger(ctx).Error(errMsg)
		return sdk.ZeroDec(), sdkerrors.Wrapf(types.ErrInvalidEpoch, errMsg)
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
		k.Logger(ctx).Error(fmt.Sprintf("ICQCB: We're %d pct through the epoch, ICQ cutoff is %d", elapsedShareOfEpoch, epochShareThresh))
	}
	return inWindow, nil
}

func (k Keeper) GetICATimeoutNanos(ctx sdk.Context, epochType string) (uint64, error) {
	epochTracker, found := k.GetEpochTracker(ctx, epochType)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to get epoch tracker for %s", epochType))
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to get epoch tracker for %s", epochType)
	}
	// BUFFER by 5% of the epoch length
	bufferSizeParam := k.GetParam(ctx, types.KeyBufferSize)
	bufferSize := epochTracker.Duration / bufferSizeParam
	// buffer size should not be negative or longer than the epoch duration
	if bufferSize > epochTracker.Duration {
		k.Logger(ctx).Error(fmt.Sprintf("Invalid buffer size %d", bufferSize))
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Invalid buffer size %d", bufferSize)
	}
	timeoutNanos := epochTracker.NextEpochStartTime - bufferSize
	timeoutNanosUint64, err := cast.ToUint64E(timeoutNanos)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to convert timeoutNanos to uint64, error: %s", err.Error()))
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to convert timeoutNanos to uint64, error: %s", err.Error())
	}
	return timeoutNanosUint64, nil
}

// safety check: ensure the redemption rate is NOT below our min safety threshold && NOT above our max safety threshold on host zone
func (k Keeper) IsRedemptionRateWithinSafetyBounds(ctx sdk.Context, zone types.HostZone) (bool, error) {
	minSafetyThresholdInt := k.GetParam(ctx, types.KeySafetyMinRedemptionRateThreshold)
	minSafetyThreshold := sdk.NewDec(int64(minSafetyThresholdInt)).Quo(sdk.NewDec(100))

	maxSafetyThresholdInt := k.GetParam(ctx, types.KeySafetyMaxRedemptionRateThreshold)
	maxSafetyThreshold := sdk.NewDec(int64(maxSafetyThresholdInt)).Quo(sdk.NewDec(100))

	redemptionRate := zone.RedemptionRate

	if redemptionRate.LT(minSafetyThreshold) || redemptionRate.GT(maxSafetyThreshold) {
		errMsg := fmt.Sprintf("IsRedemptionRateWithinSafetyBounds check failed %s is outside safety bounds [%s, %s]", redemptionRate.String(), minSafetyThreshold.String(), maxSafetyThreshold.String())
		k.Logger(ctx).Error(errMsg)
		return false, sdkerrors.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, errMsg)
	}
	return true, nil
}

// Check the max number of validators to confirm we won't exceed it when adding a new validator
// Types of additions:
//   - change a weight from zero to non-zero
//   - add a new validator with non-zero weight
func (k Keeper) ConfirmValSetHasSpace(ctx sdk.Context, validators []*types.Validator) error {

	// get max val parameter
	maxNumVals, err := cast.ToIntE(k.GetParam(ctx, types.KeySafetyNumValidators))
	if err != nil {
		errMsg := fmt.Sprintf("Error getting safety max num validators | err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrap(types.ErrMaxNumValidators, errMsg)
	}

	// count up the number of validators with non-zero weights
	numNonzeroWgtValidators := 0
	for _, validator := range validators {
		if validator.Weight > 0 {
			numNonzeroWgtValidators++
		}
	}

	// check if the number of validators with non-zero weights is below than the max
	if numNonzeroWgtValidators >= maxNumVals {
		errMsg := fmt.Sprintf("Attempting to add new validator but already reached max number of validators (%d)", maxNumVals)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrap(types.ErrMaxNumValidators, errMsg)
	}

	return nil
}
