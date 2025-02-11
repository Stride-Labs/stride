package types

import (
	"cosmossdk.io/math"
)

// Pool defines the interface that all pool types must implement
type Pool interface {
	// SpotPrice returns the spot price of the pool given a quote and base asset denom
	// Returns math.LegacyDec for the spot price and error if the operation fails
	CalcSpotPrice(quoteAssetDenom string, baseAssetDenom string) (math.LegacyDec, error)
}
