package types

// DONTCOVER

import (
	"fmt"
)

var (
	ErrTotalWeightNotSet        = fmt.Errorf("total weight not set")
	ErrTotalWeightParse         = fmt.Errorf("total weight parse error")
	ErrFailedToGetTotalWeight   = fmt.Errorf("failed to get total weight")
	ErrFailedToParseDec         = fmt.Errorf("failed to parse dec from str")
	ErrAirdropAlreadyExists     = fmt.Errorf("airdrop with same identifier already exists")
	ErrDistributorAlreadyExists = fmt.Errorf("airdrop with same distributor already exists")
	ErrInvalidAmount            = fmt.Errorf("cannot claim negative tokens")
)
