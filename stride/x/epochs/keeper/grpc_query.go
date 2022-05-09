package keeper

import (
	"github.com/Stride-labs/stride/x/epochs/types"
)

var _ types.QueryServer = Keeper{}
