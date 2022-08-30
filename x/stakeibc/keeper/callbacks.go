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

//callback handler
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

	// Query request is a byte array of the form:
	// [ {balancesPrefix} {address} {denom} ]
	// {balancePrefix} is only a single byte - and it must be removed before calling AddressFromBalancesStore
	balancesStoreKey := query.Request[1:]
	queriedAddress, err := banktypes.AddressFromBalancesStore(balancesStoreKey)
	if err != nil {
		errMsg := fmt.Sprintf("unable to derive queried address from request byte array")
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(err, errMsg)
	}

	//TODO(TEST-112) revisit this code, it's not vetted
	// Unmarshal the CB args into a coin type
	withdrawalBalanceCoin := sdk.Coin{}
	err = k.cdc.Unmarshal(args, &withdrawalBalanceCoin)
	if err != nil {
		errMsg := fmt.Sprintf("unable to unmarshal balance in callback args for zone: %s, err: %s", hostZone.ChainId, err.Error())
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrapf(types.ErrMarshalFailure, errMsg)
	}

	// SOMEONE DOUBLE CHECK ME ON THIS
	// It looks unmarshalling a nil coin amount converts it to zero, which would mean this removed branch should be impossible
	// See test TestWithdrawalBalanceCallback_ZeroBalanceImplied

	// sanity check, do not transfer if we have 0 balance!
	if withdrawalBalanceCoin.Amount.Int64() <= 0 {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: no balance to transfer for zone: %s, accAddr: %v, coin: %v",
			hostZone.ChainId, queriedAddress.String(), withdrawalBalanceCoin.String()))
		return nil
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", withdrawalBalanceCoin.Amount.String()),
		),
	)

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

	return nil
}

// get a validator and its index from a list of validators, by address
func getValidator(validators []*types.Validator, address string) (types.Validator, int64, bool) {
	for i, v := range validators {
		if v.Address == address {
			return *v, int64(i), true
		}
	}
	return types.Validator{}, 0, false
}

// ValidatorCallback is a callback handler for validator queries.
func ValidatorExchangeRateCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	zone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
	}
	queriedValidator := stakingtypes.Validator{}
	err := k.cdc.Unmarshal(args, &queriedValidator)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal queriedValidator info for zone %s, err: %s", zone.ChainId, err.Error()))
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: zone %v queriedValidator %v", zone.ChainId, queriedValidator))

	// ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return err
	} else if !valid {
		return nil
	}

	// set the validator's conversion rate
	v, i, found := getValidator(zone.Validators, queriedValidator.OperatorAddress)
	if !found {
		return fmt.Errorf("no registered validator for address: %s", queriedValidator.OperatorAddress)
	}
	// get the stride epoch number
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		k.Logger(ctx).Error("failed to find stride epoch")
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	// converting 1.0 gives us the exchange rate to later use in the next CB
	v.InternalExchangeRate = &types.ValidatorExchangeRate{
		InternalTokensToSharesRate: queriedValidator.TokensFromShares(sdk.NewDec(1.0)),
		EpochNumber:                strideEpochTracker.GetEpochNumber(),
	}
	k.Logger(ctx).Info(fmt.Sprintf("ValidatorCallback: zone %s validator %v tokensFromShares %v", zone.ChainId, v.Address, v.InternalExchangeRate.InternalTokensToSharesRate))
	// write back to state and break
	zone.Validators[i] = &v
	k.SetHostZone(ctx, zone)

	// armed with the exch rate, we can now query the (val,del) delegation
	err = k.QueryDelegationsIcq(ctx, zone, queriedValidator.OperatorAddress)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("ValidatorCallback: failed to query delegation, zone %s, err: %s", zone.ChainId, err.Error()))
		return err
	}
	return nil
}

// DelegationCallback is a callback handler for UpdateValidatorSharesExchRate queries.
func DelegatorSharesCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	// NOTE(TEST-112) for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"

	zone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
	}

	// ensure ICQ can be issued now! else fail the callback
	valid, err := k.IsWithinBufferWindow(ctx)
	if err != nil {
		return err
	} else if !valid {
		return nil
	}

	qdel := stakingtypes.Delegation{}
	err = k.cdc.Unmarshal(args, &qdel)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal qdel info for zone %s, err: %s", zone.ChainId, err.Error()))
		return err
	}
	k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: zone %s qdel %v", zone.ChainId, qdel))

	// get tokens using the validator's conversion rate
	for i, v := range zone.Validators {
		k.Logger(ctx).Info(fmt.Sprintf("DELCB %s", v.Address))
		if v.Address == qdel.ValidatorAddress {
			delAmtInt64, err := cast.ToInt64E(v.DelegationAmt)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("unable to convert delegationAmt to uint64, err: %s", err.Error()))
				return err
			}

			// convert shares to tokens using the exchange rate

			// get the validator's internal exchange rate, aborting if it hasn't been updated this epoch
			strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
			if !found {
				k.Logger(ctx).Error("failed to find stride epoch")
				return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
			}
			if v.InternalExchangeRate.EpochNumber != strideEpochTracker.GetEpochNumber() {
				k.Logger(ctx).Error(fmt.Sprintf("delegation callback: validator %s internalExchRate has not been updated this epoch", v.Address))
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "validator %s internalExchRate  has not been updated this epoch", v.Address)
			}
			// TODO: make sure conversion math precision matches the sdk's staking module's version (we did it slightly differently)
			// note: truncateInt per https://github.com/cosmos/cosmos-sdk/blob/cb31043d35bad90c4daa923bb109f38fd092feda/x/staking/types/validator.go#L431
			qNumTokens := qdel.Shares.Mul(v.InternalExchangeRate.InternalTokensToSharesRate).TruncateInt()
			k.Logger(ctx).Info(fmt.Sprintf("DelegationCallback: zone %s validator %s prevNtokens %v qNumTokens %v", zone.ChainId, v.Address, v.DelegationAmt, qNumTokens))
			if qNumTokens.Uint64() < v.DelegationAmt {
				// TODO(TESTS-171) add some safety checks here (e.g. we could query the slashing module to confirm the decr in tokens was due to slash)
				// update our records of the total stakedbal and of the validator's delegation amt
				// NOTE:  we assume any decrease in delegation amt that's not tracked via records is a slash
				slashAmt := v.DelegationAmt - qNumTokens.Uint64()
				slashAmtInt64, err := cast.ToInt64E(slashAmt)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("unable to convert slashAmt to uint64, err: %s", err.Error()))
					return err
				}
				weightInt64, err := cast.ToInt64E(v.Weight)
				if err != nil {
					k.Logger(ctx).Error(fmt.Sprintf("unable to convert weight to uint64, err: %s", err.Error()))
					return err
				}

				slashPct := sdk.NewDec(slashAmtInt64).Quo(sdk.NewDec(delAmtInt64))
				k.Logger(ctx).Info(fmt.Sprintf("ICQ'd delAmt mismatch zone %s validator %s delegator %s records was %d icq shows %d slashAmt %d slashPct %d... UPDATING!",
					zone.ChainId, v.Address, qdel.DelegatorAddress, v.DelegationAmt, qNumTokens, slashAmt, slashPct))
				// TODO(TEST-172): move rate limiting logic to new rate limiting module

				if slashPct.GT(sdk.NewDec(10).Quo(sdk.NewDec(100))) {
					k.Logger(ctx).Error(fmt.Sprintf("DELCB | slashed but ABORTING bc slash GT10pct: query shows slash of %v", slashPct))
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "DELCB | slashed but ABORTING bc slash GT10pct: query shows slash of %v", slashPct)
				}
				// slash the validator's weight
				weightMul := sdk.NewDec(qNumTokens.Int64()).Quo(sdk.NewDec(delAmtInt64))

				zone.StakedBal -= slashAmt
				v.DelegationAmt -= slashAmt
				v.Weight = sdk.NewDec(weightInt64).Mul(weightMul).TruncateInt().Uint64()

				// write back to state and break
				zone.Validators[i] = v
				k.Logger(ctx).Info(fmt.Sprintf("SLASHING! val to update to: %v", zone.Validators[i].String()))
				k.SetHostZone(ctx, zone)

				zone, found = k.GetHostZone(ctx, zone.ChainId)
				if !found {
					k.Logger(ctx).Error(fmt.Sprintf("failed to find zone %s", zone.ChainId))
					return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "no zone %s", zone.ChainId)
				}
				k.Logger(ctx).Info(fmt.Sprintf("SLASHED! val updated: %v", zone.Validators[i].String()))
				break
			}
		}
	}
	return nil
}
