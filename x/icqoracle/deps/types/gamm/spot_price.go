package gamm

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
)

// SpotPrice returns the spot price of the pool
// This is the weight-adjusted balance of the tokens in the pool.
// To reduce the propagated effect of incorrect trailing digits,
// we take the ratio of weights and divide this by ratio of supplies
// this is equivalent to spot_price = (Quote Supply / Quote Weight) / (Base Supply / Base Weight)
//
// As an example, assume equal weights. uosmo supply of 2 and uatom supply of 4.
//
// Case 1: base = uosmo, quote = uatom -> for one uosmo, get 2 uatom = 4 / 2 = 2
// In other words, it costs 2 uatom to get one uosmo.
//
// Case 2: base = uatom, quote = uosmo -> for one uatom, get 0.5 uosmo = 2 / 4 = 0.5
// In other words, it costs 0.5 uosmo to get one uatom.
//
// panics if the pool in state is incorrect, and has any weight that is 0.
//
// Forked from https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/gamm/pool-models/balancer/pool.go#L617-L649 under the Apache v2.0 License.
// Modified to return math.LegacyDec instead of osmomath.BigDec.
// Removed unused ctx param.
func (p OsmosisGammPool) SpotPrice(quoteAssetDenom string, baseAssetDenom string) (spotPrice math.LegacyDec, err error) {
	quote, base, err := p.parsePoolAssetsByDenoms(quoteAssetDenom, baseAssetDenom)
	if err != nil {
		return math.LegacyDec{}, err
	}
	if base.Weight.IsZero() || quote.Weight.IsZero() {
		return math.LegacyDec{}, errors.New("pool is misconfigured, got 0 weight")
	}

	// spot_price = (Quote Supply / Quote Weight) / (Base Supply / Base Weight)
	//            = (Quote Supply / Quote Weight) * (Base Weight / Base Supply)
	//            = (Base Weight  / Quote Weight) * (Quote Supply / Base Supply)
	invWeightRatio := base.Weight.ToLegacyDec().Quo(quote.Weight.ToLegacyDec())
	supplyRatio := quote.Token.Amount.ToLegacyDec().Quo(base.Token.Amount.ToLegacyDec())
	spotPriceDec := supplyRatio.Mul(invWeightRatio)

	return spotPriceDec, err
}

func (p OsmosisGammPool) parsePoolAssetsByDenoms(tokenADenom, tokenBDenom string) (
	Aasset PoolAsset, Basset PoolAsset, err error,
) {
	Aasset, found1 := getPoolAssetByDenom(p.PoolAssets, tokenADenom)
	Basset, found2 := getPoolAssetByDenom(p.PoolAssets, tokenBDenom)

	if !found1 {
		return PoolAsset{}, PoolAsset{}, fmt.Errorf("(%s) does not exist in the pool", tokenADenom)
	}
	if !found2 {
		return PoolAsset{}, PoolAsset{}, fmt.Errorf("(%s) does not exist in the pool", tokenBDenom)
	}
	return Aasset, Basset, nil
}

func getPoolAssetByDenom(assets []PoolAsset, denom string) (PoolAsset, bool) {
	for _, asset := range assets {
		if asset.Token.Denom == denom {
			return asset, true
		}
	}
	return PoolAsset{}, false
}
