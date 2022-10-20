package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cast"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	epochtypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// ___________________________________________________________________________________________________

// Callbacks wrapper struct for interchainstaking keeper
type Callback func(Keeper, sdk.Context, []byte, icqtypes.Query) error

type Callbacks struct {
	k         Keeper
	callbacks map[string]Callback
}

var _ icqtypes.QueryCallbacks = Callbacks{}

func (k Keeper) CallbackHandler() Callbacks {
	return Callbacks{k, make(map[string]Callback)}
}

// callback handler
func (c Callbacks) Call(ctx sdk.Context, id string, args []byte, query icqtypes.Query) error {
	return c.callbacks[id](c.k, ctx, args, query)
}

func (c Callbacks) Has(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c Callbacks) AddCallback(id string, fn interface{}) icqtypes.QueryCallbacks {
	c.callbacks[id] = fn.(Callback)
	return c
}

func (c Callbacks) RegisterCallbacks() icqtypes.QueryCallbacks {
	return c.
		AddCallback("withdrawalbalance", Callback(WithdrawalBalanceCallback)).
		AddCallback("delegation", Callback(DelegatorSharesCallback)).
		AddCallback("validator", Callback(ValidatorExchangeRateCallback))
}

// -----------------------------------
// Callback Handlers
// -----------------------------------

// WithdrawalBalanceCallback is a callback handler for WithdrawalBalance queries.
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback executing, QueryId: %vs, Host: %s, QueryType: %s, Height: %d, Connection: %s",
		query.Id, query.ChainId, query.QueryType, query.Height, query.ConnectionId))

	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Unmarshal the CB args into a coin type
	withdrawalBalanceCoin := sdk.Coin{}
	err := k.cdc.Unmarshal(args, &withdrawalBalanceCoin)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal balance in callback args for zone: %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}

	// Check if the coin is nil (which would indicate the account never had a balance)
	if withdrawalBalanceCoin.IsNil() || withdrawalBalanceCoin.Amount.IsNil() {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: balance query returned a nil coin for address %s on %s, meaning the account has never had a balance on the host",
			hostZone.WithdrawalAccount.Address, hostZone.ChainId))
		return nil
	}

	// Confirm the balance is greater than zero
	if withdrawalBalanceCoin.Amount.Int64() <= 0 {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: no balance to transfer for zone: %s, accAddr: %v, coin: %v",
			hostZone.ChainId, hostZone.WithdrawalAccount.Address, withdrawalBalanceCoin.String()))
		return nil
	}

	// Sweep the withdrawal account balance, to the commission and the delegation accounts
	k.Logger(ctx).Info(fmt.Sprintf("ICA Bank Sending %d%s from withdrawalAddr to delegationAddr.",
		withdrawalBalanceCoin.Amount.Int64(), withdrawalBalanceCoin.Denom))

	withdrawalAccount := hostZone.GetWithdrawalAccount()
	if withdrawalAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no withdrawal account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}
	delegationAccount := hostZone.GetDelegationAccount()
	if delegationAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no delegation account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}
	feeAccount := hostZone.GetFeeAccount()
	if feeAccount == nil {
		errMsg := fmt.Sprintf("WithdrawalBalanceCallback: no fee account found for zone: %s", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICAAccountNotFound, errMsg)
	}

	params := k.GetParams(ctx)
	strideCommissionInt, err := cast.ToInt64E(params.GetStrideCommission())
	if err != nil {
		return err
	}

	// check that stride commission is between 0 and 1
	strideCommission := sdk.NewDec(strideCommissionInt).Quo(sdk.NewDec(100))
	if strideCommission.LT(sdk.ZeroDec()) || strideCommission.GT(sdk.OneDec()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Aborting reinvestment callback -- Stride commission must be between 0 and 1!")
	}

	withdrawalBalanceAmount := withdrawalBalanceCoin.Amount
	strideClaim := strideCommission.Mul(withdrawalBalanceAmount.ToDec())
	strideClaimFloored := strideClaim.TruncateInt()

	// back the reinvestment amount out of the total less the commission
	reinvestAmountCeil := sdk.NewInt(withdrawalBalanceAmount.Int64()).Sub(strideClaimFloored)

	// TODO(TEST-112) safety check, balances should add to original amount
	if (strideClaimFloored.Int64() + reinvestAmountCeil.Int64()) != withdrawalBalanceAmount.Int64() {
		ctx.Logger().Error(fmt.Sprintf("Error with withdraw logic: %d, Fee portion: %d, reinvestPortion %d", withdrawalBalanceAmount.Int64(), strideClaimFloored.Int64(), reinvestAmountCeil.Int64()))
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Failed to subdivide rewards to feeAccount and delegationAccount")
	}
	strideCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, strideClaimFloored)
	reinvestCoin := sdk.NewCoin(withdrawalBalanceCoin.Denom, reinvestAmountCeil)

	var msgs []sdk.Msg
	if strideCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.GetAddress(),
			ToAddress:   feeAccount.GetAddress(),
			Amount:      sdk.NewCoins(strideCoin),
		})
	}
	if reinvestCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{
			FromAddress: withdrawalAccount.GetAddress(),
			ToAddress:   delegationAccount.GetAddress(),
			Amount:      sdk.NewCoins(reinvestCoin),
		})
	}
	ctx.Logger().Info(fmt.Sprintf("Submitting withdrawal sweep messages for: %v", msgs))

	// add callback data
	reinvestCallback := types.ReinvestCallback{
		ReinvestAmount: reinvestCoin,
		HostZoneId:     hostZone.ChainId,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling ReinvestCallback args: %v", reinvestCallback))
	marshalledCallbackArgs, err := k.MarshalReinvestCallbackArgs(ctx, reinvestCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, hostZone.ConnectionId, msgs, *withdrawalAccount, REINVEST, marshalledCallbackArgs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to SubmitTxs for %s - %s, Messages: %v | err: %s", hostZone.ChainId, hostZone.ConnectionId, msgs, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICATxFailed, errMsg)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", withdrawalBalanceCoin.Amount.String()),
		),
	)

	return nil
}

// ValidatorCallback is a callback handler for validator queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal exchange rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     num_tokens = exchange_rate * num_shares
//
// This callback from query #1
func ValidatorExchangeRateCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal queriedValidator info for zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: HostZone %s, Queried Validator %v, Jailed: %v, Tokens: %v, Shares: %v",
		hostZone.ChainId, queriedValidator.OperatorAddress, queriedValidator.Jailed, queriedValidator.Tokens, queriedValidator.DelegatorShares))

	// ensure ICQ can be issued now! else fail the callback
	withinBufferWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, errMsg)
	} else if !withinBufferWindow {
		k.Logger(ctx).Error("validator exchange rate callback is outside ICQ window")
		return nil
	}

	// get the validator from the host zone
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedValidator.OperatorAddress)
	if !found {
		errMsg := fmt.Sprintf("no registered validator for address (%s)", queriedValidator.OperatorAddress)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrValidatorNotFound, errMsg)
	}
	// get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}

	// If the validator's delegation shares is 0, we'll get a division by zero error when trying to get the exchange rate
	//  because `validator.TokensFromShares` uses delegation shares in the denominator
	if queriedValidator.DelegatorShares.IsZero() {
		errMsg := fmt.Sprintf("can't calculate validator internal exchange rate because delegation amount is 0 (validator: %s)", validator.Address)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrDivisionByZero, errMsg)
	}

	// We want the validator's internal exchange rate which is held internally behind `validator.TokensFromShares`
	//  Since,
	//     exchange_rate = num_tokens / num_shares
	//  We can use `validator.TokensFromShares`, plug in 1.0 for the number of shares,
	//    and the returned number of tokens will be equal to the internal exchange rate
	validator.InternalExchangeRate = &types.ValidatorExchangeRate{
		InternalTokensToSharesRate: queriedValidator.TokensFromShares(sdk.NewDec(1.0)),
		EpochNumber:                strideEpochTracker.GetEpochNumber(),
	}
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: HostZone %s, Validator %v, tokensFromShares %v",
		hostZone.ChainId, validator.Address, validator.InternalExchangeRate.InternalTokensToSharesRate))

	// armed with the exch rate, we can now query the (val,del) delegation
	err = k.QueryDelegationsIcq(ctx, hostZone, queriedValidator.OperatorAddress)
	if err != nil {
		errMsg := fmt.Sprintf("ValidatorCallback: failed to query delegation, zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrICQFailed, errMsg)
	}
	return nil
}

// DelegationCallback is a callback handler for UpdateValidatorSharesExchRate queries.
//
// In an attempt to get the ICA's delegation amount on a given validator, we have to query:
//  1. the validator's internal exchange rate
//  2. the Delegation ICA's delegated shares
//     And apply the following equation:
//     num_tokens = exchange_rate * num_shares
//
// This callback from query #2
//
// NOTE(TEST-112) for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func DelegatorSharesCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	hostZone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		errMsg := fmt.Sprintf("no registered zone for queried chain ID (%s)", query.GetChainId())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Unmarshal the query response which returns a delegation object for the delegator/validator pair
	queriedDelgation := stakingtypes.Delegation{}
	err := k.cdc.Unmarshal(args, &queriedDelgation)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal queried delegation info for zone %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}
	k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: HostZone: %s, Delegator: %s, Validator: %s, Shares: %v",
		hostZone.ChainId, queriedDelgation.DelegatorAddress, queriedDelgation.ValidatorAddress, queriedDelgation.Shares))

	// ensure ICQ can be issued now! else fail the callback
	isWithinWindow, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("unable to determine if ICQ callback is inside buffer window, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrOutsideIcqWindow, errMsg)
	} else if !isWithinWindow {
		k.Logger(ctx).Error("delegator shares callback is outside ICQ window")
		return nil
	}

	// Grab the validator object form the hostZone using the address returned from the query
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, queriedDelgation.ValidatorAddress)
	if !found {
		errMsg := fmt.Sprintf("no registered validator for address (%s)", queriedDelgation.ValidatorAddress)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrValidatorNotFound, errMsg)
	}

	// get the validator's internal exchange rate, aborting if it hasn't been updated this epoch
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	if validator.InternalExchangeRate.EpochNumber != strideEpochTracker.GetEpochNumber() {
		errMsg := fmt.Sprintf("DelegationCallback: validator (%s) internal exchange rate has not been updated this epoch (epoch #%d)",
			validator.Address, strideEpochTracker.GetEpochNumber())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	// TODO: make sure conversion math precision matches the sdk's staking module's version (we did it slightly differently)
	// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
	validatorTokens := queriedDelgation.Shares.Mul(validator.InternalExchangeRate.InternalTokensToSharesRate).TruncateInt()
	k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: HostZone: %s, Validator: %s, Previous NumTokens: %d, Current NumTokens: %v",
		hostZone.ChainId, validator.Address, validator.DelegationAmt, validatorTokens))

	// Confirm the validator has actually been slashed
	if validatorTokens.Uint64() == validator.DelegationAmt {
		k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: Validator (%s) was not slashed", validator.Address))
		return nil
	} else if validatorTokens.Uint64() > validator.DelegationAmt {
		errMsg := fmt.Sprintf("DelegationCallback: Validator (%s) tokens returned from query is greater than the DelegationAmt", validator.Address)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, errMsg)
	}

	// TODO(TESTS-171) add some safety checks here (e.g. we could query the slashing module to confirm the decr in tokens was due to slash)
	// update our records of the total stakedbal and of the validator's delegation amt
	// NOTE:  we assume any decrease in delegation amt that's not tracked via records is a slash

	// Get slash percentage
	delegationAmount, err := cast.ToInt64E(validator.DelegationAmt)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert validator delegation amount to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}
	slashAmountUInt := validator.DelegationAmt - validatorTokens.Uint64()
	slashAmount, err := cast.ToInt64E(slashAmountUInt)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert validator slash amount to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}
	weight, err := cast.ToInt64E(validator.Weight)
	if err != nil {
		errMsg := fmt.Sprintf("unable to convert validator weight to int64, err: %s", err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
	}

	slashPct := sdk.NewDec(slashAmount).Quo(sdk.NewDec(delegationAmount))
	k.Logger(ctx).Info(fmt.Sprintf("ICQ'd Delegation Amoount Mismatch, HostZone: %s, Validator: %s, Delegator: %s, Records Tokens: %d, Tokens from ICQ %v, Slash Amount: %d, Slash Pct: %v!",
		hostZone.ChainId, validator.Address, queriedDelgation.DelegatorAddress, validator.DelegationAmt, validatorTokens, slashAmount, slashPct))

	// Abort if the slash was greater than 10%
	tenPercent := sdk.NewDec(10).Quo(sdk.NewDec(100))
	if slashPct.GT(tenPercent) {
		errMsg := fmt.Sprintf("DelegationCallback: Validator (%s) slashed but ABORTING update, slash is greater than 0.10 (%d)", validator.Address, slashPct)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrSlashGtTenPct, errMsg)
	}

	// Update the host zone and validator to reflect the weight and delegation change
	weightAdjustment := sdk.NewDec(validatorTokens.Int64()).Quo(sdk.NewDec(delegationAmount))
	validator.Weight = sdk.NewDec(weight).Mul(weightAdjustment).TruncateInt().Uint64()
	validator.DelegationAmt -= slashAmountUInt

	hostZone.StakedBal -= slashAmountUInt
	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	k.Logger(ctx).Info(fmt.Sprintf("Validator (%s) slashed! Delegation updated to: %v", validator.Address, validator.DelegationAmt))

	return nil
}
