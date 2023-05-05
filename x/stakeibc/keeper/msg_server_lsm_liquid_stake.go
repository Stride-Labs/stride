package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto" //nolint:staticcheck

	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"

	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Exchanges a user's LSM tokenized shares for stTokens using the current redemption rate
// The LSM tokens must live on Stride as an IBC voucher (whose denomtrace we recognize)
//	 before this function is called
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

	lsmLiquidStake, err := k.StartLSMLiquidStake(ctx, *msg)
	if err != nil {
		return nil, err
	}

	if k.ShouldCheckIfValidatorWasSlashed(ctx, *lsmLiquidStake.Validator, msg.Amount) {
		if err := k.SubmitValidatorSlashQuery(ctx, lsmLiquidStake); err != nil {
			return nil, err
		}
		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	async := false
	if err := k.FinishLSMLiquidStake(ctx, lsmLiquidStake, async); err != nil {
		return nil, err
	}

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// StartLSMLiquidStake runs the transactional logic that occurs before the optional query
// This includes validation on the LSM Token and the stToken amount calculation
func (k Keeper) StartLSMLiquidStake(ctx sdk.Context, msg types.MsgLSMLiquidStake) (types.LSMLiquidStake, error) {
	// Validate the provided message parameters - including the denom and staker balance
	lsmLiquidStake, err := k.ValidateLSMLiquidStake(ctx, msg)
	if err != nil {
		return types.LSMLiquidStake{}, err
	}
	hostZone := lsmLiquidStake.HostZone

	// Check if we already have tokens with this denom in records
	_, found := k.RecordsKeeper.GetLSMTokenDeposit(ctx, hostZone.ChainId, lsmLiquidStake.Deposit.Denom)
	if found {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"there is already a previous record with this denom being processed: %s", lsmLiquidStake.Deposit.Denom)
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}
	stCoin := sdk.NewCoin(stDenom, stAmount)

	// Add the stToken to this deposit record
	lsmLiquidStake.Deposit.StToken = stCoin
	k.RecordsKeeper.SetLSMTokenDeposit(ctx, *lsmLiquidStake.Deposit)

	return lsmLiquidStake, nil
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
	callbackData := types.ValidatorExchangeRateQueryCallback{
		LsmLiquidStake: &lsmLiquidStake,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to serialize LSMLiquidStake struct for validator exchange rate query callback")
	}

	// Use a short timeout for the query so that user's can get the tokens refunded quickly should the query get stuck
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Validator,
		CallbackData:    callbackDataBz,
		TimeoutDuration: LSMSlashQueryTimeout,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrapf(err, "Unable to submit validator exchange rate query")
	}

	return nil
}

// FinishLSMLiquidStake finishes the liquid staking flow by escrowing the LSM token,
//   sending a user their stToken, and then IBC transfering the LSM Token to the host zone
//
// If the slash query interrupted the transaction, this function is called
//   asynchronously after the query callback
// If no slash query was needed, this is called synchronously after StartLSMLiquidStake
// If this is run asynchronously, we need to re-validate the transaction info (e.g. staker's balance)
func (k Keeper) FinishLSMLiquidStake(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake, async bool) error {
	hostZone := lsmLiquidStake.HostZone
	lsmTokenDeposit := *lsmLiquidStake.Deposit

	// If the transaction was interrupted by the slash query,
	//  validate the LSM Liquid stake message parameters again
	// The most significant check here is that the user still has sufficient balance for this LSM liquid stake
	if async {
		lsmLiquidStakeMsg := types.MsgLSMLiquidStake{
			Creator:          lsmTokenDeposit.StakerAddress,
			LsmTokenIbcDenom: lsmTokenDeposit.IbcDenom,
			Amount:           lsmTokenDeposit.Amount,
		}
		if _, err := k.ValidateLSMLiquidStake(ctx, lsmLiquidStakeMsg); err != nil {
			return err
		}
	}

	// Get the staker's address and the host zone's deposit account address (which will custody the tokens)
	liquidStakerAddress := sdk.MustAccAddressFromBech32(lsmTokenDeposit.StakerAddress)
	hostZoneAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Transfer the LSM token to the deposit account
	lsmIBCToken := sdk.NewCoin(lsmTokenDeposit.IbcDenom, lsmTokenDeposit.Amount)
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneAddress, sdk.NewCoins(lsmIBCToken)); err != nil {
		return errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Mint stToken and send to the user
	stToken := sdk.NewCoins(lsmTokenDeposit.StToken)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, stToken); err != nil {
		return errorsmod.Wrapf(err, "Failed to mint stTokens")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, stToken); err != nil {
		return errorsmod.Wrapf(err, "Failed to send %s from module to account", lsmTokenDeposit.StToken.String())
	}

	// Get delegation account address as the destination for the LSM Token
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// Send LSM Token via IBC transfer
	if err := k.RecordsKeeper.IBCTransferLSMToken(
		ctx,
		lsmTokenDeposit,
		hostZone.TransferChannelId,
		hostZone.DepositAddress,
		hostZone.DelegationIcaAddress,
	); err != nil {
		return errorsmod.Wrapf(err, "Failed to submit IBC transfer of LSM token")
	}

	// Update the deposit status
	k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, lsmTokenDeposit, recordstypes.LSMTokenDeposit_TRANSFER_IN_PROGRESS)

	// Emit an LSM liquid stake event
	EmitSuccessfulLSMLiquidStakeEvent(ctx, *hostZone, lsmTokenDeposit)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return nil
}
