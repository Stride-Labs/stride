package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	//"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v14/utils"
	//recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// ibc transfers tokens from the foreign hub community pool deposit ICA address onto Stride hub
func (k Keeper) IBCTransferCommunityPoolICATokensToStride(ctx sdk.Context, communityPoolHostZoneId string, token sdk.Coin) error {
	k.Logger(ctx).Info(fmt.Sprintf("Transfering %d %s tokens from community pool deposit ICA to Stride hub holding address", token.Amount.Int64(), token.Denom))

	// TODO: add safety check here on if the amount is greater than a threshold to avoid many, frequent, small transfers
	//       threshold might be a config we tune to avoid ddos type attacks, for now using 0 as hard coded threshold
	if token.Amount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] The amount %v to transfer did not meet the minimum threshold!", token.Amount))
		return errors.New("Transfer Amount below threshold!")
	}

	// TODO: add check that the denom is something we know how to handle -- here is where a token whitelist would filter for coin types


	hostZone, hostZoneFound := k.GetHostZone(ctx, communityPoolHostZoneId)
	if !hostZoneFound || hostZone.Halted {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Host zone not found or halted!"))
		return errors.New("No active host zone found!")
	}

	if hostZone.CommunityPoolDepositIcaAddress == "" || hostZone.CommunityPoolHoldingAddress == "" {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Unknown send or recieve address! DepositICAAddress: %s HoldingAddress: %s", 
			hostZone.CommunityPoolDepositIcaAddress, 
			hostZone.CommunityPoolHoldingAddress))
		return errors.New("Critical addresses missing from hostZone config!")
	}


	k.Logger(ctx).Info(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Transferring %v%s", token.Amount, token.Denom))
	
	// NOTE: this assumes no clock drift between chains, which tendermint guarantees
	ibcTransferTimeoutNanos := k.GetParam(ctx, types.KeyIBCTransferTimeoutNanos)
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + ibcTransferTimeoutNanos

	msg := ibctypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		hostZone.TransferChannelId,
		token,
		hostZone.CommunityPoolDepositIcaAddress, // ICA controlled address on foreign hub
		hostZone.CommunityPoolHoldingAddress, // Stride address, unique to each community pool / hostzone
		clienttypes.Height{},
		timeoutTimestamp,
		"Sweep stake-able tokens from community pool deposit address to Stride chain",
	)

	// ibc transfer tokens from foreign hub deposit ICA address to Stride hub "holding" address
	err := k.IBCTransferCommunityPoolTokens(ctx, msg)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] Failed to initiate IBC transfer to host zone, HostZone: %v, Channel: %v, Coin: %v, SendAddress: %v, RecAddress: %v, Timeout: %v",
			hostZone.ChainId, hostZone.TransferChannelId, token, hostZone.CommunityPoolDepositIcaAddress, hostZone.CommunityPoolHoldingAddress, timeoutTimestamp))
		k.Logger(ctx).Error(fmt.Sprintf("[IBCTransferCommunityPoolICATokensToStride] err {%s}", err.Error()))
		return errors.New("")
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "[IBCTransferCommunityPoolICATokensToStride] Successfully submitted transfer for community pool tokens"))

	return nil
}

// Calls liquid stake on all stake-able tokens in the community pool holding address, stTokens come back to holding address
func (k Keeper) LiquidStakeCommunityPoolTokens(ctx sdk.Context, communityPoolHoldingAddress string) error {

	return nil
}
