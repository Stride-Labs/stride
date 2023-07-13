package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v10/x/liquidgov/types"
	stakeibctypes "github.com/Stride-Labs/stride/v10/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) DepositVotingStake(goCtx context.Context, msg *types.MsgDepositVotingStake) (*types.MsgDepositVotingStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	//k.Keeper.Logger(ctx).Info("Depositing %d staked tokens for voting", msg.Amount)

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

	// Confirm the user has a sufficient balance of the stToken to deposit the attempted amount
	depositCoin := sdk.NewCoin(msg.Denom, msg.Amount)
	balance := k.bankKeeper.GetBalance(ctx, voterAddress, msg.Denom)
	if balance.IsLT(depositCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}


	// escrow the amount of stTokens in the module account for the purpose of voting
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, voterAddress, types.ModuleName, sdk.NewCoins(depositCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account for voting escrow", depositCoin.String())
	}
	
	// create or update the deposit amount to track how much of the coin for chainId this user has
	votingDeposit, found :=k.GetDeposit(ctx, sdk.AccAddress(msg.Creator), hostZone.ChainId)
	if found {
		votingDeposit.Amount = votingDeposit.Amount.Add(msg.Amount)
	} else {
		votingDeposit, _ = types.NewDeposit(msg.Creator, hostZone.ChainId, msg.Amount)
	}
	k.Keeper.SetDeposit(ctx, votingDeposit)

	return &types.MsgDepositVotingStakeResponse{}, nil
}
