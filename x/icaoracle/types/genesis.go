package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		Oracles:        []Oracle{},
		QueuedMetrics:  []Metric{},
		PendingMetrics: []PendingMetricUpdate{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return errorsmod.Wrapf(ErrInvalidGenesisState, err.Error())
	}
	for _, oracle := range gs.Oracles {
		if oracle.ChainId == "" {
			return errorsmod.Wrapf(ErrInvalidGenesisState, "oracle has empty chain ID")
		}
	}
	for _, metric := range gs.QueuedMetrics {
		if metric.Key == "" || metric.UpdateTime == 0 {
			return errorsmod.Wrapf(ErrInvalidGenesisState, "metric has empty key or update time")
		}
	}
	for _, pendingMetricUpdate := range gs.PendingMetrics {
		if pendingMetricUpdate.Metric.Key == "" || pendingMetricUpdate.OracleChainId == "" || pendingMetricUpdate.Metric.UpdateTime == 0 {
			return errorsmod.Wrapf(ErrInvalidGenesisState, "pending metric update has empty key, oracle chain ID, or update time")
		}
	}

	return nil
}
