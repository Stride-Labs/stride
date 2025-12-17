package keeper

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/migrations/v3"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	epochstypes "github.com/Stride-Labs/stride/v31/x/epochs/types"

	"github.com/Stride-Labs/stride/v31/utils"
	icqtypes "github.com/Stride-Labs/stride/v31/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v31/x/stakeibc/types"
)

// Submit a validator sharesToTokens rate ICQ as triggered either manually or epochly with a conservative timeout
func (k Keeper) QueryValidatorSharesToTokensRate(ctx sdk.Context, chainId string, validatorAddress string) error {
	timeoutDuration := time.Hour * 24
	timeoutPolicy := icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE
	callbackData := []byte{}
	return k.SubmitValidatorSharesToTokensRateICQ(ctx, chainId, validatorAddress, callbackData, timeoutDuration, timeoutPolicy)
}

// Submits an ICQ to get a validator's shares to tokens rate
func (k Keeper) SubmitValidatorSharesToTokensRateICQ(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	callbackDataBz []byte,
	timeoutDuration time.Duration,
	timeoutPolicy icqtypes.TimeoutPolicy,
) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Submitting ICQ for validator sharesToTokens rate to %s", validatorAddress))

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidHostZone, "Host zone not found (%s)", chainId)
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(validatorAddress, hostZone.Bech32Prefix) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must match the host zone bech32 prefix")
	}

	// Encode the validator address to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator operator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetValidatorKey(validatorAddressBz)

	// Submit validator sharesToTokens rate ICQ
	// Considering this query is executed manually, we can be conservative with the timeout
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Validator,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   timeoutPolicy,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, true); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for validator sharesToTokens rate, error %s", err.Error()))
		return err
	}
	return nil
}

// Submits an ICQ to get a validator's delegations
// This is called after the validator's sharesToTokens rate is determined
// The timeoutDuration parameter represents the length of the timeout (not to be confused with an actual timestamp)
func (k Keeper) SubmitDelegationICQ(ctx sdk.Context, hostZone types.HostZone, validatorAddress string) error {
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", validatorAddress)
	}

	// Only submit the query if there's not already one in progress
	if validator.SlashQueryInProgress {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Delegations ICQ already in progress"))
		return nil
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for delegations to %s", validatorAddress))

	// Get the validator and delegator encoded addresses to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator address, could not decode (%s)", err.Error())
	}
	_, delegatorAddressBz, err := bech32.DecodeAndConvert(hostZone.DelegationIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid delegator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)

	// Store the current validator's delegation in the callback data so we can determine if it changed
	// while the query was in flight
	callbackData := types.DelegatorSharesQueryCallback{
		InitialValidatorDelegation: validator.Delegation,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal delegator shares callback data")
	}

	// Update the validator to indicate that the slash query is in progress
	validator.SlashQueryInProgress = true
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	// Submit delegator shares ICQ
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Delegation,
		CallbackData:    callbackDataBz,
		TimeoutDuration: time.Hour,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return err
	}

	return nil
}

// Submits an ICQ to get a validator's delegations
// This is called after the validator's sharesToTokens rate is determined
// The timeoutDuration parameter represents the length of the timeout (not to be confused with an actual timestamp)
func (k Keeper) SubmitCalibrationICQ(ctx sdk.Context, hostZone types.HostZone, validatorAddress string) error {
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// ensure the validator is in the set for this host
	_, _, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "no registered validator for address (%s)", validatorAddress)
	}

	// Get the validator and delegator encoded addresses to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator address, could not decode (%s)", err.Error())
	}
	_, delegatorAddressBz, err := bech32.DecodeAndConvert(hostZone.DelegationIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid delegator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetDelegationKey(delegatorAddressBz, validatorAddressBz)

	// Submit delegator shares ICQ
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Calibrate,
		CallbackData:    []byte{},
		TimeoutDuration: time.Hour,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return err
	}

	return nil
}

// Submits an ICQ for the withdrawal account balance
func (k Keeper) SubmitWithdrawalHostBalanceICQ(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Submitting ICQ for withdrawal account balance"))

	// Get the withdrawal account address from the host zone
	if hostZone.WithdrawalIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no withdrawal account found for %s", hostZone.ChainId)
	}

	// Encode the withdrawal account address for the query request
	// The query request consists of the withdrawal account address and denom
	_, withdrawalAddressBz, err := bech32.DecodeAndConvert(hostZone.WithdrawalIcaAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid withdrawal account address, could not decode (%s)", err.Error())
	}
	queryData := append(bankTypes.CreateAccountBalancesPrefix(withdrawalAddressBz), []byte(hostZone.HostDenom)...)

	// Timeout query at end of epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochstypes.STRIDE_EPOCH)
	if !found {
		return errorsmod.Wrapf(types.ErrEpochNotFound, "epoch %s not found", epochstypes.STRIDE_EPOCH)
	}
	timeout := time.Unix(0, utils.UintToInt(strideEpochTracker.NextEpochStartTime))
	timeoutDuration := timeout.Sub(ctx.BlockTime())

	// Submit the ICQ for the withdrawal account balance
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.BANK_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_WithdrawalHostBalance,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error querying for withdrawal balance, error: %s", err.Error()))
		return err
	}

	return nil
}
