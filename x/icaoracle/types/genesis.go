package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:  DefaultParams(),
		Oracles: []Oracle{},
		Metrics: []Metric{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid genesis state")
	}
	for _, oracle := range gs.Oracles {
		if oracle.ChainId == "" {
			return errorsmod.Wrap(ErrInvalidGenesisState, "oracle has empty chain ID")
		}
	}
	for _, metric := range gs.Metrics {
		if metric.Key == "" {
			return errorsmod.Wrap(ErrInvalidGenesisState, "metric has missing key")
		}
		if metric.UpdateTime == 0 {
			return errorsmod.Wrap(ErrInvalidGenesisState, "metric has missing time")
		}
		if metric.DestinationOracle == "" {
			return errorsmod.Wrap(ErrInvalidGenesisState, "metric has missing destination oracle chain ID")
		}
	}

	return nil
}
