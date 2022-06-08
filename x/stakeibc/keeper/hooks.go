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
			// TODO move this function to the keeper
			k.Logger(ctx).Info("Triggering deposits")
			depositRecords := k.GetAllDepositRecord(ctx)
			for _, depositRecord := range depositRecords {
				pstr := fmt.Sprintf("\tProcessing deposit {%d} {%s} {%d} {%s}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount, depositRecord.Sender)
				k.Logger(ctx).Info(pstr)
				addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress().String()
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
				timeoutHeight := clienttypes.NewHeight(0, 500)
				transferCoin := sdk.NewCoin(depositRecord.Denom, sdk.NewInt(int64(depositRecord.Amount)))
				goCtx := sdk.WrapSDKContext(ctx)
				msg := ibctypes.NewMsgTransfer("transfer", hostZone.ChainId, transferCoin, addr, delegateAddress, timeoutHeight, 0)
				_, err := k.transferKeeper.Transfer(goCtx, msg)
				if err != nil {
					pstr := fmt.Sprintf("\tERROR WITH DEPOSIT RECEIPT {%d}", depositRecord.Id)
					k.Logger(ctx).Info(pstr)
					panic(err)
				} else {
					// TODO what should we do if this transfer fails
					k.RemoveDepositRecord(ctx, depositRecord.Id)
				}
			}
		}
		delegateInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegateInterval == 0 {
			icaStake := func(index int64, zoneInfo types.HostZone) (stop bool) {
				// Verify the delegation ICA is registered
				delegationIca := zoneInfo.GetDelegationAccount()
				if delegationIca == nil || delegationIca.Address == "" {
					k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
					return true
				}

				// TODO(TEST-46): Query process amount (unstaked balance) on host zone using ICQ
				processAmount := "1" + zoneInfo.BaseDenom
				amt, err := sdk.ParseCoinNormalized(processAmount)
				// Do we want to panic here? All unprocessed zones would also fail
				if err != nil {
					panic(err)
				}
				err = k.DelegateOnHost(ctx, zoneInfo, amt)
				if err != nil {
					k.Logger(ctx).Error("Did not stake %s on %s", processAmount, zoneInfo.ChainId)
					return true
				} else {
					k.Logger(ctx).Info("Successfully staked %s on %s", processAmount, zoneInfo.ChainId)
				}
				return false
			}

			// Iterate the zones and apply icaStake
			k.IterateHostZones(ctx, icaStake)
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
