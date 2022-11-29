package keeper

import (
	"github.com/Stride-labs/stride/v4/x/stakeibc/types"
)

var _ types.QueryServer = Keeper{}
