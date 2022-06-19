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
		ICAAccount:        nil,
		HostZoneList:      []HostZone{},
		DepositRecordList: []DepositRecord{},
		ControllerBalancesList: []ControllerBalances{},
// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
		PortId: PortID,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}
	// Check for duplicated ID in depositRecord
	depositRecordIdMap := make(map[uint64]bool)
	depositRecordCount := gs.GetDepositRecordCount()
	for _, elem := range gs.DepositRecordList {
		if _, ok := depositRecordIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for depositRecord")
		}
		if elem.Id >= depositRecordCount {
			return fmt.Errorf("depositRecord id should be lower or equal than the last id")
		}
		depositRecordIdMap[elem.Id] = true
	}
	// Check for duplicated index in controllerBalances
controllerBalancesIndexMap := make(map[string]struct{})

for _, elem := range gs.ControllerBalancesList {
	index := string(ControllerBalancesKey(elem.Index))
	if _, ok := controllerBalancesIndexMap[index]; ok {
		return fmt.Errorf("duplicated index for controllerBalances")
	}
	controllerBalancesIndexMap[index] = struct{}{}
}
// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
