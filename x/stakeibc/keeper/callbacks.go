package keeper

import (
	"bytes"
	"fmt"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ___________________________________________________________________________________________________

// Callbacks wrapper struct for interchainstaking keeper
type Callback func(Keeper, sdk.Context, []byte, types.Query) error

type Callbacks struct {
	k         Keeper
	callbacks map[string]Callback
}

var _ types.QueryCallbacks = Callbacks{}

func (k Keeper) CallbackHandler() Callbacks {
	return Callbacks{k, make(map[string]Callback)}
}

//callback handler
func (c Callbacks) Call(ctx sdk.Context, id string, args []byte, query types.Query) error {
	return c.callbacks[id](c.k, ctx, args, query)
}

func (c Callbacks) Has(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c Callbacks) AddCallback(id string, fn interface{}) types.QueryCallbacks {
	c.callbacks[id] = fn.(Callback)
	return c
}

func (c Callbacks) RegisterCallbacks() types.QueryCallbacks {
	a := c.
		// AddCallback("valset", Callback(ValsetCallback)).
		// AddCallback("validator", Callback(ValidatorCallback)).
		// AddCallback("delegations", Callback(DelegationsCallback)).
		// AddCallback("delegation", Callback(DelegationCallback)).
		// AddCallback("distributerewards", Callback(DistributeRewardsFromWithdrawAccount)).
		// AddCallback("depositinterval", Callback(DepositIntervalCallback)).
		// AddCallback("perfbalance", Callback(PerfBalanceCallback)).
		// AddCallback("allbalances", Callback(AllBalancesCallback)).
		// AddCallback("accountbalance", Callback(AccountBalanceCallback))
		AddCallback("withdrawalbalance", Callback(WithdrawalBalanceCallback))

	return a.(Callbacks)
}

// -----------------------------------
// Callback Handlers
// -----------------------------------

// WithdrawalBalanceCallback is a callback handler for WithdrawalBalance queries.
func WithdrawalBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	zone, found := k.GetHostZone(ctx, query.GetChainId())
	if !found {
		return fmt.Errorf("no registered zone for chain id: %s", query.GetChainId())
	}
	balancesStore := []byte(query.Request[1:])
	accAddr, err := banktypes.AddressFromBalancesStore(balancesStore)
	if err != nil {
		return err
	}

	coin := sdk.Coin{}
	err = k.cdc.Unmarshal(args, &coin)
	if err != nil {
		k.Logger(ctx).Error("unable to unmarshal balance info for zone", "zone", zone.ChainId, "err", err)
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

	// Set withdrawal balance as attribute on HostZone's withdrawal ICA account
	wa := zone.WithdrawalAccount
	wa.WithdrawalBalance = coin.Amount.Int64()
	zone.WithdrawalAccount = wa
	k.SetHostZone(ctx, zone)
	k.Logger(ctx).Info(fmt.Sprintf("Just set WithdrawalBalance to: %d", wa.WithdrawalBalance))
	// k.Logger(ctx).Info(fmt.Sprintf("Just set HeightLastQueriedUndelegatedBalance to: %d", h))
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", zone.ChainId),
			sdk.NewAttribute("totalWithdrawalBalance", coin.Amount.String()),
		),
	})

	// Sweep the withdrawal account balance, to the feeAccount and the delegationAccount
	//    - TODO(TEST-123) In the ICA bank send ack, create a deposit record for the swept amount that's in delegationAccount
	k.Logger(ctx).Info(fmt.Sprintf("ICA Bank Sending %d%s from withdrawalAddr to delegationAddr.", coin.Amount.Int64(), coin.Denom))

	withdrawalAccount := zone.GetWithdrawalAccount()
	delegationAccount := zone.GetDelegationAccount()
	feeAccount := zone.GetFeeAccount()

	params := k.GetParams(ctx)
	strideCommission := sdk.NewDec(int64(params.GetStrideCommission())).Quo(sdk.NewDec(100)) // convert to decimal
	withdrawalBalance := sdk.NewDec(coin.Amount.Int64())
	// TODO(TEST-112) don't perform unsafe uint64 to int64 conversion
	strideClaim := strideCommission.Mul(withdrawalBalance)
	strideClaimFloored := strideClaim.TruncateInt()

	reinvestAmount := (sdk.NewDec(1).Sub(strideCommission)).Mul(withdrawalBalance)
	reinvestAmountCeil := reinvestAmount.Ceil().TruncateInt()

	// TODO(TEST-112) safety check, balances should add to original amount
	if (strideClaimFloored.Int64() + reinvestAmountCeil.Int64()) != coin.Amount.Int64() {
		ctx.Logger().Error(fmt.Sprintf("Error with withdraw logic: %d, Fee portion: %d, reinvestPortion %d", coin.Amount.Int64(), strideClaimFloored.Int64(), reinvestAmountCeil.Int64()))
		panic("Failed to subdivide rewards to feeAccount and delegationAccount")
	}
	ctx.Logger().Info(fmt.Sprintf("MOOSE Total: %d, Fee portion: %d, reinvestPortion %d", coin.Amount.Int64(), strideClaimFloored.Int64(), reinvestAmountCeil.Int64()))
	strideCoin := sdk.NewCoin(coin.Denom, strideClaimFloored)
	reinvestCoin := sdk.NewCoin(coin.Denom, reinvestAmountCeil)

	var msgs []sdk.Msg
	// construct the msg
	msgs = append(msgs, &banktypes.MsgSend{FromAddress: withdrawalAccount.GetAddress(),
		ToAddress: feeAccount.GetAddress(), Amount: sdk.NewCoins(strideCoin)})
	msgs = append(msgs, &banktypes.MsgSend{FromAddress: withdrawalAccount.GetAddress(),
		ToAddress: delegationAccount.GetAddress(), Amount: sdk.NewCoins(reinvestCoin)})
	ctx.Logger().Info(fmt.Sprintf("MOOSE msg: %v", &banktypes.MsgSend{FromAddress: withdrawalAccount.GetAddress(),
		ToAddress: feeAccount.GetAddress(), Amount: sdk.NewCoins(coin)}))
	ctx.Logger().Info(fmt.Sprintf("MOOSE msgs: %v", msgs))
	// Send the transaction through SubmitTx
	err = k.SubmitTxs(ctx, zone.ConnectionId, msgs, *withdrawalAccount)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", zone.ConnectionId, zone.ChainId, msgs)
	}

	return nil
}
