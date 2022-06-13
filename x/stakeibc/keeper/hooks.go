package keeper

import (
	"fmt"

	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s %d", epochIdentifier, epochNumber))
	if epochIdentifier == "stride_epoch" {
		k.Logger(ctx).Info(fmt.Sprintf("Stride Epoch %d", epochNumber))
		depositInterval := int64(k.GetParam(ctx, types.KeyDepositInterval))
		if epochNumber%depositInterval == 0 {
			// TODO TEST-72 move this function to the keeper
			k.Logger(ctx).Info("Triggering deposits")
			depositRecords := k.GetAllDepositRecord(ctx)
			addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress().String()
			for _, depositRecord := range depositRecords {
				pstr := fmt.Sprintf("\tProcessing deposit {%d} {%s} {%d} {%s}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount, depositRecord.Sender)
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
				timeoutHeight := clienttypes.NewHeight(0, 10000)
				transferCoin := sdk.NewCoin(hostZone.GetIBCDenom(), sdk.NewInt(int64(depositRecord.Amount)))
				goCtx := sdk.WrapSDKContext(ctx)

				msg := ibctypes.NewMsgTransfer("transfer", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight, 0)
				_, err := k.transferKeeper.Transfer(goCtx, msg)
				if err != nil {
					pstr := fmt.Sprintf("\tERROR WITH DEPOSIT RECEIPT {%d}", depositRecord.Id)
					k.Logger(ctx).Info(pstr)
					panic(err)
				} else {
					// TODO TEST-71 what should we do if this transfer fails
					k.RemoveDepositRecord(ctx, depositRecord.Id)
				}
			}
		}
		delegateInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegateInterval == 0 {
			k.ProcessDelegationStaking(ctx)
		}
		exchangeRateInterval := int64(k.GetParam(ctx, types.KeyExchangeRateInterval))
		if epochNumber%exchangeRateInterval == 0 {
			// TODO(TEST-98) Decide on sequencing / interval for exch rate updates; what are the edge cases?
			k.IterateHostZones(ctx, k.UpdateDelegatedBalance)
			k.IterateHostZones(ctx, k.UpdateUndelegatedBalance)
			// TODO(TEST-97) update only when balances, delegatedBalances and stAsset supply are results from the same block
			k.IterateHostZones(ctx, k.UpdateExchangeRate)
		}

		// process withdrawals
		// TODO(TEST-88): restructure this to be more efficient, we should only have to loop
		// over host zones once
		// reinvestInterval := int64(k.GetParam(ctx, types.KeyReinvestInterval))
		// if epochNumber%reinvestInterval == 0 {
		// 	icaReinvest := func(index int64, zoneInfo types.HostZone) (stop bool) {
		// 		// Verify the delegation ICA is registered
		// 		delegationIca := zoneInfo.GetDelegationAccount()
		// 		if delegationIca == nil || delegationIca.Address == "" {
		// 			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
		// 			return true
		// 		}
		// 		withdrawIca := zoneInfo.GetWithdrawalAccount()
		// 		if withdrawIca == nil || withdrawIca.Address == "" {
		// 			k.Logger(ctx).Error("Zone %s is missing a withdrawal address!", zoneInfo.ChainId)
		// 			return true
		// 		}
		// 		err := k.ReinvestRewards(ctx, zoneInfo)
		// 		if err != nil {
		// 			k.Logger(ctx).Error("Did not withdraw rewards on %s", zoneInfo.ChainId)
		// 			return true
		// 		} else {
		// 			k.Logger(ctx).Info(fmt.Sprintf("Successfully withdrew rewards on %s", zoneInfo.ChainId))
		// 		}
		// 		return false
		// 	}

		// 	// Iterate the zones and apply icaReinvest
		// 	k.IterateHostZones(ctx, icaReinvest)
		// }
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
