package types

import (
	"errors"
	"strconv"
)

func ValidateTokenPriceQueryParams(baseDenom, quoteDenom, osmosisPoolId, osmosisBaseDenom, osmosisQuoteDenom string) error {
	if baseDenom == "" {
		return errors.New("base-denom must be specified")
	}
	if quoteDenom == "" {
		return errors.New("quote-denom must be specified")
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
