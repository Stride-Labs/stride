package types

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

func NewDeposit(creator string, hostZoneId string, amount sdktypes.Int) (Deposit, error) {
	p := Deposit{
		Creator: creator,
		HostZoneId:  hostZoneId,
		Amount: amount,
	}

	return p, nil
}

// Proposals is an array of proposal
type Deposits []Deposit
