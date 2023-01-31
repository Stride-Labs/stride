package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                    DefaultParams(),
		PortId:                    PortID,
		UserRedemptionRecordList:  []UserRedemptionRecord{},
		UserRedemptionRecordCount: sdkmath.ZeroInt(),
		EpochUnbondingRecordList:  []EpochUnbondingRecord{},
		DepositRecordList:         []DepositRecord{},
		DepositRecordCount:        sdkmath.ZeroInt(),
		// this line is used by starport scaffolding # genesis/types/default
	}
}

func (gs GenesisState) GetDepositRecordCount() sdkmath.Int {
	return gs.DepositRecordCount
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}
	// Check for duplicated ID in userRedemptionRecord
	userRedemptionRecordIdMap := make(map[string]bool)
	for _, elem := range gs.UserRedemptionRecordList {
		if _, ok := userRedemptionRecordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for userRedemptionRecord")
		}
		userRedemptionRecordIdMap[elem.Id] = true
	}
	// Check for duplicated ID in epochUnbondingRecord
	epochUnbondingRecordIdMap := make(map[uint64]bool)
	for _, elem := range gs.EpochUnbondingRecordList {
		if _, ok := epochUnbondingRecordIdMap[elem.EpochNumber.Uint64()]; ok {
			return fmt.Errorf("duplicated id for epochUnbondingRecord")
		}
		epochUnbondingRecordIdMap[elem.EpochNumber.Uint64()] = true
	}
	// Check for duplicated ID in depositRecord
	depositRecordIdMap := make(map[uint64]bool)
	depositRecordCount := gs.GetDepositRecordCount()
	for _, elem := range gs.DepositRecordList {
		if _, ok := depositRecordIdMap[elem.Id.Uint64()]; ok {
			return fmt.Errorf("duplicated id for depositRecord")
		}
		if elem.Id.GTE(depositRecordCount) {
			return fmt.Errorf("depositRecord id should be lower or equal than the last id")
		}
		depositRecordIdMap[elem.Id.Uint64()] = true
	}

	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
