package keeper

import (
	"context"
	"fmt"
	"sort"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		k.Logger(ctx).Info(fmt.Sprintf("Host Zone not found %s", msg.HostZone))
		return nil, types.ErrInvalidHostZone
	}
	maxNumRebalance := int(msg.NumRebalance)
	if maxNumRebalance < 1 {
		k.Logger(ctx).Info(fmt.Sprintf("Invalid number of validators to rebalance %d", maxNumRebalance))
		return nil, types.ErrNoValidatorWeights
	}

	validatorDeltas, err := k.GetValidatorDelegationAmtDifferences(ctx, hostZone)
	if err != nil {
		k.Logger(ctx).Info(fmt.Sprintf("Error getting validator deltas for Host Zone %s: %s", hostZone.ChainId, err))
		return nil, err
	}

	// we convert the above map into a list of tuples
	type valPair struct {
		deltaAmt int64
		valAddr  sdk.ValAddress
	}
	valDeltaList := make([]valPair, 0)
	for valAddr, deltaAmt := range validatorDeltas {
		valDeltaList = append(valDeltaList, valPair{deltaAmt, sdk.ValAddress(valAddr)})
	}
	// now we sort that list
	lessFunc := func(i, j int) bool {
		return valDeltaList[i].deltaAmt < valDeltaList[j].deltaAmt
	}
	sort.Slice(valDeltaList, lessFunc)
	// now varDeltaList is sorted by deltaAmt
	underWeightIndex := 0
	overWeightIndex := len(valDeltaList) - 1

	var msgs []sdk.Msg
	delegatorAddressStr := hostZone.DelegationAccount.Address
	delegatorAddress := sdk.AccAddress(delegatorAddressStr)

	// max_delta = utils.abs(valDeltaList[underWeightIndex].deltaAmt / )
	// max_delta = max(max_delta, utils.abs(valDeltaList[overWeightIndex].deltaAmt))
	// if max_delta < 0 {

	// }

	for i := 1; i < maxNumRebalance; i++ {
		underWeightElem := valDeltaList[underWeightIndex]
		overWeightElem := valDeltaList[overWeightIndex]
		if underWeightElem.deltaAmt > 0 {
			// if underWeightElem is positive, we're done rebalancing
			break
		}
		if overWeightElem.deltaAmt < 0 {
			// if overWeightElem is negative, we're done rebalancing
			break
		}
		if abs(underWeightElem.deltaAmt) > abs(overWeightElem.deltaAmt) {
			// if the underweight element is more overweight than the overweight element
			// we transfer stake from the underweight element to the overweight element
			underWeightElem.deltaAmt += overWeightElem.deltaAmt
			overWeightIndex -= 1
			// issue an ICA call to the host zone to rebalance the validator
			redelagateMsg := stakingTypes.NewMsgBeginRedelegate(
				delegatorAddress,
				underWeightElem.valAddr,
				overWeightElem.valAddr,
				sdk.NewInt64Coin(hostZone.HostDenom, abs(overWeightElem.deltaAmt)))
			msgs = append(msgs, redelagateMsg)
			overWeightElem.deltaAmt = 0
		} else if abs(underWeightElem.deltaAmt) < abs(overWeightElem.deltaAmt) {
			// if the overweight element is more overweight than the underweight element
			overWeightElem.deltaAmt -= underWeightElem.deltaAmt
			underWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelagateMsg := stakingTypes.NewMsgBeginRedelegate(
				delegatorAddress,
				underWeightElem.valAddr,
				overWeightElem.valAddr,
				sdk.NewInt64Coin(hostZone.HostDenom, abs(underWeightElem.deltaAmt)))
			msgs = append(msgs, redelagateMsg)
			underWeightElem.deltaAmt = 0
		} else {
			// if the two elements are equal, we increment both slices
			underWeightIndex += 1
			overWeightIndex -= 1
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
	err = k.SubmitTxs(ctx, connectionId, msgs, *hostZone.DelegationAccount)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}

	return &types.MsgRebalanceValidatorsResponse{}, nil
}
