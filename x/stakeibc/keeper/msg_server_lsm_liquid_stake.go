package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbackstypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"

	recordskeeper "github.com/Stride-Labs/stride/v8/x/records/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

// Maintains the context and progress of an LSM Liquid Stake in the event
// that the transaction finishes asynchonously after the validator exchange rate query
type LSMLiquidStake struct {
	Staker    sdk.AccAddress
	LSMToken  sdk.Coin
	StToken   sdk.Coin
	HostZone  types.HostZone
	Validator types.Validator
	Deposit   types.LSMTokenDeposit
}

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride with an IBC denomination before this function is called
// The typical flow:
//   - A user tokenizes their delegation on the host zone
//   - The user IBC transfers their tokenized shares to Stride
//   - They then call LSMLiquidStake
//   - The user's LSM Token is sent to the Stride module account
//   - The user recieves stTokens
//
// As a safety measure, at period checkpoints, the validator's exchange rate is queried and the transaction
// is not settled until the query returns
// As a result, this transaction has been split up into a (1) Start and (2) Finish function
//   If no query is needed, (2) is called immediately after (1)
//   If a query is needed, (2) is called in the query callback
//
// The transaction response indicates if the query occurred by returning an attribute `TransactionComplete` set to false
func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lsmLiquidStake, err := k.StartLSMLiquidStake(ctx, msg)
	if err != nil {
		return nil, err
	}

	if k.ShouldQueryValidatorExchangeRate(ctx, lsmLiquidStake.Validator, msg.Amount) {
		// TODO: Submit query
		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// StartLSMLiquidStake runs the transactional logic that occurs before the optional query
// This includes validation on the LSM Token, and the escrowing of tokens
func (k Keeper) StartLSMLiquidStake(ctx sdk.Context, msg *types.MsgLSMLiquidStake) (LSMLiquidStake, error) {
	// Get the denom trace from the IBC hash - this includes the full path and base denom
	denomTrace, err := k.GetLSMTokenDenomTrace(ctx, msg.LsmTokenIbcDenom)
	if err != nil {
		return LSMLiquidStake{}, err
	}

	// Get the host zone and validator address from the path and base denom respectively
	lsmTokenBaseDenom := denomTrace.BaseDenom
	hostZone, err := k.GetHostZoneFromLSMTokenPath(ctx, denomTrace.Path)
	if err != nil {
		return LSMLiquidStake{}, err
	}
	validator, err := k.GetValidatorFromLSMTokenDenom(lsmTokenBaseDenom, hostZone.Validators)
	if err != nil {
		return LSMLiquidStake{}, err
	}

	// Get the user address and the host zone module account address that will custody the tokens
	liquidStakerAddress := sdk.MustAccAddressFromBech32(msg.Creator)
	hostZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return LSMLiquidStake{}, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	stakeAmount := msg.Amount
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, msg.LsmTokenIbcDenom).Amount
	if balance.LT(stakeAmount) {
		return LSMLiquidStake{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"balance is lower than staking amount. staking amount: %v, balance: %v", stakeAmount, balance)
	}

	// Transfer the LSM token to the host zone module account
	lsmTokenCoin := sdk.NewCoin(msg.LsmTokenIbcDenom, msg.Amount)
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneAddress, sdk.NewCoins(lsmTokenCoin)); err != nil {
		return LSMLiquidStake{}, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return LSMLiquidStake{}, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}
	stCoin := sdk.NewCoin(stDenom, stAmount)

	// Store a record for the LSM token
	lsmTokenDeposit := types.LSMTokenDeposit{
		ChainId:          hostZone.ChainId,
		Denom:            lsmTokenBaseDenom,
		ValidatorAddress: validator.Address,
		Amount:           msg.Amount,
		Status:           types.DEPOSIT_PENDING,
	}
	k.AddLSMTokenDeposit(ctx, lsmTokenDeposit)

	return LSMLiquidStake{
		Staker:    liquidStakerAddress,
		LSMToken:  lsmTokenCoin,
		StToken:   stCoin,
		HostZone:  hostZone,
		Validator: validator,
		Deposit:   lsmTokenDeposit,
	}, nil
}

// FinishLSMLiquidStake finishes the liquid staking flow by sending a user an stToken and
// IBC transfering the LSM Token to the host zone
//
// If the validator exchange rate query interrupted the transaction, this function is called
// asynchronously upon the query callback
// If no validator exchange rate query was needed, this is called synchronously after StartLSMLiquidStake
func (k Keeper) FinishLSMLiquidStake(ctx sdk.Context, lsmLiquidStake LSMLiquidStake) error {
	// Mint stToken and send to the user
	stToken := sdk.NewCoins(lsmLiquidStake.StToken)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, stToken); err != nil {
		return errorsmod.Wrapf(err, "Failed to mint stTokens")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lsmLiquidStake.Staker, stToken); err != nil {
		return errorsmod.Wrapf(err, "Failed to send %s from module to account", lsmLiquidStake.StToken.String())
	}

	// Get delegation account address as the destination for the LSM Token
	hostZone := lsmLiquidStake.HostZone
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// Send LSM Token to host zone with IBC transfer
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour * 24).Nanoseconds()) // 1 day
	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            lsmLiquidStake.LSMToken,
		Sender:           hostZone.Address,
		Receiver:         delegationAccount.Address,
		TimeoutTimestamp: timeout,
	})
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to submit IBC transfer of LSM token")
	}

	// Store transfer callback data
	callbackArgs := types.TransferLSMTokenCallback{
		Deposit: &lsmLiquidStake.Deposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	if err != nil {
		return errorsmod.Wrapf(err, "Unable to marshal transfer callback data for %+v", callbackArgs)
	}

	k.RecordsKeeper.ICACallbacksKeeper.SetCallbackData(ctx, icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(transfertypes.PortID, hostZone.TransferChannelId, msgTransferResponse.Sequence),
		PortId:       transfertypes.PortID,
		ChannelId:    hostZone.TransferChannelId,
		Sequence:     msgTransferResponse.Sequence,
		CallbackId:   recordskeeper.IBCCallbacksID_LSMTransfer,
		CallbackArgs: callbackArgsBz,
	})

	// Update the deposit status
	k.UpdateLSMTokenDepositStatus(ctx, lsmLiquidStake.Deposit, types.TRANSFER_IN_PROGRESS)

	return nil
}
