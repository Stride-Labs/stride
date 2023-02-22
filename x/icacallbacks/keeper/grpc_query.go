package keeper

import (
	"github.com/Stride-Labs/stride/v6/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
