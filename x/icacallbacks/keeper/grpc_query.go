package keeper

import (
	"github.com/Stride-Labs/stride/v5/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
