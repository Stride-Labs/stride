package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
)

// Maintains the context and progress of an LSM Liquid Stake in the event
// that the transaction finishes asynchonously after the validator exchange rate query
type LSMLiquidStake struct {
	Staker      sdk.AccAddress               `json:"staker"`
	LSMIBCToken sdk.Coin                     `json:"lsm_ibc_token"`
	StToken     sdk.Coin                     `json:"st_token"`
	HostZone    HostZone                     `json:"host_zone"`
	Validator   Validator                    `json:"validator"`
	Deposit     recordstypes.LSMTokenDeposit `json:"deposit"`
}
