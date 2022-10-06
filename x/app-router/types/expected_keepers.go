package types

import (
	"context"

	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"
)

// TransferKeeper defines the expected transfer keeper
type StakeibcKeeper interface {
	LiquidStake(goCtx context.Context, msg *stakeibctypes.MsgLiquidStake) (*stakeibctypes.MsgLiquidStakeResponse, error)
}
