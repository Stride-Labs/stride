package keeper

import (
	"github.com/Stride-Labs/stride/x/records/types"
)

var _ types.QueryServer = Keeper{}
