package keeper

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	icqkeeper "github.com/Stride-Labs/stride/x/interchainquery/keeper"
	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
)

// SubmitTx sends an ICA transaction to a host chain on behalf of an account on the controller
// chain.
// NOTE: this is not a standard message; only the stakeibc module should call this function. However,
// this is temporarily in the message server to facilitate easy testing and development.
// TODO(TEST-53): Remove this pre-launch (no need for clients to create / interact with ICAs)
func (k Keeper) SubmitTx(goCtx context.Context, msg *types.MsgSubmitTx) (*types.MsgSubmitTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	portID, err := icatypes.NewControllerPortID(msg.Owner)
	if err != nil {
		return nil, err
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		return nil, sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, []sdk.Msg{msg.GetTxMsg()})
	if err != nil {
		return nil, err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	// timeoutTimestamp := time.Now().Add(time.Minute).UnixNano()
	timeoutTimestamp := ^uint64(0) >> 1
	_, err = k.ICAControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTxResponse{}, nil
}

func (k Keeper) DelegateOnHost(ctx sdk.Context, hostZone types.HostZone, amt sdk.Coin) error {
	_ = ctx
	var msgs []sdk.Msg
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
	delegationIca := hostZone.GetDelegationAccount()

	// Construct the transaction
	// TODO(TEST-39): Implement validator selection
	validator_address := "cosmosvaloper19e7sugzt8zaamk2wyydzgmg9n3ysylg6na6k6e" // gval2

	// construct the msg
	msgs = append(msgs, &stakingTypes.MsgDelegate{DelegatorAddress: delegationIca.GetAddress(), ValidatorAddress: validator_address, Amount: amt})
	// Send the transaction through SubmitTx
	err = k.SubmitTxs(ctx, connectionId, msgs, *delegationIca)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s", connectionId, hostZone.ChainId, msgs)
	}
	return nil
}

func (k Keeper) ReinvestRewards(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx
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

	SubmitTxs := k.SubmitTxs
	GetParam := k.GetParam
	cdc := k.cdc

	var cb icqkeeper.Callback = func(k icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query, h int64) error {
		var msgs []sdk.Msg
		queryRes := bankTypes.QueryAllBalancesResponse{}
		err := cdc.Unmarshal(args, &queryRes)
		if err != nil {
			k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
			return err
		}
		// Get denom dynamically
		balance := queryRes.Balances.AmountOf(hostZone.HostDenom)
		balanceDec := sdk.NewDec(balance.Int64())
		commission := sdk.NewDec(int64(GetParam(ctx, types.KeyStrideCommission))).Quo(sdk.NewDec(100))
		// Dec type has 18 decimals and the same precision as Coin types
		strideAmount := balanceDec.Mul(commission)
		reinvestAmount := balanceDec.Sub(strideAmount)
		strideCoin := sdk.NewCoin(hostZone.HostDenom, strideAmount.TruncateInt())
		reinvestCoin := sdk.NewCoin(hostZone.HostDenom, reinvestAmount.TruncateInt())

		// transfer balances from the withdraw address to the delegation account
		sendBalanceToDelegationAccount := &bankTypes.MsgSend{FromAddress: withdrawAccount.GetAddress(), ToAddress: delegationAccount.GetAddress(), Amount: sdk.NewCoins(reinvestCoin)}
		msgs = append(msgs, sendBalanceToDelegationAccount)
		// TODO: get the stride commission addresses (potentially split this up into multiple messages)
		strideCommmissionAccount := "cosmos12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
		sendBalanceToStrideAccount := &bankTypes.MsgSend{FromAddress: withdrawAccount.GetAddress(), ToAddress: strideCommmissionAccount, Amount: sdk.NewCoins(strideCoin)}
		msgs = append(msgs, sendBalanceToStrideAccount)

		// Send the transaction through SubmitTx
		err = SubmitTxs(ctx, connectionId, msgs, *withdrawAccount)
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
	// 1. query withdraw account balances using icq
	// 2. transfer withdraw account balances to the delegation account in the cb
	// 3. TODO: in the ICA ack upon transfer, reinvest those rewards and withdraw rewards
	k.InterchainQueryKeeper.QueryBalances(ctx, hostZone, cb, withdrawAccount.Address, 0)
	return nil
}

// icq to read host delegation account undeleted balance => update hostZone.delegationAccount.Balance
// TODO(TEST-97) add safety logic to query at specific block height (same as query height for delegated balances)
func (k Keeper) ProcessUpdateUndelegatedBalance(ctx sdk.Context, height int64) {
	updateUndelegatedBal := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		// Verify the delegation ICA is registered
		k.Logger(ctx).Info(fmt.Sprintf("\tProcessing delegation %s", zoneInfo.ChainId))
		delegationIca := zoneInfo.GetDelegationAccount()
		if delegationIca == nil || delegationIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
			return errors.New("Zone is missing a delegation address!")
		}
		cdc := k.cdc

		var queryUndelegatedBalanceCB icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query, h int64) error {
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation undelegatedbalance callback on %s", zoneInfo.HostDenom))
			queryRes := bankTypes.QueryAllBalancesResponse{}
			err := cdc.Unmarshal(args, &queryRes)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
				return err
			}
			// Get denom dynamically
			balance := queryRes.Balances.AmountOf(zoneInfo.HostDenom)
			k.Logger(ctx).Info(fmt.Sprintf("\tBalance on %s is %s", zoneInfo.HostDenom, balance.String()))

			// --- Update Undelegated Balance ---
			hz := zoneInfo

			da := hz.DelegationAccount
			da.Balance = balance.Int64()
			// only update height if this is the most updated query (ICQ msgresponses are not always FIFO)
			if h >= da.HeightLastQueriedUndelegatedBalance {
				da.HeightLastQueriedUndelegatedBalance = h
				hz.DelegationAccount = da
				k.SetHostZone(ctx, hz)
				k.Logger(ctx).Info(fmt.Sprintf("Just set UndelegatedBalance to: %d", da.DelegatedBalance))
				k.Logger(ctx).Info(fmt.Sprintf("Just set HeightLastQueriedUndelegatedBalance to: %d", h))
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute("hostZone", zoneInfo.ChainId),
						sdk.NewAttribute("totalUndelegatedBalance", balance.String()),
					),
				})
				// Calc redemption rate
				// 1. check equality of latest UB and DB update heights
				if hz.DelegationAccount.HeightLastQueriedDelegatedBalance == da.HeightLastQueriedUndelegatedBalance {
					// 2. check to make sure we have a corresponding ControllerBalance
					cb, found := k.GetControllerBalances(ctx, strconv.FormatInt(hz.DelegationAccount.HeightLastQueriedDelegatedBalance, 10))
					if found {
						// 2.5 abort if stSupply is 0 at this host height
						if cb.Stsupply > 0 {
							redemptionRate := (sdk.NewDec(da.HeightLastQueriedUndelegatedBalance).Add(sdk.NewDec(hz.DelegationAccount.DelegatedBalance)).Add(sdk.NewDec(cb.Moduleacctbalance))).Quo(sdk.NewDec(cb.Stsupply))
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

			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Opted to NOT set HeightLastQueriedUndelegatedBalance because query height %d is less than last update's height %d.", h, da.HeightLastQueriedUndelegatedBalance))
			}
			// ---------------------------------

			return nil
		}
		k.Logger(ctx).Info(fmt.Sprintf("\tQuerying UndelegatedBalance for %s at %d height", zoneInfo.ChainId, height))
		k.InterchainQueryKeeper.QueryBalances(ctx, zoneInfo, queryUndelegatedBalanceCB, delegationIca.Address, height)
		return nil
	}

	// Iterate the zones and apply updateUndelegatedBal
	k.IterateHostZones(ctx, updateUndelegatedBal)
}

func (k Keeper) ProcessUpdateBalances(ctx sdk.Context, height int64) {
	// TODO(now) ; rm this after debug; should never call at 0!
	if height == 0 {
		return
	}
	updateBalances := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("\tUpdating balances on %s", zoneInfo.ChainId))

		delegationIca := zoneInfo.GetDelegationAccount()
		if delegationIca == nil || delegationIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
			return errors.New("Zone is missing a delegation address!")
		}
		var queryDBCBDaisyChain icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query, h int64) error {
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation DelegatedBalance callback on %s", zoneInfo.HostDenom))
			var response stakingTypes.QueryDelegatorDelegationsResponse
			err := k.cdc.Unmarshal(args, &response)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
				return err
			}

			// Add up delegations across validators
			delegatorSum := sdk.NewCoin(zoneInfo.HostDenom, sdk.ZeroInt())
			for _, delegation := range response.DelegationResponses {
				delegatorSum = delegatorSum.Add(delegation.Balance)
				if err != nil {
					return err
				}
			}
			// --- Update Undelegated Balance ---
			hz, _ := k.GetHostZone(ctx, zoneInfo.ChainId)
			da := hz.DelegationAccount
			// TODO() make HostZone.DelegationAccount.Balance a Cosmos.Dec type (rather than int)
			// only update height if this is the most updated query (ICQ msgresponses are not always FIFO)
			if h >= da.HeightLastQueriedDelegatedBalance {
				da.HeightLastQueriedDelegatedBalance = h
				da.DelegatedBalance = delegatorSum.Amount.Int64()
				hz.DelegationAccount = da
				k.SetHostZone(ctx, hz)
				k.Logger(ctx).Info(fmt.Sprintf("Just set DelegatedBalance to: %d", da.DelegatedBalance))

				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						sdk.EventTypeMessage,
						sdk.NewAttribute("hostZone", zoneInfo.ChainId),
						sdk.NewAttribute("totalDelegatedBalance", delegatorSum.Amount.String()),
					),
				})

				// Daisy chain!
				k.Logger(ctx).Info(fmt.Sprintf("\tSecond step in our daisy chain update! Now, querying undelegatedBalanaces for %s at %d height", zoneInfo.ChainId, h))
				k.ProcessUpdateUndelegatedBalance(ctx, h)
				return nil
			} else {
				// height is below last updated height, so stop here!
				k.Logger(ctx).Info(fmt.Sprintf("Opted to NOT set HeightLastQueriedDelegatedBalance; query height %d is less than last update's height %d.", h, da.HeightLastQueriedDelegatedBalance))
				return nil
			}
		}
		k.Logger(ctx).Info(fmt.Sprintf("\tStarting our daisy chain update! First, querying delegatedBalanaces for %s at %d height", zoneInfo.ChainId, height))
		k.InterchainQueryKeeper.QueryDelegatorDelegations(ctx, zoneInfo, queryDBCBDaisyChain, delegationIca.Address, height)
		return nil
	}

	// Iterate the zones and apply updateDelegatedBal
	k.IterateHostZones(ctx, updateBalances)
}

// icq to read host delegated balance => update hostZone.delegationAccount.DelegatedBalance
// TODO(TEST-97) add safety logic to query at specific block height (same as query height for delegated balances)
func (k Keeper) ProcessUpdateDelegatedBalance(ctx sdk.Context, height int64) {
	updateDelegatedBal := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		// Verify the delegation ICA is registered
		k.Logger(ctx).Info(fmt.Sprintf("\tUpdating delegated balances on %s", zoneInfo.ChainId))
		delegationIca := zoneInfo.GetDelegationAccount()
		if delegationIca == nil || delegationIca.Address == "" {
			k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
			return errors.New("Zone is missing a delegation address!")
		}
		cdc := k.cdc

		var queryDelegatedBalanceCB icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query, h int64) error {
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation DelegatedBalance callback on %s", zoneInfo.HostDenom))
			var response stakingTypes.QueryDelegatorDelegationsResponse
			err := cdc.Unmarshal(args, &response)
			if err != nil {
				k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
				return err
			}

			// Add up delegations across validators
			delegatorSum := sdk.NewCoin(zoneInfo.HostDenom, sdk.ZeroInt())
			for _, delegation := range response.DelegationResponses {
				delegatorSum = delegatorSum.Add(delegation.Balance)
				if err != nil {
					return err
				}
			}
			// --- Update Undelegated Balance ---
			hz, _ := k.GetHostZone(ctx, zoneInfo.ChainId)
			da := hz.DelegationAccount
			// TODO() make HostZone.DelegationAccount.Balance a Cosmos.Dec type (rather than int)
			da.HeightLastQueriedDelegatedBalance = h
			da.DelegatedBalance = delegatorSum.Amount.Int64()
			hz.DelegationAccount = da
			k.SetHostZone(ctx, hz)
			k.Logger(ctx).Info(fmt.Sprintf("Just set DelegatedBalance to: %d", da.DelegatedBalance))
			k.Logger(ctx).Info(fmt.Sprintf("Just set HeightLastQueriedDelegatedBalance to: %d", h))

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute("hostZone", zoneInfo.ChainId),
					sdk.NewAttribute("totalDelegatedBalance", delegatorSum.Amount.String()),
				),
			})
			// ---------------------------------

			return nil
		}
		k.Logger(ctx).Info(fmt.Sprintf("\tQuerying delegatedBalance for %s at %d height", zoneInfo.ChainId, height))
		k.InterchainQueryKeeper.QueryDelegatorDelegations(ctx, zoneInfo, queryDelegatedBalanceCB, delegationIca.Address, height)
		return nil
	}

	// Iterate the zones and apply updateDelegatedBal
	k.IterateHostZones(ctx, updateDelegatedBal)
}

// Update the redemption rate using values of delegatedBalances, balances and stAsset supply
// TODO(TEST-97) add safety logic that checks balance, delegatedBalance and stAsset supply's block_height_updated are all equal
func (k Keeper) ProcessUpdateExchangeRate(ctx sdk.Context) {

	// updateRedemptionRate := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
	// 	k.Logger(ctx).Info(fmt.Sprintf("\tProcessing exchangeRate for %s", zoneInfo.ChainId))
	// 	// Assets: native asset balances on delegation account + staked
	// 	da := zoneInfo.GetDelegationAccount()
	// 	if da.HeightLastQueriedDelegatedBalance != da.HeightLastQueriedUndelegatedBalance {
	// 		k.Logger(ctx).Info(fmt.Sprintf("\tHost Zone %s's balance update heights don't match! Skipping exchange rate update at Stride block %s | DB updated at H=%s, UB updated at H=%s.", zoneInfo.ChainId, ctx.BlockHeight(), strconv.Itoa(da.HeightLastQueriedDelegatedBalance), strconv.Itoa(da.HeightLastQueriedUndelegatedBalance)))
	// 		return errors.New("Exchange rate calculation error: block heights don't match.")
	// 		assetBalance := da.Balance + da.DelegatedBalance

	// 	}

	// 	// Claims: stAsset supply
	// 	// TODO(NOW) move stAsset query to epochs!
	// 	stAssetSupply := k.bankKeeper.GetSupply(ctx, "st"+zoneInfo.HostDenom)

	// 	// Sanity & safety check: if either num or denom are 0, do NOT update the exchange rate
	// 	if assetBalance == int64(0) || stAssetSupply.Amount.Int64() == int64(0) {
	// 		ctx.EventManager().EmitEvents(sdk.Events{
	// 			sdk.NewEvent(
	// 				sdk.EventTypeMessage,
	// 				sdk.NewAttribute("delegatedBalance", strconv.FormatInt(delegatedBalance, 10)),
	// 				sdk.NewAttribute("unDelegatedBalance", strconv.FormatInt(unDelegatedBalance, 10)),
	// 				sdk.NewAttribute("stAssetBalance", stAssetSupply.String()),
	// 			),
	// 		})
	// 		return errors.New("Exchange rate calculation error: ")
	// 	} else {
	// 		// RedemptionRate = Assets / Claims
	// 		redemptionRate := sdk.NewDec(assetBalance).Quo(stAssetSupply.Amount.ToDec())
	// 		// Write RedemptionRate to state
	// 		hz, _ := k.GetHostZone(ctx, zoneInfo.ChainId)
	// 		hz.LastRedemptionRate = hz.RedemptionRate
	// 		hz.RedemptionRate = redemptionRate
	// 		k.SetHostZone(ctx, hz)

	// 		ctx.EventManager().EmitEvents(sdk.Events{
	// 			sdk.NewEvent(
	// 				sdk.EventTypeMessage,
	// 				sdk.NewAttribute("lastRedemptionRate", redemptionRate.String()),
	// 				sdk.NewAttribute("newRedemptionRate", hz.LastRedemptionRate.String()),
	// 			),
	// 		})
	// 	}
	// 	return nil
	// }

	// k.IterateHostZones(ctx, updateRedemptionRate)
}

// SubmitTxs submits an ICA transaction containing multiple messages
func (k Keeper) SubmitTxs(ctx sdk.Context, connectionId string, msgs []sdk.Msg, account types.ICAAccount) error {
	chainId, err := k.GetChainID(ctx, connectionId)
	if err != nil {
		return err
	}
	owner := types.FormatICAAccountOwner(chainId, account.GetTarget())
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelID, found := k.ICAControllerKeeper.GetActiveChannelID(ctx, connectionId, portID)
	if !found {
		return sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, msgs)
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	// TODO(TEST-37): Decide on timeout logic
	timeoutTimestamp := ^uint64(0) >> 1
	_, err = k.ICAControllerKeeper.SendTx(ctx, chanCap, connectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return err
	}

	return nil
}
