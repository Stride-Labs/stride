// #nosec G101
package types

const (
	DymensionChainId                   = "dymension_1100-1"
	StrideToDymensionTransferChannelId = "channel-197"
	DymensionNativeTokenDenom          = "adym"
	DymensionNativeTokenIBCDenom       = "ibc/E1C22332C083574F3418481359733BA8887D171E76C80AD9237422AEABB66018" // #nosec G101

	DelegationAddressOnDymension = "dym1d6ntc7s8gs86tpdyn422vsqc6uaz9cejnxz5p5" // C0
	RewardAddressOnDymension     = "dym15up3hegy8zuqhy0p9m8luh0c984ptu2g5p4xpf" // C1

	DepositAddress    = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04" // S0
	RedemptionAddress = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g" // S1
	ClaimAddress      = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd" // S2

	SafeAddressOnStride             = "stride18p7xg4hj2u3zpk0v9gq68pjyuuua5wa387sjjc" // S3
	OperatorAddressOnStride         = "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw" // OP-STRIDE
	DymensionUnbondingPeriodSeconds = uint64(21 * 24 * 60 * 60)                       // 21 days
)
