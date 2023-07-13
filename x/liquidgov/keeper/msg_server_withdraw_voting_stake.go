package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) WithdrawVotingStake(goCtx context.Context, msg *types.MsgWithdrawVotingStake) (*types.MsgWithdrawVotingStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//k.Keeper.Logger(ctx).Info("Withdrawing %d staked tokens for voting", msg.Amount)

	// Verify that the denom is in fact a legal stToken and get which Token and ultimately hostZone it represents
	isStToken := k.stakeibcKeeper.CheckIsStToken(ctx, msg.Denom)
	if !isStToken {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken, "token must be a staked denom (%s)", msg.Denom)
	}

	// validate that the stToken is legal and from a real hostZone
	hostDenom := types.HostZoneDenomFromStAssetDenom(msg.Denom)
	hostZone, err := k.stakeibcKeeper.GetHostZoneFromHostDenom(ctx, hostDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidToken, "no host zone found for denom (%s)", hostDenom)
	}

	// Get user and module account addresses
	voterAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}

	// Find deposits from this user and confirm they have enough for the withdraw
	votingDeposit, found :=k.GetDeposit(ctx, sdk.AccAddress(msg.Creator), hostZone.ChainId)
	if !found || votingDeposit.Amount.LT(msg.Amount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "no deposit of denom (%s) for this user", hostDenom)
	}

	// Confirm this user address has enough of token not locked into active votes that amount can be sent
	availableAmount := k.DepositAvailableNow(ctx, msg.Creator, hostZone.ChainId)
	if msg.Amount.GT(availableAmount) {
		return nil, errorsmod.Wrapf(stakeibctypes.ErrInvalidAmount, "not enough tokens available to withdraw yet (%d)", availableAmount)		
	}

	// Confirm the module account has a sufficient balance of the stToken to send back the requested amount


	withdrawCoin := sdk.NewCoin(msg.Denom, msg.Amount)

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
