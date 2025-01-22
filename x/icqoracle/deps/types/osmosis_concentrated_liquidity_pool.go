package types

import (
	fmt "fmt"

	"cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/deps/osmomath"
)

// SpotPrice returns the spot price of the pool.
// If base asset is the Token0 of the pool, we use the current sqrt price of the pool.
// If not, we calculate the inverse of the current sqrt price of the pool.
//
// Forked from https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/concentrated-liquidity/model/pool.go#L108-L129 under the Apache v2.0 License.
// Modified to return math.LegacyDec instead of osmomath.BigDec.
// Removed unused ctx param.
func (p OsmosisConcentratedLiquidityPool) SpotPrice(quoteAssetDenom string, baseAssetDenom string) (math.LegacyDec, error) {
	// validate base asset is in pool
	if baseAssetDenom != p.Token0 && baseAssetDenom != p.Token1 {
		return math.LegacyDec{}, fmt.Errorf("base asset denom (%s) is not in pool with (%s, %s) pair", baseAssetDenom, p.Token0, p.Token1)
	}
	// validate quote asset is in pool
	if quoteAssetDenom != p.Token0 && quoteAssetDenom != p.Token1 {
		return math.LegacyDec{}, fmt.Errorf("quote asset denom (%s) is not in pool with (%s, %s) pair", quoteAssetDenom, p.Token0, p.Token1)
	}
	if p.CurrentSqrtPrice.IsZero() {
		return math.LegacyDec{}, fmt.Errorf("zero sqrt price would result in either a zero spot price or division by zero when calculating the inverse price")
	}

	priceSquared := p.CurrentSqrtPrice.PowerInteger(2)
	// The reason why we convert the result to Dec and then back to BigDec is to temporarily
	// maintain backwards compatibility with the original implementation.
	// TODO: remove before https://github.com/osmosis-labs/osmosis/issues/5726 is complete
	if baseAssetDenom == p.Token0 {
		return priceSquared.Dec(), nil
	}
	return osmomath.OneBigDec().QuoMut(priceSquared).Dec(), nil
}
