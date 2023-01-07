package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v4/utils"
)

func SimulateMsgChangeValidatorWeight(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		
		simAccount, _ := simtypes.RandomAcc(r, accs)
		val2Account, _ := simtypes.RandomAcc(r, accs)
		creatorAccount, _ := simtypes.RandomAcc(r, accs)
		
		err := utils.ValidateAdminAddress(string(creatorAccount.Address))
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeValidatorWeight, "Address of random account is not creator's address" ), nil, nil
		}
		hostZone := types.HostZone{
			ChainId: "GAIA",
			Validators: []*types.Validator{
				{
					Name: simAccount.Address.String(),
					Address: simAccount.Address.String(),
					CommissionRate: 1,
					Weight: 0,
					Status: types.Validator_ACTIVE,
					DelegationAmt: sdk.ZeroInt(),
				},
				{
					Name: val2Account.Address.String(),
					Address: val2Account.Address.String(),
					CommissionRate: 1,
                    Weight: 0,
					Status: types.Validator_ACTIVE,
					DelegationAmt: sdk.ZeroInt(),
				},
			},
		}
		k.SetHostZone(ctx, hostZone)
		
		hostZone, found := k.GetHostZone(ctx, "GAIA")
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeValidatorWeight, "Hostzone with random validators not found" ), nil, nil
		}

		if len(hostZone.Validators) != 2 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgChangeValidatorWeight, "Wrong number of hostzones" ), nil, nil
		}

 		msg := &types.MsgChangeValidatorWeight{
			Creator: creatorAccount.Address.String(),
			HostZone: "GAIA",
			ValAddr: simAccount.Address.String(),
			Weight: 1,
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}
		
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
