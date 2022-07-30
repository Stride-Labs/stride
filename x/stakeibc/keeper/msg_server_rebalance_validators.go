package keeper

import (
	"context"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/utils"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func floatabs(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

func floatmax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found %s", msg.HostZone))
		return nil, types.ErrInvalidHostZone
	}
	maxNumRebalance := cast.ToInt(msg.NumRebalance)
	if maxNumRebalance < 1 {
		k.Logger(ctx).Error(fmt.Sprintf("Invalid number of validators to rebalance %d", maxNumRebalance))
		return nil, types.ErrNoValidatorWeights
	}

	validatorDeltas, err := k.GetValidatorDelegationAmtDifferences(ctx, hostZone)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting validator deltas for Host Zone %s: %s", hostZone.ChainId, err))
		return nil, err
	}

	// we convert the above map into a list of tuples
	type valPair struct {
		deltaAmt int64
		valAddr  sdk.ValAddress
	}
	valDeltaList := make([]valPair, 0)
	for _, valAddr := range utils.StringToIntMapKeys(validatorDeltas) {
		deltaAmt := validatorDeltas[valAddr]
		valDeltaList = append(valDeltaList, valPair{deltaAmt, sdk.ValAddress(valAddr)})
	}
	// now we sort that list
	lessFunc := func(i, j int) bool {
		return valDeltaList[i].deltaAmt < valDeltaList[j].deltaAmt
	}
	sort.Slice(valDeltaList, lessFunc)
	// now varDeltaList is sorted by deltaAmt
	overWeightIndex := 0
	underWeightIndex := len(valDeltaList) - 1

	// check if there is a large enough rebalance, if not, just exit
	total_delegation := float64(k.GetTotalValidatorDelegations(hostZone))
	overweight_delta := floatabs(float64(valDeltaList[overWeightIndex].deltaAmt) / total_delegation)
	underweight_delta := floatabs(float64(valDeltaList[underWeightIndex].deltaAmt) / total_delegation)
	max_delta := floatmax(overweight_delta, underweight_delta)
	rebalanceThreshold := float64(k.GetParam(ctx, types.KeyValidatorRebalancingThreshold)) / float64(10000)
	if max_delta < rebalanceThreshold {
		k.Logger(ctx).Error("Not enough validator disruption to rebalance")
		return nil, types.ErrNoValidatorWeights
	}

	var msgs []sdk.Msg
	delegatorAddressStr := hostZone.GetDelegationAccount().GetAddress()
	delegatorAddress := sdk.AccAddress(delegatorAddressStr)

	for i := 1; i < maxNumRebalance; i++ {
		underWeightElem := valDeltaList[underWeightIndex]
		overWeightElem := valDeltaList[overWeightIndex]
		if underWeightElem.deltaAmt < 0 {
			// if underWeightElem is negative, we're done rebalancing
			break
		}
		if overWeightElem.deltaAmt > 0 {
			// if overWeightElem is positive, we're done rebalancing
			break
		}
		if abs(underWeightElem.deltaAmt) > abs(overWeightElem.deltaAmt) {
			// if the underweight element is more off than the overweight element
			// we transfer stake from the underweight element to the overweight element
			underWeightElem.deltaAmt -= abs(overWeightElem.deltaAmt)
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelagateMsg := stakingTypes.NewMsgBeginRedelegate(
				delegatorAddress,
				overWeightElem.valAddr,
				underWeightElem.valAddr,
				sdk.NewInt64Coin(hostZone.HostDenom, abs(overWeightElem.deltaAmt)))
			msgs = append(msgs, redelagateMsg)
			overWeightElem.deltaAmt = 0
		} else if abs(underWeightElem.deltaAmt) < abs(overWeightElem.deltaAmt) {
			// if the overweight element is more overweight than the underweight element
			overWeightElem.deltaAmt += underWeightElem.deltaAmt
			underWeightIndex -= 1
			// issue an ICA call to the host zone to rebalance the validator
			redelagateMsg := stakingTypes.NewMsgBeginRedelegate(
				delegatorAddress,
				overWeightElem.valAddr,
				underWeightElem.valAddr,
				sdk.NewInt64Coin(hostZone.HostDenom, abs(underWeightElem.deltaAmt)))
			msgs = append(msgs, redelagateMsg)
			underWeightElem.deltaAmt = 0
		} else {
			// if the two elements are equal, we increment both slices
			underWeightIndex -= 1
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelagateMsg := stakingTypes.NewMsgBeginRedelegate(
				delegatorAddress,
				underWeightElem.valAddr,
				overWeightElem.valAddr,
				sdk.NewInt64Coin(hostZone.HostDenom, abs(underWeightElem.deltaAmt)))
			msgs = append(msgs, redelagateMsg)
			overWeightElem.deltaAmt = 0
			underWeightElem.deltaAmt = 0
		}
	}

	connectionId := hostZone.GetConnectionId()
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *hostZone.GetDelegationAccount())
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}

	return &types.MsgRebalanceValidatorsResponse{}, nil
}
