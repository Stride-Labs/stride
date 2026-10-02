// #nosec G101
package types

const (
	CelestiaChainId                   = "cosmoshub-test-1"
	StrideToCelestiaTransferChannelId = "channel-0"
	CelestiaNativeTokenDenom          = "uatom"
	CelestiaNativeTokenIBCDenom       = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2" // #nosec G101

	DelegationAddressOnCelestia = "cosmos1h0dup2qw23uhgn9nxyhyze4cxzrgu8rtrcnv7d" // C0
	RewardAddressOnCelestia     = "cosmos1mnx78sx5wcnutphpy6sfxfan6xnen07mrd2us6" // C1

	DepositAddress    = "stride1ju3xt2f8xuhzxqg6590sazctlz6l4md0wc5w6c" // S0
	RedemptionAddress = "stride19ksqv50zmntzjfflfmnegj75tdfkk89vl2q5yu" // S1
	ClaimAddress      = "stride1pjw24gg0fm26758hxee3wta35kq9jpszcslm6z" // S2

	SafeAddressOnStride            = "stride1tpzfseenwg4kq54sf9hdp3mkra652fvqtsuclq" // S3
	OperatorAddressOnStride        = "stride19xm04qaah8t2eupyeglz63vkaxzytpyc8m7kk4" // OP-STRIDE
	CelestiaUnbondingPeriodSeconds = uint64(240)                                     // rehearsal: 240s (mainnet: 14 days and one hour)

	CelestiaBechPrefix = "cosmos"
)

// The connection ID is stored as a var so it can be overriden in tests
var CelestiaConnectionId = "connection-0"
