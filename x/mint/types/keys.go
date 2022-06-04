package types

// MinterKey is the key to use for the keeper store.
var MinterKey = []byte{0x00}

// LastReductionEpochKey is the key to use for the keeper store.
var LastReductionEpochKey = []byte{0x01}

const (
	// module name.
	ModuleName = "mint"
	// module acct name for developer vesting.
	DeveloperVestingModuleAcctName = "developer_vesting_unvested"

	// StoreKey is the default store key for mint.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the minting querier.
	QueryParameters      = "parameters"
	QueryEpochProvisions = "epoch_provisions"
)
