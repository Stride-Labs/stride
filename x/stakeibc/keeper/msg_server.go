package keeper

import (
	"context"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	proto "github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v27/utils"
	epochtypes "github.com/Stride-Labs/stride/v27/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v27/x/records/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RegisterHostZone(goCtx context.Context, msg *types.MsgRegisterHostZone) (*types.MsgRegisterHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.RegisterHostZone(ctx, msg)
}

func (ms msgServer) UpdateHostZoneParams(goCtx context.Context, msg *types.MsgUpdateHostZoneParams) (*types.MsgUpdateHostZoneParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	hostZone, found := ms.Keeper.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	maxMessagesPerTx := msg.MaxMessagesPerIcaTx
	if maxMessagesPerTx == 0 {
		maxMessagesPerTx = DefaultMaxMessagesPerIcaTx
	}
	hostZone.MaxMessagesPerIcaTx = maxMessagesPerTx
	ms.Keeper.SetHostZone(ctx, hostZone)

	return &types.MsgUpdateHostZoneParamsResponse{}, nil
}

func (k msgServer) AddValidators(goCtx context.Context, msg *types.MsgAddValidators) (*types.MsgAddValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, validator := range msg.Validators {
		if err := k.AddValidatorToHostZone(ctx, msg.HostZone, *validator, false); err != nil {
			return nil, err
		}

		// Query and store the validator's sharesToTokens rate
		if err := k.QueryValidatorSharesToTokensRate(ctx, msg.HostZone, validator.Address); err != nil {
			return nil, err
		}
	}

	// Confirm none of the validator's exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, msg.HostZone); err != nil {
		return nil, err
	}

	return &types.MsgAddValidatorsResponse{}, nil
}

func (k msgServer) DeleteValidator(goCtx context.Context, msg *types.MsgDeleteValidator) (*types.MsgDeleteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.RemoveValidatorFromHostZone(ctx, msg.HostZone, msg.ValAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to remove validator %s from host zone %s", msg.ValAddr, msg.HostZone)
	}

	return &types.MsgDeleteValidatorResponse{}, nil
}

func (k msgServer) ChangeValidatorWeight(goCtx context.Context, msg *types.MsgChangeValidatorWeights) (*types.MsgChangeValidatorWeightsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.HostZone)
	if !found {
		return nil, types.ErrInvalidHostZone
	}

	for _, weightChange := range msg.ValidatorWeights {

		validatorFound := false
		for _, validator := range hostZone.Validators {
			if validator.Address == weightChange.Address {
				validator.Weight = weightChange.Weight
				k.SetHostZone(ctx, hostZone)

				validatorFound = true
				break
			}
		}

		if !validatorFound {
			return nil, types.ErrValidatorNotFound
		}
	}

	// Confirm the new weights wouldn't cause any validator to exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, msg.HostZone); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to change validator weight")
	}

	return &types.MsgChangeValidatorWeightsResponse{}, nil
}

func (k msgServer) RebalanceValidators(goCtx context.Context, msg *types.MsgRebalanceValidators) (*types.MsgRebalanceValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("RebalanceValidators executing %v", msg))

	if err := k.RebalanceDelegationsForHostZone(ctx, msg.HostZone); err != nil {
		return nil, err
	}
	return &types.MsgRebalanceValidatorsResponse{}, nil
}

func (k msgServer) ClearBalance(goCtx context.Context, msg *types.MsgClearBalance) (*types.MsgClearBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrInvalidHostZone, "chainId: %s", msg.ChainId)
	}
	if zone.FeeIcaAddress == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "fee acount not found for chainId: %s", msg.ChainId)
	}

	sourcePort := ibctransfertypes.PortID
	// Should this be a param?
	// I think as long as we have a timeout on this, it should be hard to attack (even if someone send a tx on a bad channel, it would be reverted relatively quickly)
	sourceChannel := msg.Channel
	coinString := cast.ToString(msg.Amount) + zone.GetHostDenom()
	tokens, err := sdk.ParseCoinNormalized(coinString)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("failed to parse coin (%s)", coinString))
		return nil, errorsmod.Wrapf(err, "failed to parse coin (%s)", coinString)
	}
	// KeyICATimeoutNanos are for our Stride ICA calls, KeyFeeTransferTimeoutNanos is for the IBC transfer
	feeTransferTimeoutNanos := k.GetParam(ctx, types.KeyFeeTransferTimeoutNanos)
	timeoutTimestamp := cast.ToUint64(ctx.BlockTime().UnixNano()) + feeTransferTimeoutNanos
	msgs := []proto.Message{
		&ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            tokens,
			Sender:           zone.FeeIcaAddress, // fee account on the host zone
			Receiver:         types.FeeAccount,   // fee account on stride
			TimeoutTimestamp: timeoutTimestamp,
		},
	}

	connectionId := zone.GetConnectionId()

	icaTimeoutNanos := k.GetParam(ctx, types.KeyICATimeoutNanos)
	icaTimeoutNanos = cast.ToUint64(ctx.BlockTime().UnixNano()) + icaTimeoutNanos

	_, err = k.SubmitTxs(ctx, connectionId, msgs, types.ICAAccountType_FEE, icaTimeoutNanos, "", nil)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to submit txs")
	}
	return &types.MsgClearBalanceResponse{}, nil
}

// Exchanges a user's native tokens for stTokens using the current redemption rate
// The native tokens must live on Stride with an IBC denomination before this function is called
// The typical flow consists, first, of a transfer of native tokens from the host zone to Stride,
//
//	and then the invocation of this LiquidStake function
//
// WARNING: This function is invoked from the begin/end blocker in a way that does not revert partial state when
//
//	an error is thrown (i.e. the execution is non-atomic).
//	As a result, it is important that the validation steps are positioned at the top of the function,
//	and logic that creates state changes (e.g. bank sends, mint) appear towards the end of the function
func (k msgServer) LiquidStake(goCtx context.Context, msg *types.MsgLiquidStake) (*types.MsgLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the host zone from the base denom in the message (e.g. uatom)
	hostZone, err := k.GetHostZoneFromHostDenom(ctx, msg.HostDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "no host zone found for denom (%s)", msg.HostDenom)
	}

	// Error immediately if the host zone is halted
	if hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHaltedHostZone, "halted host zone found for denom (%s)", msg.HostDenom)
	}

	// Get user and module account addresses
	liquidStakerAddress, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "user's address is invalid")
	}
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Safety check: redemption rate must be within safety bounds
	rateIsSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, *hostZone)
	if !rateIsSafe || (err != nil) {
		return nil, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds, "HostZone: %s, err: %s", hostZone.ChainId, err.Error())
	}

	// Grab the deposit record that will be used for record keeping
	strideEpochTracker, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no epoch number for epoch (%s)", epochtypes.STRIDE_EPOCH)
	}
	depositRecord, found := k.RecordsKeeper.GetTransferDepositRecordByEpochAndChain(ctx, strideEpochTracker.EpochNumber, hostZone.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "no deposit record for epoch (%d)", strideEpochTracker.EpochNumber)
	}

	// The tokens that are sent to the protocol are denominated in the ibc hash of the native token on stride (e.g. ibc/xxx)
	nativeDenom := hostZone.IbcDenom
	nativeCoin := sdk.NewCoin(nativeDenom, msg.Amount)
	if !types.IsIBCToken(nativeDenom) {
		return nil, errorsmod.Wrapf(types.ErrInvalidToken, "denom is not an IBC token (%s)", nativeDenom)
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, nativeDenom)
	if balance.IsLT(nativeCoin) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "balance is lower than staking amount. staking amount: %v, balance: %v", msg.Amount, balance.Amount)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Transfer the native tokens from the user to module account
	// Note: checkBlockedAddr=false because hostZoneDepositAddress is a module
	if err := utils.SafeSendCoins(false, k.bankKeeper, ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(nativeCoin)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Mint the stTokens and transfer them to the user
	stDenom := types.StAssetDenomFromHostZoneDenom(msg.HostDenom)
	stCoin := sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to mint coins")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account", stCoin.String())
	}

	// Update the liquid staked amount on the deposit record
	depositRecord.Amount = depositRecord.Amount.Add(msg.Amount)
	k.RecordsKeeper.SetDepositRecord(ctx, *depositRecord)

	// Emit liquid stake event
	EmitSuccessfulLiquidStakeEvent(ctx, msg, *hostZone, stAmount)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return &types.MsgLiquidStakeResponse{StToken: stCoin}, nil
}

func (k msgServer) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.RedeemStake(ctx, msg)
}

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride as an IBC voucher (whose denomtrace we recognize)
// before this function is called
//
// The typical flow:
//   - A staker tokenizes their delegation on the host zone
//   - The staker IBC transfers their tokenized shares to Stride
//   - They then call LSMLiquidStake
//   - - The staker's LSM Tokens are sent to the Stride module account
//   - - The staker recieves stTokens
//
// As a safety measure, at period checkpoints, the validator's sharesToTokens rate is queried and the transaction
// is not settled until the query returns
// As a result, this transaction has been split up into a (1) Start and (2) Finish function
//   - If no query is needed, (2) is called immediately after (1)
//   - If a query is needed, (2) is called in the query callback
//
// The transaction response indicates if the query occurred by returning an attribute `TransactionComplete` set to false
func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lsmLiquidStake, err := k.StartLSMLiquidStake(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if k.ShouldCheckIfValidatorWasSlashed(ctx, *lsmLiquidStake.Validator, msg.Amount) {
		if err := k.SubmitValidatorSlashQuery(ctx, lsmLiquidStake); err != nil {
			return nil, err
		}

		EmitPendingLSMLiquidStakeEvent(ctx, *lsmLiquidStake.HostZone, *lsmLiquidStake.Deposit)

		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	async := false
	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake, async); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// Gov tx to register a trade route that swaps reward tokens for a different denom
//
// Example proposal:
//
//		{
//		   "title": "Create a new trade route for host chain X",
//		   "metadata": "Create a new trade route for host chain X",
//		   "summary": "Create a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgCreateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//
//		         "stride_to_host_connection_id": "connection-0",
//				 "stride_to_reward_connection_id": "connection-1",
//			     "stride_to_trade_connection_id": "connection-2",
//
//				 "host_to_reward_transfer_channel_id": "channel-0",
//				 "reward_to_trade_transfer_channel_id": "channel-1",
//			     "trade_to_host_transfer_channel_id": "channel-2",
//
//				 "reward_denom_on_host": "ibc/rewardTokenXXX",
//				 "reward_denom_on_reward": "rewardToken",
//				 "reward_denom_on_trade": "ibc/rewardTokenYYY",
//				 "host_denom_on_trade": "ibc/hostTokenZZZ",
//				 "host_denom_on_host": "hostToken",
//
//				 "min_transfer_amount": "10000000",
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) CreateTradeRoute(goCtx context.Context, msg *types.MsgCreateTradeRoute) (*types.MsgCreateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	// Validate trade route does not already exist for this denom
	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenomOnReward, msg.HostDenomOnHost)
	if found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteAlreadyExists,
			"trade route already exists for rewardDenom %s, hostDenom %s", msg.RewardDenomOnReward, msg.HostDenomOnHost)
	}

	// Confirm the host chain exists and the withdrawal address has been initialized
	hostZone, err := ms.Keeper.GetActiveHostZone(ctx, msg.HostChainId)
	if err != nil {
		return nil, err
	}
	if hostZone.WithdrawalIcaAddress == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "withdrawal account not initialized on host zone")
	}

	// Register the new ICA accounts
	tradeRouteId := types.GetTradeRouteId(msg.RewardDenomOnReward, msg.HostDenomOnHost)
	hostICA := types.ICAAccount{
		ChainId:      msg.HostChainId,
		Type:         types.ICAAccountType_WITHDRAWAL,
		ConnectionId: hostZone.ConnectionId,
		Address:      hostZone.WithdrawalIcaAddress,
	}

	unwindConnectionId := msg.StrideToRewardConnectionId
	unwindICAType := types.ICAAccountType_CONVERTER_UNWIND
	unwindICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, unwindConnectionId, unwindICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the unwind ICA account")
	}

	tradeConnectionId := msg.StrideToTradeConnectionId
	tradeICAType := types.ICAAccountType_CONVERTER_TRADE
	tradeICA, err := ms.Keeper.RegisterTradeRouteICAAccount(ctx, tradeRouteId, tradeConnectionId, tradeICAType)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register the trade ICA account")
	}

	// Finally build and store the main trade route
	tradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   msg.RewardDenomOnHost,
		RewardDenomOnRewardZone: msg.RewardDenomOnReward,
		RewardDenomOnTradeZone:  msg.RewardDenomOnTrade,
		HostDenomOnTradeZone:    msg.HostDenomOnTrade,
		HostDenomOnHostZone:     msg.HostDenomOnHost,

		HostAccount:   hostICA,
		RewardAccount: unwindICA,
		TradeAccount:  tradeICA,

		HostToRewardChannelId:  msg.HostToRewardTransferChannelId,
		RewardToTradeChannelId: msg.RewardToTradeTransferChannelId,
		TradeToHostChannelId:   msg.TradeToHostTransferChannelId,

		MinTransferAmount: msg.MinTransferAmount,
	}

	ms.Keeper.SetTradeRoute(ctx, tradeRoute)

	return &types.MsgCreateTradeRouteResponse{}, nil
}

// Gov tx to remove a trade route
//
// Example proposal:
//
//		{
//		   "title": "Remove a new trade route for host chain X",
//		   "metadata": "Remove a new trade route for host chain X",
//		   "summary": "Remove a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgDeleteTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//				 "reward_denom": "rewardToken",
//				 "host_denom": "hostToken
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) DeleteTradeRoute(goCtx context.Context, msg *types.MsgDeleteTradeRoute) (*types.MsgDeleteTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	ms.Keeper.RemoveTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)

	return &types.MsgDeleteTradeRouteResponse{}, nil
}

// Gov tx to update the trade route
//
// Example proposal:
//
//		{
//		   "title": "Update a the trade route for host chain X",
//		   "metadata": "Update a the trade route for host chain X",
//		   "summary": "Update a the trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgUpdateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//				 "min_transfer_amount": "10000000",
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
func (ms msgServer) UpdateTradeRoute(goCtx context.Context, msg *types.MsgUpdateTradeRoute) (*types.MsgUpdateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	route, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	route.MinTransferAmount = msg.MinTransferAmount
	ms.Keeper.SetTradeRoute(ctx, route)

	return &types.MsgUpdateTradeRouteResponse{}, nil
}

func (k msgServer) RestoreInterchainAccount(goCtx context.Context, msg *types.MsgRestoreInterchainAccount) (*types.MsgRestoreInterchainAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get ConnectionEnd (for counterparty connection)
	connectionEnd, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)
	if !found {
		return nil, errorsmod.Wrapf(connectiontypes.ErrConnectionNotFound, "connection %s not found", msg.ConnectionId)
	}
	counterpartyConnection := connectionEnd.Counterparty

	// only allow restoring an account if it already exists
	portID, err := icatypes.NewControllerPortID(msg.AccountOwner)
	if err != nil {
		return nil, err
	}
	_, exists := k.ICAControllerKeeper.GetInterchainAccountAddress(ctx, msg.ConnectionId, portID)
	if !exists {
		return nil, errorsmod.Wrapf(types.ErrInvalidInterchainAccountAddress,
			"ICA controller account address not found: %s", msg.AccountOwner)
	}

	appVersion := string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: msg.ConnectionId,
		HostConnectionId:       counterpartyConnection.ConnectionId,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))

	if err := k.ICAControllerKeeper.RegisterInterchainAccountWithOrdering(ctx, msg.ConnectionId, msg.AccountOwner, appVersion, channeltypes.ORDERED); err != nil {
		return nil, errorsmod.Wrapf(err, "unable to register account for owner %s", msg.AccountOwner)
	}

	// If we're restoring a delegation account, we also have to reset record state
	if msg.AccountOwner == types.FormatHostZoneICAOwner(msg.ChainId, types.ICAAccountType_DELEGATION) {
		hostZone, found := k.GetHostZone(ctx, msg.ChainId)
		if !found {
			return nil, types.ErrHostZoneNotFound.Wrapf("delegation ICA supplied, but no associated host zone")
		}

		// Since any ICAs along the original channel will never get relayed,
		// we have to reset the delegation_changes_in_progress field on each validator
		for _, validator := range hostZone.Validators {
			validator.DelegationChangesInProgress = 0
		}
		k.SetHostZone(ctx, hostZone)

		// revert DELEGATION_IN_PROGRESS records for the closed ICA channel (so that they can be staked)
		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)
		for _, depositRecord := range depositRecords {
			// only revert records for the select host zone
			if depositRecord.HostZoneId == hostZone.ChainId && depositRecord.Status == recordtypes.DepositRecord_DELEGATION_IN_PROGRESS {
				depositRecord.Status = recordtypes.DepositRecord_DELEGATION_QUEUE
				depositRecord.DelegationTxsInProgress = 0

				k.Logger(ctx).Info(fmt.Sprintf("Setting DepositRecord %d to status DepositRecord_DELEGATION_IN_PROGRESS", depositRecord.Id))
				k.RecordsKeeper.SetDepositRecord(ctx, depositRecord)
			}
		}

		// revert epoch unbonding records for the closed ICA channel
		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		for _, epochUnbondingRecord := range epochUnbondingRecords {
			// only revert records for the select host zone
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			if !found {
				k.Logger(ctx).Info(fmt.Sprintf("No HostZoneUnbonding found for chainId: %s, epoch: %d", hostZone.ChainId, epochUnbondingRecord.EpochNumber))
				continue
			}

			// Reset the number of undelegation txs in progress
			hostZoneUnbonding.UndelegationTxsInProgress = 0

			// Revert UNBONDING_IN_PROGRESS records to UNBONDING_RETRY_QUEUE
			// and EXIT_TRANSFER_IN_PROGRESS records to EXIT_TRANSFER_QUEUE
			if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d is stuck in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS.String(),
				))
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE

			} else if hostZoneUnbonding.Status == recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS {
				k.Logger(ctx).Info(fmt.Sprintf("HostZoneUnbonding for %s at EpochNumber %d to in status %s",
					hostZone.ChainId, epochUnbondingRecord.EpochNumber, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS.String(),
				))
				hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
			}

			err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId, *hostZoneUnbonding)
			if err != nil {
				return nil, err
			}
		}

		// Revert all pending LSM Detokenizations from status DETOKENIZATION_IN_PROGRESS to status DETOKENIZATION_QUEUE
		pendingDeposits := k.RecordsKeeper.GetLSMDepositsForHostZoneWithStatus(ctx, hostZone.ChainId, recordtypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS)
		for _, lsmDeposit := range pendingDeposits {
			k.Logger(ctx).Info(fmt.Sprintf("Setting LSMTokenDeposit %s to status DETOKENIZATION_QUEUE", lsmDeposit.Denom))
			k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, lsmDeposit, recordtypes.LSMTokenDeposit_DETOKENIZATION_QUEUE)
		}
	}

	return &types.MsgRestoreInterchainAccountResponse{}, nil
}

// Admin transaction to close an ICA channel by sending an ICA with a 1 nanosecond timeout (which will force a timeout and closure)
// This can be used if there are records stuck in state IN_PROGRESS after a channel has been re-opened after a timeout
// After the closure, the a new channel can be permissionlessly re-opened with RestoreInterchainAccount
func (k msgServer) CloseDelegationChannel(goCtx context.Context, msg *types.MsgCloseDelegationChannel) (*types.MsgCloseDelegationChannelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("chain id %s", msg.ChainId)
	}

	// Submit an ICA bank send from the delegation ICA account to itself for just 1utoken
	delegationIcaOwner := types.FormatHostZoneICAOwner(msg.ChainId, types.ICAAccountType_DELEGATION)
	msgSend := []proto.Message{&banktypes.MsgSend{
		FromAddress: hostZone.DelegationIcaAddress,
		ToAddress:   hostZone.DelegationIcaAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(hostZone.HostDenom, sdkmath.OneInt())),
	}}

	// Timeout the ICA 1 nanosecond after the current block time (so it's impossible to be relayed)
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().UnixNano() + 1)
	err := k.SubmitICATxWithoutCallback(ctx, hostZone.ConnectionId, delegationIcaOwner, msgSend, timeoutTimestamp)
	if err != nil {
		return nil, err
	}

	return &types.MsgCloseDelegationChannelResponse{}, nil
}

// This kicks off two ICQs, each with a callback, that will update the number of tokens on a validator
// after being slashed. The flow is:
// 1. QueryValidatorSharesToTokensRate (ICQ)
// 2. ValidatorSharesToTokensRate (CALLBACK)
// 3. SubmitDelegationICQ (ICQ)
// 4. DelegatorSharesCallback (CALLBACK)
func (k msgServer) UpdateValidatorSharesExchRate(goCtx context.Context, msg *types.MsgUpdateValidatorSharesExchRate) (*types.MsgUpdateValidatorSharesExchRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.QueryValidatorSharesToTokensRate(ctx, msg.ChainId, msg.Valoper); err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorSharesExchRateResponse{}, nil
}

// Submits an ICQ to get the validator's delegated shares
func (k msgServer) CalibrateDelegation(goCtx context.Context, msg *types.MsgCalibrateDelegation) (*types.MsgCalibrateDelegationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound
	}

	if err := k.SubmitCalibrationICQ(ctx, hostZone, msg.Valoper); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for delegation, error : %s", err.Error()))
		return nil, err
	}

	return &types.MsgCalibrateDelegationResponse{}, nil
}

func (k msgServer) UpdateInnerRedemptionRateBounds(goCtx context.Context, msg *types.MsgUpdateInnerRedemptionRateBounds) (*types.MsgUpdateInnerRedemptionRateBoundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: we're intentionally not checking the zone is halted
	zone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("Host Zone not found: %s", msg.ChainId))
		return nil, types.ErrInvalidHostZone
	}

	// Get the wide bounds
	outerMinSafetyThreshold, outerMaxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	innerMinSafetyThreshold := msg.MinInnerRedemptionRate
	innerMaxSafetyThreshold := msg.MaxInnerRedemptionRate

	// Confirm the inner bounds are within the outer bounds
	if innerMinSafetyThreshold.LT(outerMinSafetyThreshold) {
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds,
			"inner min safety threshold (%s) is less than outer min safety threshold (%s)",
			innerMinSafetyThreshold, outerMinSafetyThreshold)
	}

	if innerMaxSafetyThreshold.GT(outerMaxSafetyThreshold) {
		return nil, errorsmod.Wrapf(types.ErrInvalidBounds,
			"inner max safety threshold (%s) is greater than outer max safety threshold (%s)",
			innerMaxSafetyThreshold, outerMaxSafetyThreshold)
	}

	// Set the inner bounds on the host zone
	zone.MinInnerRedemptionRate = innerMinSafetyThreshold
	zone.MaxInnerRedemptionRate = innerMaxSafetyThreshold

	k.SetHostZone(ctx, zone)

	return &types.MsgUpdateInnerRedemptionRateBoundsResponse{}, nil
}

func (k msgServer) ResumeHostZone(goCtx context.Context, msg *types.MsgResumeHostZone) (*types.MsgResumeHostZoneResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get Host Zone
	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone %s not found", msg.ChainId)
	}

	// Check the zone is halted
	if !hostZone.Halted {
		return nil, errorsmod.Wrapf(types.ErrHostZoneNotHalted, "host zone %s is not halted", msg.ChainId)
	}

	// remove from blacklist
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	k.RatelimitKeeper.RemoveDenomFromBlacklist(ctx, stDenom)

	// Resume zone
	hostZone.Halted = false
	k.SetHostZone(ctx, hostZone)

	return &types.MsgResumeHostZoneResponse{}, nil
}

// Registers or updates a community pool rebate, configuring the rebate percentage and liquid stake amount
func (k msgServer) SetCommunityPoolRebate(
	goCtx context.Context,
	msg *types.MsgSetCommunityPoolRebate,
) (*types.MsgSetCommunityPoolRebateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	hostZone, found := k.GetHostZone(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", msg.ChainId)
	}

	// Get the current stToken supply and confirm it's greater than or equal to the liquid staked amount
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stTokenSupply := k.bankKeeper.GetSupply(ctx, stDenom).Amount
	if msg.LiquidStakedStTokenAmount.GT(stTokenSupply) {
		return nil, types.ErrFailedToRegisterRebate.Wrapf("liquid staked stToken amount (%v) is greater than current supply (%v)",
			msg.LiquidStakedStTokenAmount, stTokenSupply)
	}

	// If a zero rebate rate or zero LiquidStakedStTokenAmount is specified, set the rebate to nil
	// Otherwise, update the struct
	if msg.LiquidStakedStTokenAmount.IsZero() || msg.RebateRate.IsZero() {
		hostZone.CommunityPoolRebate = nil
	} else {
		hostZone.CommunityPoolRebate = &types.CommunityPoolRebate{
			LiquidStakedStTokenAmount: msg.LiquidStakedStTokenAmount,
			RebateRate:                msg.RebateRate,
		}
	}

	k.SetHostZone(ctx, hostZone)

	return &types.MsgSetCommunityPoolRebateResponse{}, nil
}

// Submits an ICA tx to either grant or revoke authz permisssions to an address
// to execute trades on behalf of the trade ICA
func (k msgServer) ToggleTradeController(
	goCtx context.Context,
	msg *types.MsgToggleTradeController,
) (*types.MsgToggleTradeControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Fetch the trade ICA which will be the granter
	tradeRoute, found := k.GetTradeRouteFromTradeAccountChainId(ctx, msg.ChainId)
	if !found {
		return nil, types.ErrTradeRouteNotFound.Wrapf("trade route not found for chain ID %s", msg.ChainId)
	}

	// Build the authz message that grants or revokes trade permissions to the specified address
	authzMsg, err := k.BuildTradeAuthzMsg(ctx, tradeRoute, msg.PermissionChange, msg.Address, msg.Legacy)
	if err != nil {
		return nil, err
	}

	// Build the ICA channel owner from the trade route
	tradeRouteAccountOwner := types.FormatTradeRouteICAOwnerFromRouteId(
		msg.ChainId,
		tradeRoute.GetRouteId(),
		types.ICAAccountType_CONVERTER_TRADE,
	)

	// Submit the ICA tx from the trade ICA account
	// Timeout the ICA at 1 hour
	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(time.Hour).UnixNano())
	err = k.SubmitICATxWithoutCallback(
		ctx,
		tradeRoute.TradeAccount.ConnectionId,
		tradeRouteAccountOwner,
		authzMsg,
		timeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgToggleTradeControllerResponse{}, nil
}
