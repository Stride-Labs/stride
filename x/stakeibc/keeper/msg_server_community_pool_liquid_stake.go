package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// A user can specify a hub with a community pool, a token denom, and an amount to instantly kick off a sweep in and liquid stake
// If the communityPoolDepositAddress ICA on the foreign hub does not contain the tokens specified it will just async fail silently
func (k msgServer) CommunityPoolLiquidStake(goCtx context.Context, msg *types.MsgCommunityPoolLiquidStake) (*types.MsgCommunityPoolLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	token := sdk.NewCoin(msg.TokenDenom, msg.Amount)
	err := k.IBCTransferCommunityPoolICATokensToStride(ctx, msg.ChainId, token)
	if err != nil {
		return nil, err
	}

	return &types.MsgCommunityPoolLiquidStakeResponse{}, nil
}
