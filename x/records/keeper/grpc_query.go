package keeper

import (
	"github.com/Stride-Labs/stride/v9/x/records/types"
)

var _ types.QueryServer = Keeper{}
