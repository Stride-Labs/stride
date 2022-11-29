package keeper

import (
	"github.com/Stride-labs/stride/v4/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
