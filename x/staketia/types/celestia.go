// #nosec G101
package types

const (
	CelestiaChainId                   = "celestia"
	StrideToCelestiaTransferChannelId = "channel-162"
	CelestiaNativeTokenDenom          = "utia"
	CelestiaNativeTokenIBCDenom       = "ibc/BF3B4F53F3694B66E13C23107C84B6485BD2B96296BB7EC680EA77BBA75B4801" // #nosec G101

	DelegationAddressOnCelestia = "celestia1d6ntc7s8gs86tpdyn422vsqc6uaz9cejnxz5p5" // C0
	RewardAddressOnCelestia     = "celestia15up3hegy8zuqhy0p9m8luh0c984ptu2g5p4xpf" // C1

	DepositAddress    = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04" // S0
	RedemptionAddress = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g" // S1
	ClaimAddress      = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd" // S2

	SafeAddressOnStride            = "stride18p7xg4hj2u3zpk0v9gq68pjyuuua5wa387sjjc" // S3
	OperatorAddressOnStride        = "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw" // OP-STRIDE
	CelestiaUnbondingPeriodSeconds = uint64(1213200)                                 // 14 days and one hour

	CelestiaBechPrefix = "celestia"
)

// The connection ID is stored as a var so it can be overriden in tests
var CelestiaConnectionId = "connection-125"
