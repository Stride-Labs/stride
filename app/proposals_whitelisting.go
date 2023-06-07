package app

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ccvgov "github.com/cosmos/interchain-security/x/ccv/democracy/governance"

	minttypes "github.com/Stride-Labs/stride/v10/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v10/x/stakeibc/types"
)

func IsProposalWhitelisted(content govv1beta1.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isParamChangeWhitelisted(getParamChangesMapFromArray(c.Changes))
	case *stakeibctypes.AddValidatorsProposal,
		*upgradetypes.SoftwareUpgradeProposal,
		*upgradetypes.CancelSoftwareUpgradeProposal:
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
	for paramChangeKey, _ := range paramChanges {
		_, found := WhitelistedParams[paramChangeKey]
		if !found {
			return false
		}
	}
	return true
}

var WhitelistedParams = map[ccvgov.ParamChangeKey]struct{}{
	//bank
	{MsgType: banktypes.ModuleName, Key: string(banktypes.KeySendEnabled)}: {},
	//governance
	// {MsgType: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyDepositParams)}: {}, //min_deposit, max_deposit_period
	// {MsgType: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyVotingParams)}:  {}, //voting_period
	// {MsgType: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyTallyParams)}:   {}, //quorum,threshold,veto_threshold
	//staking
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyUnbondingTime)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxValidators)}:     {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxEntries)}:        {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyHistoricalEntries)}: {},
	{MsgType: stakingtypes.ModuleName, Key: string(stakingtypes.KeyBondDenom)}:         {},
	//distribution
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyCommunityTax)}: {},
	// {MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyBaseProposerReward)}:  {},
	// {MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyBonusProposerReward)}: {},
	{MsgType: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyWithdrawAddrEnabled)}: {},
	//mint
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyMintDenom)}:                            {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyGenesisEpochProvisions)}:               {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyEpochIdentifier)}:                      {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyReductionPeriodInEpochs)}:              {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyReductionFactor)}:                      {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyPoolAllocationRatio)}:                  {},
	{MsgType: minttypes.ModuleName, Key: string(minttypes.KeyMintingRewardsDistributionStartEpoch)}: {},
	//ibc transfer
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{MsgType: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	//ibc staking
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDepositInterval)}:        {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDelegateInterval)}:       {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRewardsInterval)}:        {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRedemptionRateInterval)}: {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyStrideCommission)}:       {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyReinvestInterval)}:       {},
	// {MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyValidatorRebalancingThreshold)}:    {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyICATimeoutNanos)}:          {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyBufferSize)}:               {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIbcTimeoutBlocks)}:         {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyFeeTransferTimeoutNanos)}:  {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyMaxStakeICACallsPerEpoch)}: {},
	// {MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyMinRedemptionRateThreshold)}: {},
	// {MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyMaxRedemptionRateThreshold)}: {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIBCTransferTimeoutNanos)}: {},
	{MsgType: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyNumValidators)}:     {},
	//ica
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{MsgType: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
}
