package types

import (
	"errors"
)

func ValidateTokenPriceQueryParams(
	baseDenom string,
	quoteDenom string,
	baseDenomDecimals int64,
	quoteDenomDecimals int64,
	osmosisPoolId uint64,
	osmosisBaseDenom string,
	osmosisQuoteDenom string,
) error {
	if baseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if quoteDenom == "" {
		return errors.New("quote-denom must be specified")
	}
	if baseDenomDecimals <= 0 {
		return errors.New("base-denom-decimals must be bigger than 0")
	}
	if quoteDenomDecimals <= 0 {
		return errors.New("quote-denom-decimals must be bigger than 0")
	}
	if osmosisPoolId == 0 {
		return errors.New("osmosis-pool-id must be specified")
	}
	if osmosisBaseDenom == "" {
		return errors.New("osmosis-base-denom must be specified")
	}
	if osmosisQuoteDenom == "" {
		return errors.New("osmosis-quote-denom must be specified")
	}

	return nil
}
