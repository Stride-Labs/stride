package types

// DONTCOVER

import errorsmod "cosmossdk.io/errors"

// x/claim module sentinel errors
var (
	ErrTotalWeightNotSet = errorsmod.Register(ModuleName, 1101,
		"total weight not set")
	ErrTotalWeightParse = errorsmod.Register(ModuleName, 1102,
		"total weight parse error")
	ErrFailedToGetTotalWeight = errorsmod.Register(ModuleName, 1104,
		"failed to get total weight")
	ErrFailedToParseDec = errorsmod.Register(ModuleName, 1105,
		"failed to parse dec from str")
	ErrAirdropAlreadyExists = errorsmod.Register(ModuleName, 1106,
		"airdrop with same identifier already exists")
	ErrDistributorAlreadyExists = errorsmod.Register(ModuleName, 1107,
		"airdrop with same distributor already exists")
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1108,
		"cannot claim negative tokens")
	ErrInvalidAccount = errorsmod.Register(ModuleName, 1109,
		"only BaseAccount and StridePeriodicVestingAccount can claim")
	ErrAirdropNotFound = errorsmod.Register(ModuleName, 1110,
		"the airdrop was not found")
	ErrClaimNotFound = errorsmod.Register(ModuleName, 1111,
		"the claim record was not found")
	ErrModifyingClaimRecord = errorsmod.Register(ModuleName, 1112,
		"failed to modify claim record")
	ErrAirdropChainIdAlreadyExists = errorsmod.Register(ModuleName, 1113,
		"airdrop with same chain-id already exists")
)
