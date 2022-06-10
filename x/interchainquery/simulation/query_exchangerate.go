package simulation

import (
	"math/rand"

	"github.com/Stride-Labs/stride/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgQueryExchangerate(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgQueryExchangerate{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the QueryExchangerate simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "QueryExchangerate simulation not implemented"), nil, nil
	}
}
