package keeper

import (
	"bytes"
	"fmt"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	return nil
}
