package keeper

import (
	"context"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"
	icqtypes "github.com/Stride-Labs/stride/v8/x/interchainquery/types"

	recordskeeper "github.com/Stride-Labs/stride/v8/x/records/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride as an IBC voucher (whose denomtrace we recognize)
//   before this function is called
//
// The typical flow:
//   - A staker tokenizes their delegation on the host zone
//   - The staker IBC transfers their tokenized shares to Stride
//   - They then call LSMLiquidStake
//     - The staker's LSM Tokens are sent to the Stride module account
//     - The staker recieves stTokens
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

	if k.ShouldCheckIfValidatorWasSlashed(ctx, lsmLiquidStake.Validator, msg.Amount) {
		if err := k.SubmitValidatorSlashQuery(ctx, lsmLiquidStake); err != nil {
			return nil, err
		}
		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// StartLSMLiquidStake runs the transactional logic that occurs before the optional query
// This includes validation on the LSM Token, and the escrowing of tokens
func (k Keeper) StartLSMLiquidStake(ctx sdk.Context, msg *types.MsgLSMLiquidStake) (types.LSMLiquidStake, error) {
	// Get the denom trace from the IBC hash - this includes the full path and base denom
	// Ex: LSMTokenIbcDenom of `ibc/XXX` might create a DenomTrace with:
	//     BaseDenom: cosmosvaloperXXX/42, Path: transfer/channel-0
	denomTrace, err := k.GetLSMTokenDenomTrace(ctx, msg.LsmTokenIbcDenom)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}

	// Get the host zone and validator address from the path and base denom respectively
	lsmTokenBaseDenom := denomTrace.BaseDenom
	hostZone, err := k.GetHostZoneFromLSMTokenPath(ctx, denomTrace.Path)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}
	validator, err := k.GetValidatorFromLSMTokenDenom(lsmTokenBaseDenom, hostZone.Validators)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}

	// Get the staker's address and the host zone module account address that will custody the tokens
	liquidStakerAddress := sdk.MustAccAddressFromBech32(msg.Creator)
	hostZoneAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Confirm the staker has a sufficient balance to execute the liquid stake
	stakeAmount := msg.Amount
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, msg.LsmTokenIbcDenom).Amount
	if balance.LT(stakeAmount) {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"balance is lower than staking amount. staking amount: %v, balance: %v", stakeAmount, balance)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}
	stCoin := sdk.NewCoin(stDenom, stAmount)

	// Transfer the LSM token to the host zone module account
	lsmTokenCoin := sdk.NewCoin(msg.LsmTokenIbcDenom, msg.Amount)
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneAddress, sdk.NewCoins(lsmTokenCoin)); err != nil {
		return types.LSMLiquidStake{}, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Store an deposit record for the LSM token
	lsmTokenDeposit := types.LSMTokenDeposit{
		ChainId:          hostZone.ChainId,
		Denom:            lsmTokenBaseDenom,
		ValidatorAddress: validator.Address,
		Amount:           msg.Amount,
		Status:           types.DEPOSIT_PENDING,
	}
	k.AddLSMTokenDeposit(ctx, lsmTokenDeposit)

	return types.LSMLiquidStake{
		Staker:      liquidStakerAddress,
		LSMIBCToken: lsmTokenCoin,
		StToken:     stCoin,
		HostZone:    hostZone,
		Validator:   validator,
		Deposit:     lsmTokenDeposit,
	}, nil
}

// SubmitValidatorSlashQuery submits an interchain query for the validator's exchange rate
// This is done periodically at checkpoints denominated in native tokens
// (e.g. every 100k ATOM that's LSM liquid staked with validator X)
func (k Keeper) SubmitValidatorSlashQuery(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
	hostZone := lsmLiquidStake.HostZone
	validator := lsmLiquidStake.Validator

	// Encode the validator address for the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validator.Address)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid validator operator address, could not decode")
	}
	queryData := stakingtypes.GetValidatorKey(validatorAddressBz)

	// Build and serialize the callback data required to complete the LSM Liquid stake upon query callback
	callbackData, err := json.Marshal(lsmLiquidStake)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to serialize LSMLiquidStake struct for validator exchange rate query callback")
	}

	// Use a short timeout for the query so that user's can get the tokens refunded quickly should the query get stuck
	timeout := uint64(ctx.BlockTime().UnixNano() + (SlashQueryTimeout).Nanoseconds())
	query := icqtypes.Query{
		ChainId:        hostZone.ChainId,
		ConnectionId:   hostZone.ConnectionId,
		QueryType:      icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:    queryData,
		CallbackModule: types.ModuleName,
		CallbackId:     ICQCallbackID_Validator,
		CallbackData:   callbackData,
		Timeout:        timeout,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrapf(err, "Unable to submit validator exchange rate query")
	}

	return nil
}

// FinishLSMLiquidStake finishes the liquid staking flow by sending a user an stToken and
// IBC transfering the LSM Token to the host zone
//
// If the validator exchange rate query interrupted the transaction, this function is called
// asynchronously after the query callback
// If no validator exchange rate query was needed, this is called synchronously after StartLSMLiquidStake
func (k Keeper) FinishLSMLiquidStake(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
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
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// Send LSM Token to host zone via IBC transfer
	timeout := uint64(ctx.BlockTime().UnixNano() + (LSMDepositTransferTimeout).Nanoseconds()) // 1 day
	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            lsmLiquidStake.LSMIBCToken,
		Sender:           hostZone.DepositAddress,
		Receiver:         hostZone.DelegationIcaAddress,
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
