package types

import (
	"fmt"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PortId:                   PortID,
		UserRedemptionRecordList: []UserRedemptionRecord{},
		EpochUnbondingRecordList: []EpochUnbondingRecord{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}
	// Check for duplicated ID in userRedemptionRecord
	userRedemptionRecordIdMap := make(map[uint64]bool)
	userRedemptionRecordCount := gs.GetUserRedemptionRecordCount()
	for _, elem := range gs.UserRedemptionRecordList {
		if _, ok := userRedemptionRecordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for userRedemptionRecord")
		}
		if elem.Id >= userRedemptionRecordCount {
			return fmt.Errorf("userRedemptionRecord id should be lower or equal than the last id")
		}
		userRedemptionRecordIdMap[elem.Id] = true
	}
	// Check for duplicated ID in epochUnbondingRecord
	epochUnbondingRecordIdMap := make(map[uint64]bool)
	epochUnbondingRecordCount := gs.GetEpochUnbondingRecordCount()
	for _, elem := range gs.EpochUnbondingRecordList {
		if _, ok := epochUnbondingRecordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for epochUnbondingRecord")
		}
		if elem.Id >= epochUnbondingRecordCount {
			return fmt.Errorf("epochUnbondingRecord id should be lower or equal than the last id")
		}
		epochUnbondingRecordIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
