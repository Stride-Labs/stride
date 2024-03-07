// #nosec G101
package types

const (
	DymensionChainId                   = "dymension_1100-1"
	StrideToDymensionTransferChannelId = "channel-30"
	DymensionNativeTokenDenom          = "adym"
	DymensionNativeTokenIBCDenom       = "ibc/D88D8EB90FC32E1B8DF2DB0D61F5E0D704183F0CDB5C5FFB327606BF8ABF4486" // #nosec G101

	DelegationAddressOnDymension = "dym1d7cg95lhnm3dha0w3u4pdj74htwhm7y9clr49y" // C0
	RewardAddressOnDymension     = "dym1r7n32g433aqgdda3jxvdzhlz5qcfrjqj53cnml" // C1

	DepositAddress    = "stride1eegwnh63638vhd4rmsckqdhw09vasrsaus5eus" // S4
	RedemptionAddress = "stride14uddgyq3yy54yexpej3jq3cckwp34yq3es7gch" // S5
	ClaimAddress      = "stride1mrs0axtu39rag8qd8gjce89axdaxpcvxrzceln" // S6

	SafeAddressOnStride             = "stride1tvyr5yz2qpaquef73uzcluz8lex5jj9r4c3wpp" // S7
	OperatorAddressOnStride         = "stride17yyyw5dgcp8dwse8j566t6qu00efwh9remtk4p" // S8
	DymensionUnbondingPeriodSeconds = uint64(21 * 24 * 60 * 60)                       // 21 days
)
