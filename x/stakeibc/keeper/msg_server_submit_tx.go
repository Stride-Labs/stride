package keeper

import (
	"context"

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

	var cb icqkeeper.Callback = func(k icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
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
	k.InterchainQueryKeeper.QueryBalances(ctx, hostZone, cb, withdrawAccount.Address)
	return nil
}

// icq to read host delegated balance => update hostZone.delegationAccount.DelegatedBalance
// TODO(TEST-97) add safety logic to query at specific block height (same as query height for delegated balances)
func (k Keeper) UpdateDelegatedBalance(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx
	// Fetch the relevant ICA
	delegationAccount := hostZone.GetDelegationAccount()

	var cb icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {

		var response stakingTypes.QueryDelegatorDelegationsResponse
		err := k.cdc.Unmarshal(args, &response)
		if err != nil {
			return err
		}

		// Get denom dynamically
		hz := hostZone
		delegatorSum := sdk.NewCoin(hz.HostDenom, sdk.ZeroInt())
		for _, delegation := range response.DelegationResponses {
			delegatorSum = delegatorSum.Add(delegation.Balance)
			if err != nil {
				return err
			}
		}

		// Set delegation account balance to ICQ result
		da := hz.DelegationAccount
		da.Balance = delegatorSum.Amount.Int64()
		hz.DelegationAccount = da
		Keeper.SetHostZone(k, ctx, hz)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("totalDelegatedBalance", delegatorSum.Amount.String()),
			),
		})

		return nil
	}
	// 1. query delegation account delegated balances using icq
	// 2. sum up the resulting delegations (across validators)
	// 2. write the result to hostZone.delegationAccount.delegatedBalance
	k.InterchainQueryKeeper.QueryDelegatorDelegations(ctx, hostZone, cb, delegationAccount.Address)
	return nil
}

// icq to read host delegation account undeleted balance => update hostZone.delegationAccount.Balance
// TODO(TEST-97) add safety logic to query at specific block height (same as query height for delegated balances)
func (k Keeper) UpdateUndelegatedBalance(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx
	// Fetch the relevant ICA
	delegationAccount := hostZone.GetDelegationAccount()

	var cb icqkeeper.Callback = func(icqk icqkeeper.Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {

		queryRes := bankTypes.QueryAllBalancesResponse{}
		err := k.cdc.Unmarshal(args, &queryRes)
		if err != nil {
			k.Logger(ctx).Error("Unable to unmarshal balances info for zone", "err", err)
			return err
		}

		// Get denom dynamically
		hz := hostZone
		balance := queryRes.Balances.AmountOf(hz.HostDenom)

		// Set delegation account balance to ICQ result
		hz, found := k.GetHostZone(ctx, hostZone.ChainId)
		if found {
			k.Logger(ctx).Error("invalid chain id, zone for \"%s\" already registered", hostZone.ChainId)
		}

		da := hz.DelegationAccount
		da.Balance = balance.Int64()
		hz.DelegationAccount = da
		k.SetHostZone(ctx, hz)

		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("totalUndelegatedBalance", balance.String()),
			),
		})

		return nil
	}
	// 1. query delegation account undelegated balances using icq
	// 2. write the result to hostZone.delegationAccount.Balance
	k.InterchainQueryKeeper.QueryBalances(ctx, hostZone, cb, delegationAccount.Address)
	return nil
}

// Update the redemption rate using values of delegatedBalances, balances and stAsset supply
// TODO(TEST-97) add safety logic that checks balance, delegatedBalance and stAsset supply's block_height_updated are all equal
func (k Keeper) UpdateExchangeRate(ctx sdk.Context, hostZone types.HostZone) error {
	_ = ctx

	// Assets: native asset balances on delegation account + staked
	delegatedBalance := hostZone.GetDelegationAccount().DelegatedBalance
	unDelegatedBalance := hostZone.GetDelegationAccount().Balance
	assetBalance := delegatedBalance + unDelegatedBalance

	// Claims: stAsset supply
	stAssetSupply := k.bankKeeper.GetSupply(ctx, hostZone.IBCDenom)

	// ExchRate = Assets / Claims
	redemptionRate := sdk.NewDec(assetBalance).Quo(stAssetSupply.Amount.ToDec())

	// Write ExchRate to state
	hostZone.LastRedemptionRate = hostZone.RedemptionRate
	hostZone.RedemptionRate = redemptionRate
	k.SetHostZone(ctx, hostZone)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("lastRedemptionRate", redemptionRate.String()),
			sdk.NewAttribute("newRedemptionRate", hostZone.LastRedemptionRate.String()),
		),
	})

	return nil
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
