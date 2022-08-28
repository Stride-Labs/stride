package keeper

import (
	"bytes"
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
func WithdrawalBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	// NOTE(TEST-112) for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
	k.Logger(ctx).Error(fmt.Sprintf("WithdrawalBalanceCallback: %v", query))

	zone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
	}
	balancesStore := query.Request[1:]
	accAddr, err := banktypes.AddressFromBalancesStore(balancesStore)
	if err != nil {
		return err
	}

	//TODO(TEST-112) revisit this code, it's not vetted
	coin := sdk.Coin{}
	err = k.cdc.Unmarshal(args, &coin)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("unable to unmarshal balance info for zone: %s, err: %s", zone.ChainId, err.Error()))
		return err
	}

	if coin.IsNil() {
		denom := ""

		for i := 0; i < len(query.Request)-len(accAddr); i++ {
			if bytes.Equal(query.Request[i:i+len(accAddr)], accAddr) {
				denom = string(query.Request[i+len(accAddr):])
				break
			}

		}
		// if balance is nil, the response sent back is nil, so we don't receive the denom. Override that now.
		coin = sdk.NewCoin(denom, sdk.ZeroInt())
	}

	// sanity check, do not transfer if we have 0 balance!
	if coin.Amount.Int64() <= 0 {
		k.Logger(ctx).Info(fmt.Sprintf("WithdrawalBalanceCallback: no balance to transfer for zone: %s, accAddr: %v, coin: %v", zone.ChainId, accAddr.String(), coin))
		return nil
	}

	// Set withdrawal balance as attribute on HostZone's withdrawal ICA account
	wa := zone.GetWithdrawalAccount()
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("host zone (%s) does not have a withdrawal account, err %s", zone.ChainId, err.Error()))
		return err
	}
	zone.WithdrawalAccount = wa
	k.SetHostZone(ctx, zone)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", zone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", coin.Amount.String()),
		),
	)

	// Sweep the withdrawal account balance, to the commission and the delegation accounts
	k.Logger(ctx).Info(fmt.Sprintf("ICA Bank Sending %d%s from withdrawalAddr to delegationAddr.", coin.Amount.Int64(), coin.Denom))

	withdrawalAccount := zone.GetWithdrawalAccount()
	if withdrawalAccount == nil {
		k.Logger(ctx).Error(fmt.Sprintf("WithdrawalBalanceCallback: no withdrawal account found for zone: %s", zone.ChainId))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, fmt.Sprintf("WithdrawalBalanceCallback: no withdrawal account found for zone: %s", zone.ChainId))
	}
	delegationAccount := zone.GetDelegationAccount()
	if delegationAccount == nil {
		k.Logger(ctx).Error(fmt.Sprintf("WithdrawalBalanceCallback: no delegation account found for zone: %s", zone.ChainId))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, fmt.Sprintf("WithdrawalBalanceCallback: no delegation account found for zone: %s", zone.ChainId))
	}
	feeAccount := zone.GetFeeAccount()
	if feeAccount == nil {
		k.Logger(ctx).Error(fmt.Sprintf("WithdrawalBalanceCallback: no fee account found for zone: %s", zone.ChainId))
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, fmt.Sprintf("WithdrawalBalanceCallback: no fee account found for zone: %s", zone.ChainId))
	}

	params := k.GetParams(ctx)
	stCommission, err := cast.ToInt64E(params.GetStrideCommission())
	if err != nil {
		return err
	}
	strideCommission := sdk.NewDec(stCommission).Quo(sdk.NewDec(100))
	// check that stride commission is between 0 and 1
	if strideCommission.LT(sdk.ZeroDec()) || strideCommission.GT(sdk.OneDec()) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Aborting reinvestment callback -- Stride commission must be between 0 and 1!")
	}
	withdrawalBalance := sdk.NewDec(coin.Amount.Int64())
	// TODO(TEST-112) don't perform unsafe uint64 to int64 conversion
	strideClaim := strideCommission.Mul(withdrawalBalance)
	strideClaimFloored := strideClaim.TruncateInt()

	// back the reinvestment amount out of the total less the commission
	reinvestAmountCeil := sdk.NewInt(coin.Amount.Int64()).Sub(strideClaimFloored)

	// TODO(TEST-112) safety check, balances should add to original amount
	if (strideClaimFloored.Int64() + reinvestAmountCeil.Int64()) != coin.Amount.Int64() {
		ctx.Logger().Error(fmt.Sprintf("Error with withdraw logic: %d, Fee portion: %d, reinvestPortion %d", coin.Amount.Int64(), strideClaimFloored.Int64(), reinvestAmountCeil.Int64()))
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Failed to subdivide rewards to feeAccount and delegationAccount")
	}
	strideCoin := sdk.NewCoin(coin.Denom, strideClaimFloored)
	reinvestCoin := sdk.NewCoin(coin.Denom, reinvestAmountCeil)

	var msgs []sdk.Msg
	// construct the msg
	if strideCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{FromAddress: withdrawalAccount.GetAddress(),
			ToAddress: feeAccount.GetAddress(), Amount: sdk.NewCoins(strideCoin)})
	}
	if reinvestCoin.Amount.Int64() > 0 {
		msgs = append(msgs, &banktypes.MsgSend{FromAddress: withdrawalAccount.GetAddress(),
			ToAddress: delegationAccount.GetAddress(), Amount: sdk.NewCoins(reinvestCoin)})
	}
	ctx.Logger().Info(fmt.Sprintf("Submitting withdrawal sweep messages for: %v", msgs))

	// add callback data
	reinvestCallback := types.ReinvestCallback{
		ReinvestAmount: reinvestCoin,
		HostZoneId:     zone.ChainId,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling ReinvestCallback args: %v", reinvestCallback))
	marshalledCallbackArgs, err := k.MarshalReinvestCallbackArgs(ctx, reinvestCallback)
	if err != nil {
		return err
	}

	// Send the transaction through SubmitTx
	_, err = k.SubmitTxsStrideEpoch(ctx, zone.ConnectionId, msgs, *withdrawalAccount, REINVEST, marshalledCallbackArgs)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", zone.ConnectionId, zone.ChainId, msgs)
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
