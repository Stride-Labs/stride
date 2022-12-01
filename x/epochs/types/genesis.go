package types

import (
	"errors"
	"time"
)

func NewGenesisState(epochs []EpochInfo) *GenesisState {
	return &GenesisState{Epochs: epochs}
}

var (
	STRIDE_EPOCH = "stride_epoch"
	DAY_EPOCH    = "day"
	MINT_EPOCH   = "mint"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	epochs := []EpochInfo{
		{
			Identifier:              "week",
			StartTime:               time.Time{},
			Duration:                time.Hour * 24 * 7,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              DAY_EPOCH,
			StartTime:               time.Time{},
			Duration:                time.Hour * 24,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              STRIDE_EPOCH,
			StartTime:               time.Time{},
			Duration:                time.Hour * 6,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
		{
			Identifier:              MINT_EPOCH,
			StartTime:               time.Time{},
			Duration:                time.Minute * 60,
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
			CurrentEpochStartTime:   time.Time{},
			EpochCountingStarted:    false,
		},
	}
	return NewGenesisState(epochs)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	epochIdentifiers := map[string]bool{}
	for _, epoch := range gs.Epochs {
		if epoch.Identifier == "" {
			return errors.New("epoch identifier should NOT be empty")
		}
		if epochIdentifiers[epoch.Identifier] {
			return errors.New("epoch identifier should be unique")
		}
		if epoch.Duration == 0 {
			return errors.New("epoch duration should NOT be 0")
		}
		// enforce EpochCountingStarted is false for all epochs
		if epoch.EpochCountingStarted {
			return errors.New("epoch counting should NOT be started at genesis")
		}
		epochIdentifiers[epoch.Identifier] = true
	}
	return nil
}
