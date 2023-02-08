package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/Stride-Labs/stride/v5/x/auction/types"
    "github.com/Stride-Labs/stride/v5/x/auction/keeper"
    keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.AuctionKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
