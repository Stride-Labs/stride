package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/Stride-Labs/stride/x/icacallbacks/types"
    "github.com/Stride-Labs/stride/x/icacallbacks/keeper"
    keepertest "github.com/Stride-Labs/stride/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.IcacallbacksKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
