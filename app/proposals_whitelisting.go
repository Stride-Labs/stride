package app

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	minttypes "github.com/Stride-Labs/stride/v3/x/mint/types"
	stakeibctypes "github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

func IsProposalWhitelisted(content govtypes.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isParamChangeWhitelisted(c.Changes)
	case *stakeibctypes.AddValidatorProposal,
		*upgradetypes.SoftwareUpgradeProposal,
		*upgradetypes.CancelSoftwareUpgradeProposal:
		return true

	default:
		return false
	}
}

func isParamChangeWhitelisted(paramChanges []proposal.ParamChange) bool {
	for _, paramChange := range paramChanges {
		_, found := WhitelistedParams[paramChangeKey{Subspace: paramChange.Subspace, Key: paramChange.Key}]
		if !found {
			return false
		}
	}
	return true
}

type paramChangeKey struct {
	Subspace, Key string
}

var WhitelistedParams = map[paramChangeKey]struct{}{
	//bank
	{Subspace: banktypes.ModuleName, Key: string(banktypes.KeySendEnabled)}: {},
	//governance
	{Subspace: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyDepositParams)}: {}, //min_deposit, max_deposit_period
	{Subspace: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyVotingParams)}:  {}, //voting_period
	{Subspace: govtypes.ModuleName, Key: string(govtypes.ParamStoreKeyTallyParams)}:   {}, //quorum,threshold,veto_threshold
	//staking
	{Subspace: stakingtypes.ModuleName, Key: string(stakingtypes.KeyUnbondingTime)}:     {},
	{Subspace: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxValidators)}:     {},
	{Subspace: stakingtypes.ModuleName, Key: string(stakingtypes.KeyMaxEntries)}:        {},
	{Subspace: stakingtypes.ModuleName, Key: string(stakingtypes.KeyHistoricalEntries)}: {},
	{Subspace: stakingtypes.ModuleName, Key: string(stakingtypes.KeyBondDenom)}:         {},
	//distribution
	{Subspace: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyCommunityTax)}:        {},
	{Subspace: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyBaseProposerReward)}:  {},
	{Subspace: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyBonusProposerReward)}: {},
	{Subspace: distrtypes.ModuleName, Key: string(distrtypes.ParamStoreKeyWithdrawAddrEnabled)}: {},
	//mint
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyMintDenom)}:                            {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyGenesisEpochProvisions)}:               {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyEpochIdentifier)}:                      {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyReductionPeriodInEpochs)}:              {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyReductionFactor)}:                      {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyPoolAllocationRatio)}:                  {},
	{Subspace: minttypes.ModuleName, Key: string(minttypes.KeyMintingRewardsDistributionStartEpoch)}: {},
	//ibc transfer
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	//ibc staking
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDepositInterval)}:                  {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyDelegateInterval)}:                 {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRewardsInterval)}:                  {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyRedemptionRateInterval)}:           {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyStrideCommission)}:                 {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyReinvestInterval)}:                 {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyValidatorRebalancingThreshold)}:    {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyICATimeoutNanos)}:                  {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyBufferSize)}:                       {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIbcTimeoutBlocks)}:                 {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyFeeTransferTimeoutNanos)}:          {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyMaxStakeICACallsPerEpoch)}:         {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyMinRedemptionRateThreshold)}: {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyMaxRedemptionRateThreshold)}: {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeyIBCTransferTimeoutNanos)}:          {},
	{Subspace: stakeibctypes.ModuleName, Key: string(stakeibctypes.KeySafetyNumValidators)}:              {},
	//ica
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
}
