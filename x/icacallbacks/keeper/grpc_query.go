package keeper

import (
	"github.com/Stride-Labs/stride/v3/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
