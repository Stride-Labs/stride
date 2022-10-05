package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/types"
)

type msgServer struct {
	keeper Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) DepositAirdrop(goCtx context.Context, msg *types.MsgDepositAirdrop) (*types.MsgDepositAirdropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(msg.Distributor)
	if err != nil {
		return nil, err
	}

	err = server.keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, msg.AirdropAmount)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositAirdropResponse{}, nil
}
