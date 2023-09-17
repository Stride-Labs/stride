package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

func (k Keeper) GetNextHubUnbonding(ctx sdk.Context) error {
	// log that we started getnextHubUnbonding
	k.Logger(ctx).Info("GetNextHubUnbonding started, moose")
	hubChainId := "cosmoshub-4" // TODO change this to cosmoshub-4
	hostZone, found := k.GetHostZone(ctx, hubChainId)
	if !found {
		return fmt.Errorf("host zone %s not found", hubChainId) //, nil, nil, nil
	}

	// Iterate through every unbonding record and sum the total amount to unbond for the given host zone
	totalUnbondAmount, _ := k.GetTotalUnbondAmountAndRecordsIds(ctx, hostZone.ChainId)
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Total unbonded amount: %v%s", totalUnbondAmount, hostZone.HostDenom))

	// If there's nothing to unbond, return and move on to the next host zone
	if totalUnbondAmount.IsZero() {
		return nil //, nil, nil, nil
	}

	// >>>>>>>>>>>>>>>>>>>>>>>>>> SET WEIGHTS, TOTALTOUNBOND >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	// TODO do we need to set the recordsIds too?

	// log the current totalUnbondAmount then set the totalUnbondAmount
	k.Logger(ctx).Info(fmt.Sprintf("totalUnbondedAmount before clobbering it: %v", totalUnbondAmount))
	totalUnbondAmount = sdk.NewInt(1850000000000)

	// set new val weights
	newWeights := make(map[string]uint64)
	newWeights["cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv"] = 1
	newWeights["cosmosvaloper1083svrca4t350mphfv9x45wq9asrs60cdmrflj"] = 0
	newWeights["cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"] = 0
	newWeights["cosmosvaloper1ukpah0340rx7k3x2njnavwyjv6pfpvn632df9q"] = 0
	newWeights["cosmosvaloper1gp957czryfgyvxwn3tfnyy2f0t9g2p4pqeemx8"] = 0
	newWeights["cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc"] = 0
	newWeights["cosmosvaloper14qazscc80zgzx3m0m0aa30ths0p9hg8vdglqrc"] = 0
	newWeights["cosmosvaloper16k579jk6yt2cwmqx9dz5xvq9fug2tekvlu9qdv"] = 1
	newWeights["cosmosvaloper106yp7zw35wftheyyv9f9pe69t8rteumjrx52jg"] = 0
	newWeights["cosmosvaloper124maqmcqv8tquy764ktz7cu0gxnzfw54n3vww8"] = 0
	newWeights["cosmosvaloper140l6y2gp3gxvay6qtn70re7z2s0gn57zfd832j"] = 1
	newWeights["cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnfz55jv3vp77d"] = 1
	newWeights["cosmosvaloper140e7u946a2nqqkvcnjpjm83d0ynsqem8dnp684"] = 0
	newWeights["cosmosvaloper140l6y2gp3gxvay6qtn70re7z2s0gn57zfd832j"] = 1

	// iterate the hostZone validator weights to set the new weights
	for _, val := range hostZone.Validators {
		// if val is in newWeight keys, set its weight to the new weight
		if _, ok := newWeights[val.Address]; ok {
			val.Weight = newWeights[val.Address]
		}
	}
	// iterate the vals and print their addresses and weights
	for _, val := range hostZone.Validators {
		k.Logger(ctx).Info(fmt.Sprintf("val address: %v, val weight: %v, val delegation: %v", val.Address, val.Weight, val.Delegation))
	}
	k.Logger(ctx).Info(fmt.Sprintf("totalUnbondedAmount after clobbering it: %v", totalUnbondAmount))
	k.SetHostZone(ctx, hostZone)

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

	// Determine the ideal balanced delegation for each validator after the unbonding
	//   (as if we were to unbond and then rebalance)
	// This will serve as the starting point for determining how much to unbond each validator
	delegationAfterUnbonding := hostZone.TotalDelegations.Sub(totalUnbondAmount)
	balancedDelegationsAfterUnbonding, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, delegationAfterUnbonding)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId) //, nil, nil, nil
	}

	// Determine the unbond capacity for each validator
	// Each validator can only unbond up to the difference between their current delegation and their balanced delegation
	// The validator's current delegation will be above their balanced delegation if they've received LSM Liquid Stakes
	//   (which is only rebalanced once per unbonding period)
	validatorUnbondCapacity := k.GetValidatorUnbondCapacity(ctx, hostZone.Validators, balancedDelegationsAfterUnbonding)
	if len(validatorUnbondCapacity) == 0 {
		return fmt.Errorf("there are no validators on %s with sufficient unbond capacity", hostZone.ChainId) //, nil, nil, nil
	}

	// Sort the unbonding capacity by priority
	// Priority is determined by checking the how proportionally unbalanced each validator is
	// Zero weight validators will come first in the list
	prioritizedUnbondCapacity, err := SortUnbondingCapacityByPriority(validatorUnbondCapacity)
	if err != nil {
		return err //, nil, nil, nil
	}

	// Get the undelegation ICA messages and split delegations for the callback
	msgs, unbondings, err := k.GetUnbondingICAMessages(hostZone, totalUnbondAmount, prioritizedUnbondCapacity)
	if err != nil {
		return err //, nil, nil, nil
	}

	// Shouldn't be possible, but if all the validator's had a target unbonding of zero, do not send an ICA
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target unbonded amount was 0 for each validator") //, nil, nil, nil
	}

	// log the messages
	for _, msg := range msgs {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"Undelegate ICA message: %s", msg.String()))
	}
	// print the full msgs array
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Undelegate ICA messages: %v", msgs))

	// print the unbondings
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Undelegate ICA unbondings: %v", unbondings))

	// // print the epochUnbondingRecordIds
	// k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
	// 	"Undelegate ICA epochUnbondingRecordIds: %v", epochUnbondingRecordIds))

	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

	return nil //, msgs, unbondings, epochUnbondingRecordIds
}
