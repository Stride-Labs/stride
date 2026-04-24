package keeper

import (
	"github.com/Stride-Labs/stride/v32/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
