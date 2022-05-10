package keeper

import (
	"github.com/Stride-labs/stride/x/interchainquery/types"
)

var _ types.QueryServer = Keeper{}
