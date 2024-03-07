// #nosec G101
package types

const (
	DymensionChainId                   = "dymension_1100-1"
	StrideToDymensionTransferChannelId = "channel-197"
	DymensionNativeTokenDenom          = "adym"
	DymensionNativeTokenIBCDenom       = "ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA" // #nosec G101

	DelegationAddressOnDymension = "dym1gl9j2hyyukqvlmzzcxl99mqfgu4y4frgzlv3zz" // C0
	RewardAddressOnDymension     = "dym1ww3m6h5e3dk2musft9w8f2zu4rkuxgh6zwu0d0" // C1

	DepositAddress    = "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj" // S4
	RedemptionAddress = "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n" // S5
	ClaimAddress      = "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras" // S6

	SafeAddressOnStride             = "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7" // S7
	OperatorAddressOnStride         = "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw" // OP-STRIDE
	DymensionUnbondingPeriodSeconds = uint64(21 * 24 * 60 * 60)                       // 21 days
)
