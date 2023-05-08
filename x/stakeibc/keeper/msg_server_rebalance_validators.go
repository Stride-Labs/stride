package keeper

import (
	"context"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("RebalanceValidators executing %v", msg))

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found %s", msg.HostZone))
		return nil, types.ErrInvalidHostZone
	}

	validatorDeltas, err := k.GetValidatorDelegationAmtDifferences(ctx, hostZone)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting validator deltas for Host Zone %s: %s", hostZone.ChainId, err))
		return nil, err
	}

	// we convert the above map into a list of tuples
	type valPair struct {
		deltaAmt sdkmath.Int
		valAddr  string
	}
	valDeltaList := make([]valPair, 0)
	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for _, valAddr := range utils.StringMapKeys(validatorDeltas) {
		deltaAmt := validatorDeltas[valAddr]
		k.Logger(ctx).Info(fmt.Sprintf("Adding deltaAmt: %v to validator: %s", deltaAmt, valAddr))
		valDeltaList = append(valDeltaList, valPair{deltaAmt, valAddr})
	}
	// now we sort that list
	lessFunc := func(i, j int) bool {
		return valDeltaList[i].deltaAmt.LT(valDeltaList[j].deltaAmt)
	}
	sort.SliceStable(valDeltaList, lessFunc)
	// now varDeltaList is sorted by deltaAmt
	overWeightIndex := 0
	underWeightIndex := len(valDeltaList) - 1

	// check if there is a large enough rebalance, if not, just exit
	total_delegation := k.GetTotalValidatorDelegations(hostZone)
	if total_delegation.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "no validator delegations found for Host Zone %s, cannot rebalance 0 delegations!", hostZone.ChainId)
	}

	var msgs []proto.Message
	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.GetAddress() == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
	}

	delegatorAddress := delegationIca.GetAddress()

	// start construction callback
	rebalanceCallback := types.RebalanceCallback{
		HostZoneId:   hostZone.ChainId,
		Rebalancings: []*types.Rebalancing{},
	}

	for i := uint64(1); i <= msg.NumRebalance; i++ {
		underWeightElem := valDeltaList[underWeightIndex]
		overWeightElem := valDeltaList[overWeightIndex]
		if underWeightElem.deltaAmt.LT(sdkmath.ZeroInt()) {
			// if underWeightElem is negative, we're done rebalancing
			break
		}
		if overWeightElem.deltaAmt.GT(sdkmath.ZeroInt()) {
			// if overWeightElem is positive, we're done rebalancing
			break
		}
		// underweight Elem is positive, overweight Elem is negative
		overWeightElemAbs := overWeightElem.deltaAmt.Abs()
		var redelegateMsg *stakingTypes.MsgBeginRedelegate
		if underWeightElem.deltaAmt.GT(overWeightElemAbs) {
			// if the underweight element is more off than the overweight element
			// we transfer stake from the underweight element to the overweight element
			underWeightElem.deltaAmt = underWeightElem.deltaAmt.Sub(overWeightElemAbs)
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingTypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, overWeightElemAbs)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.deltaAmt = sdkmath.ZeroInt()
		} else if underWeightElem.deltaAmt.LT(overWeightElemAbs) {
			// if the overweight element is more overweight than the underweight element
			overWeightElem.deltaAmt = overWeightElem.deltaAmt.Add(underWeightElem.deltaAmt)
			underWeightIndex -= 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingTypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.deltaAmt)}
			msgs = append(msgs, redelegateMsg)
			underWeightElem.deltaAmt = sdkmath.ZeroInt()
		} else {
			// if the two elements are equal, we increment both slices
			underWeightIndex -= 1
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingTypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.deltaAmt)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.deltaAmt = sdkmath.ZeroInt()
			underWeightElem.deltaAmt = sdkmath.ZeroInt()
		}
		// add the rebalancing to the callback
		// lastMsg grabs rebalanceMsg from above (due to the type, it's hard to )
		// lastMsg := (msgs[len(msgs)-1]).(*stakingTypes.MsgBeginRedelegate)
		rebalanceCallback.Rebalancings = append(rebalanceCallback.Rebalancings, &types.Rebalancing{
			SrcValidator: redelegateMsg.ValidatorSrcAddress,
			DstValidator: redelegateMsg.ValidatorDstAddress,
			Amt:          redelegateMsg.Amount.Amount,
		})
	}
	// marshall the callback
	marshalledCallbackArgs, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	connectionId := hostZone.GetConnectionId()
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *hostZone.GetDelegationAccount(), ICACallbackID_Rebalance, marshalledCallbackArgs)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s, %s", connectionId, hostZone.ChainId, msgs, err.Error())
	}

	return &types.MsgRebalanceValidatorsResponse{}, nil
}
