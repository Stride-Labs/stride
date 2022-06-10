package types

// // DefaultGenesis returns the default Capability genesis state
func NewDepositRecord(amount int32, denom string, hostZoneId string, sender string, purpose DepositRecord_Purpose) *DepositRecord {
	return &DepositRecord{
		Id:         0,
		Amount:     amount,
		Denom:      denom,
		HostZoneId: hostZoneId,
		Sender:     sender,
		Purpose:    purpose,
	}
}
