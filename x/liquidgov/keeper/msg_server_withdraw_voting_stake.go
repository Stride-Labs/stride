package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v11/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v11/x/stakeibc/types"
)

func (k msgServer) WithdrawVotingStake(goCtx context.Context, msg *types.MsgWithdrawVotingStake) (*types.MsgWithdrawVotingStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Keeper.Logger(ctx).Info(fmt.Sprintf("Withdrawing %d staked tokens from escrow", msg.Amount))

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

	// Find deposits from this user and confirm they have enough for the withdraw
	votingDeposit, found := k.GetDeposit(ctx, msg.Creator, hostZone.ChainId)
	k.Keeper.Logger(ctx).Info(fmt.Sprintf("voting deposit %+v found for creator %+v on chain %s", votingDeposit, voterAddress, hostZone.ChainId))		
	if !found || votingDeposit.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "no deposit of denom (%s) for this user", hostDenom)
	}

	// Confirm this user address has enough of token not locked into active votes that amount can be sent
	availableAmount := k.DepositAvailableNow(ctx, msg.Creator, hostZone.ChainId)
	if msg.Amount.GT(availableAmount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough tokens available to withdraw yet (%d)", availableAmount)		
	}

	// Confirm the module account has a sufficient balance of the stToken to send back the requested amount


	withdrawCoin := sdk.NewCoin(msg.StDenom, msg.Amount)

	// send back the amount of stTokens from the module account to the original owner
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, voterAddress, sdk.NewCoins(withdrawCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account for escrow withdraw", withdrawCoin.String())
	}

	// update the voting deposit amount... if zero delete the deposit structure
	if votingDeposit.Amount.GT(msg.Amount) {
		votingDeposit.Amount = votingDeposit.Amount.Sub(msg.Amount)
		k.Keeper.SetDeposit(ctx, votingDeposit)
	} else if votingDeposit.Amount.Equal(msg.Amount) {
		k.Keeper.DeleteDeposit(ctx, msg.Creator, hostZone.ChainId)
	}

	return &types.MsgWithdrawVotingStakeResponse{}, nil
}
