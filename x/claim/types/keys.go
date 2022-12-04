package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "claim"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// ClaimRecordsStorePrefix defines the store prefix for the claim records
	ClaimRecordsStorePrefix = "claimrecords"

	// ParamsKey defines the store key for claim module parameters
	ParamsKey = "params"

	// ActionKey defines the store key to store user accomplished actions
	ActionKey = "action"

	// TotalWeightKey defines the store key for total weight
	TotalWeightKey = "totalweight"
)

var (
	// Percentages for actions
	PercentageForFree        = sdk.NewDecWithPrec(20, 2) //20%
	PercentageForStake       = sdk.NewDecWithPrec(20, 2) //20%
	PercentageForLiquidStake = sdk.NewDecWithPrec(60, 2) //60%
)
