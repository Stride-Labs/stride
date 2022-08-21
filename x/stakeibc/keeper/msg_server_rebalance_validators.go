package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	// maxNumRebalance := cast.ToInt(msg.NumRebalance)
	// if maxNumRebalance < 1 {
	// 	k.Logger(ctx).Error(fmt.Sprintf("Invalid number of validators to rebalance %d", maxNumRebalance))
	// 	return nil, types.ErrNoValidatorWeights
	// }

	// validatorDeltas, err := k.GetValidatorDelegationAmtDifferences(ctx, hostZone)
	// if err != nil {
	// 	k.Logger(ctx).Error(fmt.Sprintf("Error getting validator deltas for Host Zone %s: %s", hostZone.ChainId, err))
	// 	return nil, err
	// }

	// // we convert the above map into a list of tuples
	// type valPair struct {
	// 	deltaAmt int64
	// 	valAddr  string
	// }
	// valDeltaList := make([]valPair, 0)
	// for _, valAddr := range utils.StringToIntMapKeys(validatorDeltas) {
	// 	deltaAmt := validatorDeltas[valAddr]
	// 	k.Logger(ctx).Info(fmt.Sprintf("Adding deltaAmt: %d to validator: %s", deltaAmt, valAddr))
	// 	valDeltaList = append(valDeltaList, valPair{deltaAmt, valAddr})
	// }
	// // now we sort that list
	// lessFunc := func(i, j int) bool {
	// 	return valDeltaList[i].deltaAmt < valDeltaList[j].deltaAmt
	// }
	// sort.Slice(valDeltaList, lessFunc)
	// // now varDeltaList is sorted by deltaAmt
	// overWeightIndex := 0
	// underWeightIndex := len(valDeltaList) - 1

	// // check if there is a large enough rebalance, if not, just exit
	// total_delegation := float64(k.GetTotalValidatorDelegations(hostZone))
	// overweight_delta := floatabs(float64(valDeltaList[overWeightIndex].deltaAmt) / total_delegation)
	// underweight_delta := floatabs(float64(valDeltaList[underWeightIndex].deltaAmt) / total_delegation)
	// max_delta := floatmax(overweight_delta, underweight_delta)
	// rebalanceThreshold := float64(k.GetParam(ctx, types.KeyValidatorRebalancingThreshold)) / float64(10000)
	// if max_delta < rebalanceThreshold {
	// 	k.Logger(ctx).Error("Not enough validator disruption to rebalance")
	// 	return nil, types.ErrNoValidatorWeights
	// }

	// var msgs []sdk.Msg
	// delegationIca := hostZone.GetDelegationAccount()
	// if delegationIca == nil || delegationIca.GetAddress() == "" {
	// 	k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
	// 	return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
	// }

	// delegatorAddressStr := delegationIca.GetAddress()

	// for i := 1; i < maxNumRebalance; i++ {
	// 	underWeightElem := valDeltaList[underWeightIndex]
	// 	overWeightElem := valDeltaList[overWeightIndex]
	// 	if underWeightElem.deltaAmt < 0 {
	// 		// if underWeightElem is negative, we're done rebalancing
	// 		break
	// 	}
	// 	if overWeightElem.deltaAmt > 0 {
	// 		// if overWeightElem is positive, we're done rebalancing
	// 		break
	// 	}
	// 	if abs(underWeightElem.deltaAmt) > abs(overWeightElem.deltaAmt) {
	// 		// if the underweight element is more off than the overweight element
	// 		// we transfer stake from the underweight element to the overweight element
	// 		underWeightElem.deltaAmt -= abs(overWeightElem.deltaAmt)
	// 		overWeightIndex += 1
	// 		// issue an ICA call to the host zone to rebalance the validator
	// 		redelegateMsg := stakingTypes.MsgBeginRedelegate{
	// 			DelegatorAddress:    delegatorAddressStr,
	// 			ValidatorSrcAddress: overWeightElem.valAddr,
	// 			ValidatorDstAddress: underWeightElem.valAddr,
	// 			Amount:              sdk.NewInt64Coin(hostZone.HostDenom, abs(overWeightElem.deltaAmt)),
	// 		}
	// 		msgs = append(msgs, &redelegateMsg)
	// 		k.Logger(ctx).Info(fmt.Sprintf("Rebalancing validators, case 1, with msg %v", msgs))
	// 		overWeightElem.deltaAmt = 0
	// 	} else if abs(underWeightElem.deltaAmt) < abs(overWeightElem.deltaAmt) {
	// 		// if the overweight element is more overweight than the underweight element
	// 		overWeightElem.deltaAmt += underWeightElem.deltaAmt
	// 		underWeightIndex -= 1
	// 		// issue an ICA call to the host zone to rebalance the validator
	// 		redelegateMsg := stakingTypes.MsgBeginRedelegate{
	// 			DelegatorAddress:    delegatorAddressStr,
	// 			ValidatorSrcAddress: overWeightElem.valAddr,
	// 			ValidatorDstAddress: underWeightElem.valAddr,
	// 			Amount:              sdk.NewInt64Coin(hostZone.HostDenom, abs(underWeightElem.deltaAmt)),
	// 		}
	// 		msgs = append(msgs, &redelegateMsg)
	// 		k.Logger(ctx).Info(fmt.Sprintf("Rebalancing validators, case 2, with msg %v", msgs))
	// 		underWeightElem.deltaAmt = 0
	// 	} else {
	// 		// if the two elements are equal, we increment both slices
	// 		underWeightIndex -= 1
	// 		overWeightIndex += 1
	// 		// issue an ICA call to the host zone to rebalance the validator
	// 		redelegateMsg := stakingTypes.MsgBeginRedelegate{
	// 			DelegatorAddress:    delegatorAddressStr,
	// 			ValidatorSrcAddress: underWeightElem.valAddr,
	// 			ValidatorDstAddress: overWeightElem.valAddr,
	// 			Amount:              sdk.NewInt64Coin(hostZone.HostDenom, abs(underWeightElem.deltaAmt)),
	// 		}
	// 		k.Logger(ctx).Info(fmt.Sprintf("Rebalancing validators, case 3, with msg %v", msgs))
	// 		msgs = append(msgs, &redelegateMsg)
	// 		overWeightElem.deltaAmt = 0
	// 		underWeightElem.deltaAmt = 0
	// 	}
	// }
	// k.Logger(ctx).Info(fmt.Sprintf("Rebalancing validators with msg %v", msgs))

	// ===========================
	var msgs []sdk.Msg
	msgs = append(msgs, &stakingTypes.MsgBeginRedelegate{
		DelegatorAddress:    "cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm",
		ValidatorSrcAddress: "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne",
		ValidatorDstAddress: "cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz",
		Amount:              sdk.NewInt64Coin(hostZone.HostDenom, abs(10)),
	})
	k.Logger(ctx).Info(fmt.Sprintf("Rebalancing validators with msg %v", msgs))

	// ===========================

	connectionId := hostZone.GetConnectionId()
	_, err := k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *hostZone.GetDelegationAccount(), "", nil)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}

	return &types.MsgRebalanceValidatorsResponse{}, nil
}
