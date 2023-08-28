package app

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"
	ccvgov "github.com/cosmos/interchain-security/v3/x/ccv/democracy/governance"

	autopilottypes "github.com/Stride-Labs/stride/v14/x/autopilot/types"
	minttypes "github.com/Stride-Labs/stride/v14/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

var WhiteListModule = map[string]struct{}{
	"/cosmos.gov.v1.MsgUpdateParams":               {},
	"/cosmos.bank.v1beta1.MsgUpdateParams":         {},
	"/cosmos.staking.v1beta1.MsgUpdateParams":      {},
	"/cosmos.distribution.v1beta1.MsgUpdateParams": {},
	"/cosmos.mint.v1beta1.MsgUpdateParams":         {},
	"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade":   {},
	"/cosmos.upgrade.v1beta1.MsgCancelUpgrade":     {},
	"/stride.icaoracle.MsgToggleOracle":            {},
	"/stride.icaoracle.MsgRemoveOracle":            {},
}

func IsModuleWhiteList(typeUrl string) bool {
	_, found := WhiteListModule[typeUrl]
	return found
}

func IsProposalWhitelisted(content govv1beta1.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isParamChangeWhitelisted(getParamChangesMapFromArray(c.Changes))
	case *stakeibctypes.AddValidatorsProposal,
		*upgradetypes.SoftwareUpgradeProposal,       //nolint:staticcheck
		*upgradetypes.CancelSoftwareUpgradeProposal: //nolint:staticcheck
		return true

	default:
		return false
	}
}

func getParamChangesMapFromArray(paramChanges []proposal.ParamChange) map[ccvgov.ParamChangeKey]struct{} {
	res := map[ccvgov.ParamChangeKey]struct{}{}
	for _, paramChange := range paramChanges {
		key := ccvgov.ParamChangeKey{
			MsgType: paramChange.Subspace,
			Key:     paramChange.Key,
		}

		res[key] = struct{}{}
	}

	return res
}

func isParamChangeWhitelisted(paramChanges map[ccvgov.ParamChangeKey]struct{}) bool {
	for paramChangeKey := range paramChanges {
		_, found := WhitelistedParams[paramChangeKey]
		if !found {
			return false
		}
	}
	return true
}

var WhitelistedParams = map[ccvgov.ParamChangeKey]struct{}{
	//bank
	{MsgType: banktypes.ModuleName, Key: string(banktypes.KeySendEnabled)}:        {},
	{MsgType: banktypes.ModuleName, Key: string(banktypes.KeyDefaultSendEnabled)}: {},
	//governance
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyDepositParams)}: {}, //min_deposit, max_deposit_period
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyVotingParams)}:  {}, //voting_period
	{MsgType: govtypes.ModuleName, Key: string(govv1.ParamStoreKeyTallyParams)}:   {}, //quorum,threshold,veto_threshold
	//staking
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyUnbondingTime)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxValidators)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxEntries)}:        {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyHistoricalEntries)}: {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyBondDenom)}:         {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMinCommissionRate)}: {},
	//autopilot
	{MsgType: autopilottypes.ModuleName, Key: string(autopilottypes.KeyStakeibcActive)}: {},
	{MsgType: autopilottypes.ModuleName, Key: string(autopilottypes.KeyClaimActive)}:    {},
	//ccv consumer
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyRewardDenoms)}:                      {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyEnabled)}:                           {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyBlocksPerDistributionTransmission)}: {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyDistributionTransmissionChannel)}:   {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyProviderFeePoolAddrStr)}:            {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyTransferTimeoutPeriod)}:             {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyConsumerRedistributionFrac)}:        {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyHistoricalEntries)}:                 {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyConsumerUnbondingPeriod)}:           {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeySoftOptOutThreshold)}:               {},
	{MsgType: ccvconsumertypes.ModuleName, Key: string(ccvconsumertypes.KeyProviderRewardDenoms)}:              {},
	//distribution
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyCommunityTax)}:        {},
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyWithdrawAddrEnabled)}: {},
	//mint
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyMintDenom)}:                            {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyGenesisEpochProvisions)}:               {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyEpochIdentifier)}:                      {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyReductionPeriodInEpochs)}:              {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyReductionFactor)}:                      {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyPoolAllocationRatio)}:                  {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyDeveloperRewardsReceiver)}:             {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyMintingRewardsDistributionStartEpoch)}: {},
	//ibc transfer
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	//ibc staking
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDepositInterval)}:                   {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDelegateInterval)}:                  {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRewardsInterval)}:                   {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRedemptionRateInterval)}:            {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyStrideCommission)}:                  {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyReinvestInterval)}:                  {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyICATimeoutNanos)}:                   {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyBufferSize)}:                        {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIbcTimeoutBlocks)}:                  {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyFeeTransferTimeoutNanos)}:           {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyMaxStakeICACallsPerEpoch)}:          {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDefaultMinRedemptionRateThreshold)}: {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDefaultMaxRedemptionRateThreshold)}: {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIBCTransferTimeoutNanos)}:           {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyMaxRedemptionRates)}:                {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyMinRedemptionRates)}:                {},
	//ica
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
}
