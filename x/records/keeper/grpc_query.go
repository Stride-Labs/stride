package keeper

import (
	"github.com/Stride-labs/stride/v4/x/records/types"
)

var _ types.QueryServer = Keeper{}
