package keeper

import (
	"github.com/Stride-Labs/stride/v6/x/records/types"
)

var _ types.QueryServer = Keeper{}
