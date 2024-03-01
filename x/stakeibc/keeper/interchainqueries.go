package keeper

import (
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v18/utils"
	icqtypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// Submit a validator sharesToTokens rate ICQ as triggered either manually or epochly with a conservative timeout
func (k Keeper) QueryValidatorSharesToTokensRate(ctx sdk.Context, chainId string, validatorAddress string) error {
	timeoutDuration := time.Hour * 24
	timeoutPolicy := icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE
	callbackData := []byte{}
	return k.SubmitValidatorSharesToTokensRateICQ(ctx, chainId, validatorAddress, callbackData, timeoutDuration, timeoutPolicy)
}

// Submits an ICQ to get a validator's shares to tokens rate
func (k Keeper) SubmitValidatorSharesToTokensRateICQ(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	callbackDataBz []byte,
	timeoutDuration time.Duration,
	timeoutPolicy icqtypes.TimeoutPolicy,
) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(chainId, "Submitting ICQ for validator sharesToTokens rate to %s", validatorAddress))

	// Confirm the host zone exists
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidHostZone, "Host zone not found (%s)", chainId)
	}

	// check that the validator address matches the bech32 prefix of the hz
	if !strings.Contains(validatorAddress, hostZone.Bech32Prefix) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "validator operator address must match the host zone bech32 prefix")
	}

	// Encode the validator address to form the query request
	_, validatorAddressBz, err := bech32.DecodeAndConvert(validatorAddress)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid validator operator address, could not decode (%s)", err.Error())
	}
	queryData := stakingtypes.GetValidatorKey(validatorAddressBz)

	// Submit validator sharesToTokens rate ICQ
	// Considering this query is executed manually, we can be conservative with the timeout
	query := icqtypes.Query{
		ChainId:         hostZone.ChainId,
		ConnectionId:    hostZone.ConnectionId,
		QueryType:       icqtypes.STAKING_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_Validator,
		CallbackData:    callbackDataBz,
		TimeoutDuration: timeoutDuration,
		TimeoutPolicy:   timeoutPolicy,
	}
	if err := k.InterchainQueryKeeper.SubmitICQRequest(ctx, query, true); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting ICQ for validator sharesToTokens rate, error %s", err.Error()))
		return err
	}
	return nil
}
