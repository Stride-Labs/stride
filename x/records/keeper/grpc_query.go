package keeper

import (
	"github.com/Stride-Labs/stride/v3/x/records/types"
)

var _ types.QueryServer = Keeper{}
