// x/stakeibc/keeper/wind_down_staketia_claim.go
package keeper

import (
	transfertypes "github.com/cosmos/ibc-go/v11/modules/apps/transfer/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v34/utils"
	"github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiatypes "github.com/Stride-Labs/stride/v34/x/staketia/types"
)

// TransferStaketiaClaimBalance moves TIA vouchers from staketia's claim address (a multisig
// BaseAccount whose signers are not on the critical path) to the stakeibc celestia zone's
// delegation ICA, where they unwind to native TIA and leave with the ICA balance (spec §7).
// The keeper signs for the claim address the way staketia's keeper already does for its
// deposit address. amount zero means the whole balance; a timeout refunds the claim address.
func (k Keeper) TransferStaketiaClaimBalance(ctx sdk.Context, msg *types.MsgTransferStaketiaClaimBalance) (sdk.Coin, error) {
	hostZone, found := k.GetHostZone(ctx, staketiatypes.CelestiaChainId)
	if !found {
		return sdk.Coin{}, types.ErrHostZoneNotFound.Wrapf("host zone %s not found", staketiatypes.CelestiaChainId)
	}
	if hostZone.DelegationIcaAddress == "" {
		return sdk.Coin{}, types.ErrICAAccountNotFound.Wrapf("delegation ICA for %s has no address", staketiatypes.CelestiaChainId)
	}
	if hostZone.TransferChannelId == "" {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "host zone %s has no transfer channel", staketiatypes.CelestiaChainId)
	}

	claimAddress, err := sdk.AccAddressFromBech32(staketiatypes.ClaimAddress)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(err, "invalid staketia claim address constant")
	}
	balance := k.bankKeeper.GetBalance(ctx, claimAddress, staketiatypes.CelestiaNativeTokenIBCDenom)
	if balance.IsZero() {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "claim address %s holds no %s", staketiatypes.ClaimAddress, staketiatypes.CelestiaNativeTokenIBCDenom)
	}

	amount := msg.Amount
	if amount.IsNil() || amount.IsZero() {
		amount = balance.Amount
	}
	if amount.GT(balance.Amount) {
		return sdk.Coin{}, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "requested %v but the claim address holds %v", amount, balance.Amount)
	}
	token := sdk.NewCoin(balance.Denom, amount)

	timeoutTimestamp := utils.IntToUint(ctx.BlockTime().Add(types.WindDownTransferTimeout).UnixNano())
	transferMsg := BuildStaketiaClaimTransferMsg(hostZone, token, timeoutTimestamp)
	if _, err := k.RecordsKeeper.TransferKeeper.Transfer(ctx, &transferMsg); err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(err, "unable to transfer %v from the staketia claim address", token)
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(staketiatypes.CelestiaChainId,
		"Wind-down transfer of %v from the staketia claim address to %s over %s", token, hostZone.DelegationIcaAddress, hostZone.TransferChannelId))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransferStaketiaClaimBalance,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeKeyAmount, token.String()),
			sdk.NewAttribute(types.AttributeKeyChannel, hostZone.TransferChannelId),
			sdk.NewAttribute(types.AttributeKeyReceiver, hostZone.DelegationIcaAddress),
		),
	)
	return token, nil
}

// BuildStaketiaClaimTransferMsg is the ICS-20 transfer from the claim address to the celestia
// delegation ICA. Exported so tests can assert every field.
func BuildStaketiaClaimTransferMsg(hostZone types.HostZone, token sdk.Coin, timeoutTimestamp uint64) transfertypes.MsgTransfer {
	return transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    hostZone.TransferChannelId,
		Token:            token,
		Sender:           staketiatypes.ClaimAddress,
		Receiver:         hostZone.DelegationIcaAddress,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             "",
	}
}
