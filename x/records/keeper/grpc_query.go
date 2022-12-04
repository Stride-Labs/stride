package keeper

import (
	"github.com/Stride-Labs/stride/v4/x/records/types"
)

var _ types.QueryServer = Keeper{}
