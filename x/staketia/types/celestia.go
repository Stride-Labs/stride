// #nosec G101
package types

const (
	CelestiaChainId                   = "mocha-4"
	StrideToCelestiaTransferChannelId = "channel-38"
	CelestiaNativeTokenDenom          = "utia"
	CelestiaNativeTokenIBCDenom       = "ibc/1A7653323C1A9E267FF7BEBF40B3EEA8065E8F069F47F2493ABC3E0B621BF793" // #nosec G101

	DelegationAddressOnCelestia = "celestia1cr67t725a76hyaapx3degxua3ey8zqrrxyxqu8" // C0
	RewardAddressOnCelestia     = "celestia18etj3sqjj49l2vna303fmszqf40j4k5wnwha0d" // C1

	DepositAddress    = "stride1e9gf63q0wpqyrge5xv2wkt7kz95wl5ntp94m9c" // S0
	RedemptionAddress = "stride1kjylv75j4t9za70k38uxnesywc4x620t42edt9" // S1
	ClaimAddress      = "stride1nm5hd2vvutksxu5cl35gn2k77yjrd5tdfeawy4" // S2

	SafeAddressOnStride            = "stride19zwp2gs73dwfa7zwvhc36lhcke6ur8k3plkrvy" // S3
	OperatorAddressOnStride        = "stride19zwp2gs73dwfa7zwvhc36lhcke6ur8k3plkrvy" // OP-STRIDE
	CelestiaUnbondingPeriodDays    = 21
	CelestiaUnbondingPeriodSeconds = uint64(CelestiaUnbondingPeriodDays * 24 * 60 * 60) // 21 days

	CelestiaBechPrefix = "celestia"
)

// The connection ID is stored as a var so it can be overriden in tests
var CelestiaConnectionId = "connection-26"
