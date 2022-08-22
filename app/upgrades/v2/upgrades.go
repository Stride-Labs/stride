package v2

import (
	"fmt"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	recordsmodulekeeper "github.com/Stride-Labs/stride/x/records/keeper"
	stakeibcmodulekeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/x/stakeibc/types"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
)

const (
	UpgradeName = "Upgrade to Resolve Consensus Bug"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	IBCKeeper			*ibckeeper.Keeper,
	RecordsKeeper  		recordsmodulekeeper.Keeper,
	StakeibcKeeper		stakeibcmodulekeeper.Keeper,
	BankKeeper          bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm[recordstypes.ModuleName] = 2

		// denoms
		atomDenom := "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
		osmoDenom := "ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34"

		// module accounts
		osmo, found := StakeibcKeeper.GetHostZone(ctx, "OSMO")
		if !found {
			return nil, fmt.Errorf("could not find osmo")
		}
		osmoModuleAddr := osmo.Address
		if osmoModuleAddr == "" {
			return nil, fmt.Errorf("could not find osmo address")
		}
		osmoModuleAcc := sdk.AccAddress(osmoModuleAddr)

		gaia, found := StakeibcKeeper.GetHostZone(ctx, "GAIA")
		if !found {
			return nil, fmt.Errorf("could not find gaia")
		}
		gaiaModuleAddr := gaia.Address
		if gaiaModuleAddr == "" {
			return nil, fmt.Errorf("could not find gaia address")
		}
		gaiaModuleAcc := sdk.AccAddress(gaiaModuleAddr) 

		// TODO: if we update a packet commitment for an in-flight packet, will it cause the ack to fail?
		commitments := IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, "transfer", "channel-5")
		for _, commitment := range commitments {
			// STEP 1: clear packets
			_ = commitment
			srcPort := commitment.PortId
			_ = srcPort
			srcChannel := commitment.ChannelId
			_ = srcChannel
			sequence := commitment.Sequence
			_ = sequence
			// I think this is an ack, but low confidence
			ack := commitment.Data
			_ = ack
			// commitment := channeltypes.CommitPacket(IBCKeeper.ChannelKeeper.ModuleCdc, packet)
			// IBCKeeper.ChannelKeeper.SetPacketCommitment(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence(), commitment)
			// unclear how we can clear packets here, deletePacketCommitment is not exposed in the SDK

			// STEP 2: refund tokens
			denom := osmoDenom
			amount := "1" // TODO: get amount from packet
			trace := ibctransfertypes.ParseDenomTrace(denom)
			transferAmount, ok := sdk.NewIntFromString(amount)
			if !ok {
				panic("unable to parse transfer amount into sdk.Int")
			}
			token := sdk.NewCoin(trace.IBCDenom(), transferAmount)			
			if err := BankKeeper.MintCoins(
				ctx, ibctransfertypes.ModuleName, sdk.NewCoins(token),
			); err != nil {
				panic(err)
			}
			if err := BankKeeper.SendCoinsFromModuleToAccount(ctx, ibctransfertypes.ModuleName, osmoModuleAcc, sdk.NewCoins(token)); err != nil {
				panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
			}
		}

		// STEP 3: transfer tokens to the new zone module accounts
		// transfer tokens from stakeibc's module account to the zone accounts
		osmoAmount := BankKeeper.GetBalance(ctx, osmoModuleAcc, osmoDenom)
		if err := BankKeeper.SendCoinsFromModuleToAccount(ctx, stakeibctypes.ModuleName, osmoModuleAcc, sdk.NewCoins(osmoAmount)); err != nil {
			panic(fmt.Sprintf("unable to send coins to osmo module account: %v", err))
		}

		// transfer tokens from stakeibc's module account to the zone accounts
		atomAmount := BankKeeper.GetBalance(ctx, gaiaModuleAcc, atomDenom)
		if err := BankKeeper.SendCoinsFromModuleToAccount(ctx, stakeibctypes.ModuleName, gaiaModuleAcc, sdk.NewCoins(atomAmount)); err != nil {
			panic(fmt.Sprintf("unable to send coins to gaia module account: %v", err))
		}
		

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
