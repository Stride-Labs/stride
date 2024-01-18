package types

const (
	ModuleName = "staketia"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	// Prefix store keys
	HostZoneKey                         = []byte("host-zone")
	DelegationRecordsKeyPrefix          = []byte("delegation-records")
	UnbondingRecordsKeyPrefix           = []byte("unbonding-records")
	RedemptionRecordsKeyPrefix          = []byte("redemption-records")
	SlashRecordsKeyPrefix               = []byte("slash-records")
	TransferInProgressRecordIdKeyPrefix = []byte("transfer-in-progress")
)
