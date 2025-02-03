package types

import (
	"errors"
	"strconv"
)

func ValidateTokenPriceQueryParams(
	baseDenom string,
	quoteDenom string,
	baseDenomDecimals int64,
	quoteDenomDecimals int64,
	osmosisPoolId string,
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
	if _, err := strconv.ParseUint(osmosisPoolId, 10, 64); err != nil {
		return errors.New("osmosis-pool-id must be uint64")
	}
	if osmosisBaseDenom == "" {
		return errors.New("osmosis-base-denom must be specified")
	}
	if osmosisQuoteDenom == "" {
		return errors.New("osmosis-quote-denom must be specified")
	}

	return nil
}
