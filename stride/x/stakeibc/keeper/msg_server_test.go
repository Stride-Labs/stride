package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/x/stakeibc/keeper"
	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.StakeibcKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
