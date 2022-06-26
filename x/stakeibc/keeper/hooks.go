package keeper

import (
	"fmt"
	"strconv"

	utils "github.com/Stride-Labs/stride/utils"
	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

// TODO: ensure all timeouts are less than the epoch length
// TODO: add events from event manager, e.g.
// ctx.EventManager().EmitEvents(sdk.Events{
// 	sdk.NewEvent(
// 		sdk.EventTypeMessage,
// 		sdk.NewAttribute("hostZone", zoneInfo.ChainId),
// 		sdk.NewAttribute("newAmountStaked", balance.String()),
// 	),
// })

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s %d", epochIdentifier, epochNumber))
	if epochIdentifier == epochstypes.STRIDE_EPOCH {
		k.Logger(ctx).Info(fmt.Sprintf("Stride Epoch %d", epochNumber))

		// NOTE: We could nest this under `if epochNumber%depositInterval == 0 {`
		// -- should we?
		// e.g. CreateDepositRecordsForDepositInterval
		// Imagine it will be slightly cleaner to track state by epoch, rather than
		// by DepositInterval
		if epochNumber < 0 {
			k.Logger(ctx).Error(fmt.Sprintf("Stride Epoch %d negative", epochNumber))
			return
		}
		epochTracker := types.EpochTracker{
			EpochIdentifier: epochIdentifier,
			EpochNumber:     uint64(epochNumber),
		}
		// deposit records *must* exist for this epoch
		k.SetEpochTracker(ctx, epochTracker)
		k.Logger(ctx).Info("Triggering deposits")
		// Create a new deposit record for each host zone for the upcoming epoch
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)

		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
		depositInterval := int64(k.GetParam(ctx, types.KeyDepositInterval))
		if epochNumber%depositInterval == 0 {
			// process previous deposit records
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}

		// NOTE: the stake ICA timeout *must* be l.t. the staking epoch length, otherwise
		// we could send a stake ICA call (which could succeed), without deleting the record.
		// This could happen if the ack doesn't return by the next epoch. We would then send
		// *another* stake ICA call, for a portion of the balance which has *already* been staked,
		// which is very bad! This could result in the protocol becoming insolvent, by staking balances
		// that were earmarked for another purpose, e.g. redemptions.
		// The same holds true for IBC transfers.
		// Given these assumptions, the order of staking / transfers is not important, because stride deposit
		// records always accurately reflect the state of the controller / host chain by the next epoch.
		// Put another way, all outstanding ICA calls / IBC transfers must be settled on the controller
		// chain before the next epoch begins.
		delegationInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegationInterval == 0 {
			k.StakeExistingDepositsOnHostZones(ctx, epochNumber, depositRecords)
		}

		// TODO(TEST-88): Close this ticket
		reinvestInterval := int64(k.GetParam(ctx, types.KeyReinvestInterval))
		if epochNumber%reinvestInterval == 0 {
			k.ProcessRewardsInterval(ctx)
		}
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch end %s %d", epochIdentifier, epochNumber))

}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

// -------------------- helper functions --------------------
func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber int64) {
	// Create one new deposit record / host zone for the next epoch
	createDepositRecords := func(index int64, zoneInfo types.HostZone) (stop bool) {
		// create a deposit record / host zone
		depositRecord := types.NewDepositRecord(0, zoneInfo.HostDenom, zoneInfo.ChainId, types.DepositRecord_TRANSFER, uint64(epochNumber))
		k.AppendDepositRecord(ctx, *depositRecord)
		return false
	}
	// Iterate the zones and apply icaReinvest
	k.IterateHostZones(ctx, createDepositRecords)
}

func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber int64, depositRecords []types.DepositRecord) {
	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record types.DepositRecord) (condition bool) {
		return record.Status == types.DepositRecord_STAKE
	})
	for _, depositRecord := range stakeDepositRecords {
		if depositRecord.EpochNumber < uint64(epochNumber) {
			pstr := fmt.Sprintf("\tProcessing deposit {%d} {%s} {%d} {%s}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)
			hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
			if !hostZoneFound {
				k.Logger(ctx).Error("Host zone not found for deposit record {%d}", depositRecord.Id)
				continue
			}
			delegateAccount := hostZone.GetDelegationAccount()
			if delegateAccount == nil || delegateAccount.Address == "" {
				k.Logger(ctx).Error("Zone %s is missing a delegation address!", hostZone.ChainId)
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation staking on %s", hostZone.HostDenom))
			processAmount := utils.Int64ToCoinString(depositRecord.Amount, hostZone.HostDenom)
			amt, err := sdk.ParseCoinNormalized(processAmount)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Could not process coin %s: %s", hostZone.HostDenom, err))
				return
			}
			err = k.DelegateOnHost(ctx, hostZone, amt)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s", processAmount, hostZone.ChainId))
				return
			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Successfully staked %s on %s", processAmount, hostZone.ChainId))
			}

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute("hostZone", hostZone.ChainId),
					sdk.NewAttribute("newAmountStaked", strconv.FormatInt(depositRecord.Amount, 10)),
				),
			})
		}
	}
}

func (k Keeper) ProcessRewardsInterval(ctx sdk.Context) {
	icaReinvest := func(index int64, zoneInfo types.HostZone) (stop bool) {
		// Verify the delegation and withdrawal accounts are registered
		delegationIca := zoneInfo.GetDelegationAccount()
		if delegationIca == nil || delegationIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
			return true
		}
		withdrawIca := zoneInfo.GetWithdrawalAccount()
		if withdrawIca == nil || withdrawIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a withdrawal address!", zoneInfo.ChainId)
			return true
		}
		err := k.ReinvestRewards(ctx, zoneInfo)
		if err != nil {
			k.Logger(ctx).Error("Did not withdraw rewards on %s", zoneInfo.ChainId)
			return true
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Successfully withdrew rewards on %s", zoneInfo.ChainId))
		}
		return false
	}

	// Iterate the zones and apply icaReinvest
	k.IterateHostZones(ctx, icaReinvest)
}

func (k Keeper) ReinvestRewards(ctx sdk.Context, hostZone types.HostZone) error {
	// 1) [X] Query for rewards on the withdrawal account (entire outstanding balance)
	// 2) [X] Transfer rewards (entire outstanding balance) to delegation account from withdrawal account
	// 		[] transfer timeout must be l.t. epoch length
	// 		[] In the ack, create a deposit record for the rewards transferred out

	// Assumptions
	// - By the next epoch, any outstanding balance must be accounted for on stride (query timeout must be scoped to
	// 		a block height OR must timeout before the next epoch)

	// the relevant ICA is the delegate account
	owner := types.FormatICAAccountOwner(hostZone.ChainId, types.ICAAccountType_DELEGATION)
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "%s has no associated portId", owner)
	}
	connectionId, err := k.GetConnectionId(ctx, portID)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidChainID, "%s has no associated connection", portID)
	}

	// Fetch the relevant ICA
	delegationAccount := hostZone.GetDelegationAccount()
	withdrawAccount := hostZone.GetWithdrawalAccount()

	var cb icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
		var msgs []sdk.Msg
		queryRes := bankTypes.QueryAllBalancesResponse{}
		err := k.cdc.Unmarshal(args, &queryRes)
		if err != nil {
			k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
			return err
		}
		// Get denom dynamically
		balance := queryRes.Balances.AmountOf(hostZone.HostDenom)
		balanceDec := sdk.NewDec(balance.Int64())
		commission := sdk.NewDec(int64(k.GetParam(ctx, types.KeyStrideCommission))).Quo(sdk.NewDec(100))
		// Dec type has 18 decimals and the same precision as Coin types
		strideAmount := balanceDec.Mul(commission)
		reinvestAmount := balanceDec.Sub(strideAmount)
		strideCoin := sdk.NewCoin(hostZone.HostDenom, strideAmount.TruncateInt())
		reinvestCoin := sdk.NewCoin(hostZone.HostDenom, reinvestAmount.TruncateInt())

		// transfer balances from the withdraw address to the delegation account
		sendBalanceToDelegationAccount := &bankTypes.MsgSend{FromAddress: withdrawAccount.GetAddress(), ToAddress: delegationAccount.GetAddress(), Amount: sdk.NewCoins(reinvestCoin)}
		msgs = append(msgs, sendBalanceToDelegationAccount)
		// TODO: [TEST-115] get the stride commission addresses (potentially split this up into multiple messages)
		strideCommmissionAccount := "cosmos12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
		sendBalanceToStrideAccount := &bankTypes.MsgSend{FromAddress: withdrawAccount.GetAddress(), ToAddress: strideCommmissionAccount, Amount: sdk.NewCoins(strideCoin)}
		msgs = append(msgs, sendBalanceToStrideAccount)

		// Send the transaction through SubmitTx
		err = k.SubmitTxs(ctx, connectionId, msgs, *withdrawAccount)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
		}

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("WithdrawalAccountBalance", balance.String()),
				sdk.NewAttribute("ReinvestedPortion", reinvestCoin.String()),
				sdk.NewAttribute("StrideCommission", strideCoin.String()),
			),
		})

		return nil
	}
	k.InterchainQueryKeeper.QueryBalances(ctx, hostZone, cb, withdrawAccount.Address)
	return nil
}

func (k Keeper) TransferExistingDepositsToHostZones(ctx sdk.Context, epochNumber int64, depositRecords []types.DepositRecord) {
	transferDepositRecords := utils.FilterDepositRecords(depositRecords, func(record types.DepositRecord) (condition bool) {
		return record.Status == types.DepositRecord_TRANSFER
	})
	addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress().String()
	for _, depositRecord := range transferDepositRecords {
		if depositRecord.EpochNumber < uint64(epochNumber) {
			pstr := fmt.Sprintf("\tProcessing deposit {%d} {%s} {%d} {%s}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)
			hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
			if !hostZoneFound {
				k.Logger(ctx).Error("Host zone not found for deposit record {%d}", depositRecord.Id)
				continue
			}
			delegateAccount := hostZone.GetDelegationAccount()
			if delegateAccount == nil || delegateAccount.Address == "" {
				k.Logger(ctx).Error("Zone %s is missing a delegation address!", hostZone.ChainId)
				continue
			}
			delegateAddress := delegateAccount.Address
			// TODO(TEST-89): Set NewHeight relative to the most recent known gaia height (based on the LC)
			// TODO(TEST-90): why do we have two gaia LCs?
			timeoutHeight := clienttypes.NewHeight(0, 1000000)
			transferCoin := sdk.NewCoin(hostZone.GetIBCDenom(), sdk.NewInt(int64(depositRecord.Amount)))
			goCtx := sdk.WrapSDKContext(ctx)

			msg := ibctypes.NewMsgTransfer("transfer", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight, 0)
			_, err := k.transferKeeper.Transfer(goCtx, msg)
			if err != nil {
				pstr := fmt.Sprintf("\tERROR WITH DEPOSIT RECEIPT {%d}", depositRecord.Id)
				k.Logger(ctx).Info(pstr)
				return
			}
		}
	}
}
