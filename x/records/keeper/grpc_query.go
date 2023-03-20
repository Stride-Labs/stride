package keeper

import (
	"github.com/Stride-Labs/stride/v7/x/records/types"
)

var _ types.QueryServer = Keeper{}
