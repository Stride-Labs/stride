package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func SendCoinsExample(ctx sdk.Context, k bankkeeper.Keeper, msg *types.AddValidatorProposal) error {
	fmt.Println("ABOUT TO SEND COINS")

	fromAddr, _ := sdk.AccAddressFromBech32("stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7")
	toAddr, _ := sdk.AccAddressFromBech32("stride17kht2x2ped6qytr2kklevtvmxpw7wq9rmuc3ca")
	err := k.SendCoins(ctx, fromAddr, toAddr, sdk.NewCoins(sdk.NewInt64Coin("ustrd", int64(6789))))
	fmt.Println("ERR:", err)

	fmt.Println("SENT COINS")

	return nil
}

func AddHostZoneExample(ctx sdk.Context, k stakeibckeeper.Keeper, msg *types.AddValidatorProposal) error {
	fmt.Println("ABOUT TO ADD HOST ZONE")

	k.SetHostZone(ctx, types.HostZone{
		ChainId:   "FAKE",
		HostDenom: "fake",
	})

	fmt.Println("DONE ADDING HOST ZONE")

	return nil
}
