package keeper

import (
	"fmt"
	"strconv"

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
				timeoutHeight := clienttypes.NewHeight(0, 1000000000000)
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

		// DELEGATE FROM DELEGATION ACCOUNT
		delegateInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegateInterval == 0 {
			// get Gaia LC height
			k.ProcessDelegationStaking(ctx)
		}

		exchangeRateInterval := int64(k.GetParam(ctx, types.KeyExchangeRateInterval))
		if epochNumber%exchangeRateInterval == 0 && (epochNumber > 100) { // allow a few blocks from UpdateUndelegatedBal to avoid conflicts
			// TODO(NOW) parameterize connection by hostZone
			// TODO(NOW) update LC before getting latest height
			connectionID := "connection-0"
			latestHeightGaia, found := k.GetLightClientHeightSafely(ctx, connectionID)
			if !found {
				k.Logger(ctx).Error("client id not found for connection \"%s\"", connectionID)
			} else {
				// TODO(119) generalize to host_zones
				// SET STASSETSUPPLY
				hz, _ := k.GetHostZone(ctx, "GAIA")
				//TODO(TEST-119) replace below with StAssetDenomFromHostZoneDenom() at merge
				currStSupply := k.bankKeeper.GetSupply(ctx, "st"+hz.HostDenom)
				// GET MODULE ACCT BALANCE
				addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
				modAcctBal := k.bankKeeper.GetBalance(ctx, addr, hz.IBCDenom)

				ControllerBalancesRecord := types.ControllerBalances{
					Index:             strconv.FormatInt(latestHeightGaia, 10),
					Height:            latestHeightGaia,
					Stsupply:          currStSupply.Amount.Int64(),
					Moduleacctbalance: modAcctBal.Amount.Int64(),
				}
				k.SetControllerBalances(ctx, ControllerBalancesRecord)
				k.Logger(ctx).Info(fmt.Sprintf("Set ControllerBalances at H=%d to stSupply=%d, moduleAcctBalances=%d", latestHeightGaia, currStSupply.Amount.Int64(), modAcctBal.Amount.Int64()))

				// TODO(TEST-97) update only when balances, delegatedBalances and stAsset supply are results from the same block
				k.UpdateRedemptionRatePart1(ctx, latestHeightGaia)
			}
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
