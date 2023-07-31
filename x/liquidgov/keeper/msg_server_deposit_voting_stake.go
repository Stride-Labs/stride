package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func (k msgServer) DepositVotingStake(goCtx context.Context, msg *types.MsgDepositVotingStake) (*types.MsgDepositVotingStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Keeper.Logger(ctx).Info(fmt.Sprintf("Depositing %d staked tokens in escrow for voting", msg.Amount))

	// Verify that the denom is in fact a legal stToken and get which Token and ultimately hostZone it represents
	isStToken := k.stakeibcKeeper.CheckIsStToken(ctx, msg.StDenom)
	k.Keeper.Logger(ctx).Info(fmt.Sprintf("stdenom was %s and st-check was %v", msg.StDenom, isStToken))
	if !isStToken {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken, "token must be a staked denom (%s)", msg.StDenom)
	}

	// validate that the stToken is legal and from a real hostZone
	hostDenom := types.HostZoneDenomFromStAssetDenom(msg.StDenom)
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, hostDenom)
	k.Keeper.Logger(ctx).Info(fmt.Sprintf("base denom %s host zone found %v", hostDenom, hostZone))	
	if err != nil {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken, "no host zone found for denom (%s)", hostDenom)
	}

	// Get user and module account addresses
	voterAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}

	// Confirm the user has a sufficient balance of the stToken to deposit the attempted amount
	depositCoin := sdk.NewCoin(msg.StDenom, msg.Amount)
	balance := k.bankKeeper.GetBalance(ctx, voterAddress, msg.StDenom)
	if balance.IsLT(depositCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}


	// escrow the amount of stTokens in the module account for the purpose of voting
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, voterAddress, types.ModuleName, sdk.NewCoins(depositCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account for voting escrow", depositCoin.String())
	}
	
	// create or update the deposit amount to track how much of the coin for chainId this user has
	votingDeposit, found := k.GetDeposit(ctx, msg.Creator, hostZone.ChainId)
	if found {
		votingDeposit.Amount = votingDeposit.Amount.Add(msg.Amount)
	} else {
		votingDeposit, _ = types.NewDeposit(msg.Creator, hostZone.ChainId, msg.Amount)
	}
	k.Keeper.SetDeposit(ctx, votingDeposit)
	k.Keeper.Logger(ctx).Info(fmt.Sprintf("setting the voting deposit %v", votingDeposit))

	return &types.MsgDepositVotingStakeResponse{}, nil
}
