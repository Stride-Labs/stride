package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride as an IBC voucher (whose denomtrace we recognize)
// before this function is called
//
// The typical flow:
//   - A staker tokenizes their delegation on the host zone
//   - The staker IBC transfers their tokenized shares to Stride
//   - They then call LSMLiquidStake
//   - - The staker's LSM Tokens are sent to the Stride module account
//   - - The staker recieves stTokens
//
// As a safety measure, at period checkpoints, the validator's sharesToTokens rate is queried and the transaction
// is not settled until the query returns
// As a result, this transaction has been split up into a (1) Start and (2) Finish function
//   - If no query is needed, (2) is called immediately after (1)
//   - If a query is needed, (2) is called in the query callback
//
// The transaction response indicates if the query occurred by returning an attribute `TransactionComplete` set to false
func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lsmLiquidStake, err := k.StartLSMLiquidStake(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if k.ShouldCheckIfValidatorWasSlashed(ctx, *lsmLiquidStake.Validator, msg.Amount) {
		if err := k.SubmitValidatorSlashQuery(ctx, lsmLiquidStake); err != nil {
			return nil, err
		}

		EmitPendingLSMLiquidStakeEvent(ctx, *lsmLiquidStake.HostZone, *lsmLiquidStake.Deposit)

		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	async := false
	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake, async); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}
