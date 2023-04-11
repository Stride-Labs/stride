package keeper

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbackstypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"

	recordskeeper "github.com/Stride-Labs/stride/v8/x/records/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

var (
	// QUESTION: Is this too restrictive? Would we ever want to accept a LSM token from more than 1 hop away?
	// A valid IBC path for the LSM token must only consist of 1 channel hop along a transfer channel
	// (e.g. "transfer/channel-0")
	IsValidIBCPath = regexp.MustCompile(fmt.Sprintf(`^%s/(%s[0-9]{1,20})$`, transfertypes.PortID, channeltypes.ChannelPrefix)).MatchString
)

func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the denom trace from the IBC hash - this includes the full path and base denom
	denomTrace, err := k.GetLSMTokenDenomTrace(ctx, msg.LsmTokenIbcDenom)
	if err != nil {
		return nil, err
	}

	// Get the host zone and validator address from the path and base denom respectively
	lsmTokenBaseDenom := denomTrace.BaseDenom
	hostZone, err := k.GetHostZoneFromLSMTokenPath(ctx, denomTrace.Path)
	if err != nil {
		return nil, err
	}
	validator, err := k.GetValidatorFromLSMTokenDenom(lsmTokenBaseDenom, hostZone.Validators)
	if err != nil {
		return nil, err
	}

	// Get the user address and the host zone module account address that will custody the tokens
	liquidStakerAddress := sdk.MustAccAddressFromBech32(msg.Creator)
	hostZoneAddress, err := sdk.AccAddressFromBech32(hostZone.Address)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "host zone address is invalid")
	}
	delegationAccount := hostZone.DelegationAccount
	if delegationAccount == nil || delegationAccount.Address == "" {
		return nil, errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation address found for %s", hostZone.ChainId)
	}

	// Confirm the user has a sufficient balance to execute the liquid stake
	stakeAmount := msg.Amount
	balance := k.bankKeeper.GetBalance(ctx, liquidStakerAddress, msg.LsmTokenIbcDenom).Amount
	if balance.LT(stakeAmount) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds,
			"balance is lower than staking amount. staking amount: %v, balance: %v", stakeAmount, balance)
	}

	// Transfer the LSM token to the host zone module account
	lsmTokenCoin := sdk.NewCoin(msg.LsmTokenIbcDenom, msg.Amount)
	if err := k.bankKeeper.SendCoins(ctx, liquidStakerAddress, hostZoneAddress, sdk.NewCoins(lsmTokenCoin)); err != nil {
		return nil, errorsmod.Wrap(err, "failed to send tokens from Account to Module")
	}

	// Determine the amount of stTokens to mint using the redemption rate
	stDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stAmount := (sdk.NewDecFromInt(msg.Amount).Quo(hostZone.RedemptionRate)).TruncateInt()
	if stAmount.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInsufficientLiquidStake,
			"Liquid stake of %s%s would return 0 stTokens", msg.Amount.String(), hostZone.HostDenom)
	}

	// Check if we need to issue an ICQ to check if a validator was slashed
	params := k.GetParams(ctx)
	queryInterval := sdk.NewIntFromUint64(params.ValidatorExchangeRateQueryInterval)
	if queryInterval.IsZero() {
		return nil, errors.New("Invalid validator exchange rate query interval - must be non-zero")
	}

	oldProgress := validator.ProgressTowardsExchangeRateQuery
	newProgress := validator.ProgressTowardsExchangeRateQuery.Add(msg.Amount)

	if oldProgress.Quo(queryInterval).LT(newProgress.Quo(queryInterval)) {
		// TODO: Submit query
		return &types.MsgLSMLiquidStakeResponse{TransactionComplete: false}, nil
	}

	// Mint stToken and send to the user
	stCoin := sdk.NewCoin(stDenom, stAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to mint stTokens")
	}
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidStakerAddress, sdk.NewCoins(stCoin)); err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to send %s from module to account", stCoin.String())
	}

	// Store a record for the LSM token
	lsmTokenDeposit := types.LSMTokenDeposit{
		ChainId:          hostZone.ChainId,
		Denom:            lsmTokenBaseDenom,
		ValidatorAddress: validator.Address,
		Amount:           msg.Amount,
		Status:           types.TRANSFER_IN_PROGRESS,
	}
	k.AddLSMTokenDeposit(ctx, lsmTokenDeposit)

	// Send LSM Token to host zone in IBC transfer
	timeout := uint64(ctx.BlockTime().UnixNano() + (time.Hour * 24).Nanoseconds()) // 1 day
	msgTransferResponse, err := k.RecordsKeeper.TransferKeeper.Transfer(sdk.WrapSDKContext(ctx), &transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            lsmTokenCoin,
		Sender:           hostZoneAddress.String(),
		Receiver:         delegationAccount.Address,
		TimeoutTimestamp: timeout,
	})
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Failed to submit IBC transfer of LSM token")
	}

	// Store transfer callback data
	callbackArgs := types.TransferLSMTokenCallback{
		Deposit: &lsmTokenDeposit,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "Unable to marshal transfer callback data for %+v", callbackArgs)
	}

	k.RecordsKeeper.ICACallbacksKeeper.SetCallbackData(ctx, icacallbackstypes.CallbackData{
		CallbackKey:  icacallbackstypes.PacketID(transfertypes.PortID, hostZone.TransferChannelId, msgTransferResponse.Sequence),
		PortId:       transfertypes.PortID,
		ChannelId:    hostZone.TransferChannelId,
		Sequence:     msgTransferResponse.Sequence,
		CallbackId:   recordskeeper.IBCCallbacksID_LSMTransfer,
		CallbackArgs: callbackArgsBz,
	})

	return &types.MsgLSMLiquidStakeResponse{TransactionComplete: true}, nil
}

// Parse the LSM Token's IBC denom has into a DenomTrace object that contains the path and base denom
func (k Keeper) GetLSMTokenDenomTrace(ctx sdk.Context, denom string) (transfertypes.DenomTrace, error) {
	ibcPrefix := transfertypes.DenomPrefix + "/"

	// Confirm the LSM Token is a valid IBC token (has "ibc/" prefix)
	if !strings.HasPrefix(denom, ibcPrefix) {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "lsm token is not an IBC token (%s)", denom)
	}

	// Parse the hex hash string into hex bytes
	hexHash := denom[len(ibcPrefix):]
	hash, err := transfertypes.ParseHexHash(hexHash)
	if err != nil {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(err, "unable to get ibc hex hash from denom %s", denom)
	}

	// Lookup the trace from the hash
	denomTrace, found := k.RecordsKeeper.TransferKeeper.GetDenomTrace(ctx, hash)
	if !found {
		return transfertypes.DenomTrace{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "denom trace not found for %s", denom)
	}

	return denomTrace, nil
}

// Parses the LSM token's IBC path (e.g. transfer/channel-0) and confirms the channel ID matches
// the transfer channel of a supported host zone
func (k Keeper) GetHostZoneFromLSMTokenPath(ctx sdk.Context, path string) (types.HostZone, error) {
	if !IsValidIBCPath(path) {
		return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken, "ibc path of LSM token (%s) cannot be more than 1 hop away", path)
	}

	// Remove the "transfer/" prefix
	channelId := strings.ReplaceAll(path, transfertypes.PortID+"/", "")

	// Confirm the channel is from one of Stride's supported host zones
	for _, hostZone := range k.GetAllHostZone(ctx) {
		if hostZone.TransferChannelId == channelId {
			return hostZone, nil
		}
	}

	return types.HostZone{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"transfer channel-id from LSM token (%s) does not match any registered host zone", channelId)
}

// Parses the LSM token's denom (of the form {validatorAddress}/{recordId}) and confirms that the validator
// is in the Stride validator set
func (k Keeper) GetValidatorFromLSMTokenDenom(denom string, validators []*types.Validator) (types.Validator, error) {
	// Denom is of the form {validatorAddress}/{recordId}
	split := strings.Split(denom, "/")
	if len(split) != 2 {
		return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
			"lsm token base denom is not of the format {val-address}/{record-id} (%s)", denom)
	}
	validatorAddress := split[0]

	// Confirm validator is in Stride's validator set
	for _, validator := range validators {
		if validator.Address == validatorAddress {
			return *validator, nil
		}
	}

	return types.Validator{}, errorsmod.Wrapf(types.ErrInvalidLSMToken,
		"validator (%s) is not registered in the Stride validator set", validatorAddress)
}
