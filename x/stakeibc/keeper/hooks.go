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
			// GET LATEST HEIGHT
			// TODO(NOW) wrap this into a function
			var latestHeightGaia int64 // defaults to 0
			// get light client's latest height
			connectionID := "connection-0"
			conn, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("invalid connection id, \"%s\" not found", connectionID))
			}
			//TODO(TEST-112) make sure to update host LCs here!
			clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("client id \"%s\" not found for connection \"%s\"", conn.ClientId, connectionID))
				// latestHeightGaia = 0
			} else {
				// TODO(TEST-119) get stAsset supply at SAME time as gaia height
				// TODO(TEST-112) check on safety of castng uint64 to int64
				latestHeightGaia = int64(clientState.GetLatestHeight().GetRevisionHeight())

				// TODO(119) generalize to host_zones
				// SET STASSETSUPPLY
				hz, _ := k.GetHostZone(ctx, "GAIA")
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
				k.ProcessUpdateBalances(ctx, latestHeightGaia)
			}

			// Calc redemption rate
			// 1. check equality of latest UB and DB update heights
			hz, _ := k.GetHostZone(ctx, "GAIA")
			if hz.DelegationAccount.HeightLastQueriedDelegatedBalance == hz.DelegationAccount.HeightLastQueriedUndelegatedBalance {
				// 2. check to make sure we have a corresponding ControllerBalance
				cb, found := k.GetControllerBalances(ctx, strconv.FormatInt(hz.DelegationAccount.HeightLastQueriedDelegatedBalance, 10))
				if found {
					// 2.5 abort if stSupply is 0 at this host height
					if cb.Stsupply > 0 {
						redemptionRate := (sdk.NewDec(hz.DelegationAccount.Balance).Add(sdk.NewDec(hz.DelegationAccount.DelegatedBalance)).Add(sdk.NewDec(cb.Moduleacctbalance))).Quo(sdk.NewDec(cb.Stsupply))
						hz.LastRedemptionRate = hz.RedemptionRate
						hz.RedemptionRate = redemptionRate
						k.SetHostZone(ctx, hz)
						k.Logger(ctx).Info(fmt.Sprintf("Set Redemptions Rate at H=%d to RR=%d", hz.DelegationAccount.HeightLastQueriedDelegatedBalance, redemptionRate))
					} else {
						k.Logger(ctx).Info(fmt.Sprintf("Did NOT set redemption rate at H=%d because stAsset supply was 0", hz.DelegationAccount.HeightLastQueriedDelegatedBalance))
					}
				} else {
					k.Logger(ctx).Info(fmt.Sprintf("Did NOT set redemption rate at H=%d because no controller balances", hz.DelegationAccount.HeightLastQueriedDelegatedBalance))
				}
			}
			k.Logger(ctx).Info(fmt.Sprintf("Did NOT set redemption rate at H=%d because last UB and DB update heights didn't match.", hz.DelegationAccount.HeightLastQueriedDelegatedBalance))
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
