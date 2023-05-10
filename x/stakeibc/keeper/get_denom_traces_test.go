package keeper_test

// Note: this is for dockernet

import (
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
)

func (s *KeeperTestSuite) TestIBCDenom() {
	validators := []string{
		"cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p",
		"cosmosvaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rarvcqz",
		"cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n",
		"cosmosvaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttr0ks75",
	}
	channels := []string{
		"0",
		"1",
		"2",
	}
	recordIds := []string{
		"1", "2", "3", "4", "5",
	}
	for _, channel := range channels {
		fmt.Printf("## Channel-%s\n", channel)
		for i := 0; i < 4; i++ {
			validatorId := fmt.Sprintf("%d", i+1)
			address := validators[i]

			fmt.Printf("# Val %s\n", validatorId)

			for _, record := range recordIds {
				sourcePrefix := transfertypes.GetDenomPrefix("transfer", "channel-"+channel)
				prefixedDenom := sourcePrefix + address + "/" + record

				fmt.Printf("LSM_TOKEN_DENOM_VAL_%s_RECORD_ID_%s_CHANNEL_%s='%s'\n", validatorId, record, channel, transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom())
			}
			fmt.Println()
		}
	}
}
