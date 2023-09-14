package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

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

	// Determine the amount of stTokens to mint using the redemption rate and the validator's sharesToTokens rate
	//    StTokens = LSMTokenShares * Validator SharesToTokens Rate / Redemption Rate
	// Note: in the event of a slash query, these tokens will be minted only if the
	// validator's sharesToTokens rate did not change
	stCoin := k.CalculateLSMStToken(msg.Amount, lsmLiquidStake)
	if stCoin.Amount.IsZero() {
		return types.LSMLiquidStake{}, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Add the stToken to this deposit record
	lsmLiquidStake.Deposit.StToken = stCoin
	k.RecordsKeeper.SetLSMTokenDeposit(ctx, *lsmLiquidStake.Deposit)

	return lsmLiquidStake, nil
}

// SubmitValidatorSlashQuery submits an interchain query for the validator's sharesToTokens rate
// This is done periodically at checkpoints denominated in native tokens
// (e.g. every 100k ATOM that's LSM liquid staked with validator X)
func (k Keeper) SubmitValidatorSlashQuery(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
	chainId := lsmLiquidStake.HostZone.ChainId
	validatorAddress := lsmLiquidStake.Validator.Address
	timeoutDuration := LSMSlashQueryTimeout
	timeoutPolicy := icqtypes.TimeoutPolicy_EXECUTE_QUERY_CALLBACK

	// Build and serialize the callback data required to complete the LSM Liquid stake upon query callback
	callbackData := types.ValidatorSharesToTokensQueryCallback{
		LsmLiquidStake: &lsmLiquidStake,
	}
	callbackDataBz, err := proto.Marshal(&callbackData)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to serialize LSMLiquidStake struct for validator sharesToTokens rate query callback")
	}

	return k.SubmitValidatorSharesToTokensRateICQ(ctx, chainId, validatorAddress, callbackDataBz, timeoutDuration, timeoutPolicy)
}

// FinishLSMLiquidStake finishes the liquid staking flow by escrowing the LSM token,
// sending a user their stToken, and then IBC transfering the LSM Token to the host zone
//
// If the slash query interrupted the transaction, this function is called
// asynchronously after the query callback
//
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
	hostZoneDepositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "host zone address is invalid")
	}

	// Transfer the LSM token to the deposit account
	lsmIBCToken := sdk.NewCoin(lsmTokenDeposit.IbcDenom, lsmTokenDeposit.Amount)
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneDepositAddress, sdk.NewCoins(lsmIBCToken)); err != nil {
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

	// Update the deposit status
	k.RecordsKeeper.UpdateLSMTokenDepositStatus(ctx, lsmTokenDeposit, recordstypes.LSMTokenDeposit_TRANSFER_QUEUE)

	// Update the slash query progress on the validator
	if err := k.IncrementValidatorSlashQueryProgress(
		ctx,
		hostZone.ChainId,
		lsmTokenDeposit.ValidatorAddress,
		lsmTokenDeposit.Amount,
	); err != nil {
		return err
	}

	// Emit an LSM liquid stake event
	EmitSuccessfulLSMLiquidStakeEvent(ctx, *hostZone, lsmTokenDeposit)

	k.hooks.AfterLiquidStake(ctx, liquidStakerAddress)
	return nil
}
