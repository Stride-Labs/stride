// #nosec G101
package types

const (
	DymensionChainId                   = "dymension_1100-1"
	StrideToDymensionTransferChannelId = "channel-197"
	DymensionNativeTokenDenom          = "adym"
	DymensionNativeTokenIBCDenom       = "ibc/E1C22332C083574F3418481359733BA8887D171E76C80AD9237422AEABB66018" // #nosec G101

	DelegationAddressOnDymension = "dym1nwpk5ekw74tl9eswvumvnwr3y7wwg99d8m0vzj" // C0
	RewardAddressOnDymension     = "dym1exfargyp2lpe2yscpe0406zzj4vfedqn5uvw72" // C1

	DepositAddress    = "stride1z9n8gk3837pagnnqv23ukruh8t606d0aj8u784" // S0
	RedemptionAddress = "stride10vmnxwgf4647nqxd9a6sh4qlyeytk5tpy2wtku" // S1
	ClaimAddress      = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd" // S2

	SafeAddressOnStride             = "stride1gatmguzwv9p6y8amz32457094z8hjevezlfp4m" // S3
	OperatorAddressOnStride         = "stride15s6xjemlhg3dqqeqhyu273ucfv56fss7l5tgf7" // OP-STRIDE
	DymensionUnbondingPeriodSeconds = uint64(21 * 24 * 60 * 60)                       // 21 days
)
