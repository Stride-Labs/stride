package utils

import sdk "github.com/cosmos/cosmos-sdk/types"

type PoaValidators struct {
	Name    string
	Address sdk.AccAddress
}

// TODO: Fil out addresses
var PoaValidatorSet = []PoaValidators{
	{
		Name:    "Polkachu",
		Address: nil,
	},
	{
		Name:    "Keplr",
		Address: nil,
	},
	{
		Name:    "Everstake",
		Address: nil,
	},
	{
		Name:    "Imperator",
		Address: nil,
	},
	{
		Name:    "L5",
		Address: nil,
	},
	{
		Name:    "Stakecito",
		Address: nil,
	},
	{
		Name:    "Cosmostation",
		Address: nil,
	},
	{
		Name:    "Citadel.one",
		Address: nil,
	},
}
