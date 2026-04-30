package v33

import (
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"

	ccvconsumerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/consumer/keeper"
)

// SnapshotValidatorsFromICS reads the current CCV validator set from the
// consumer keeper and converts it into a slice of POA Validators ready to
// be passed to poaKeeper.InitGenesis.
//
// Monikers are looked up from the pre-baked ValidatorMonikers map (keyed by
// hex of the consensus address). Missing entries fall back to empty string.
//
// Operator addresses are derived from the consensus address bytes by
// bech32-encoding with the "stridevaloper" prefix. Stride's ICS validators
// have no Stride-side operator address (the operator key lives on the Hub),
// so this is a metadata-only string — POA's runtime logic keys off the
// consensus address, not OperatorAddress.
func SnapshotValidatorsFromICS(
	ctx sdk.Context,
	consumerKeeper ccvconsumerkeeper.Keeper,
) ([]poatypes.Validator, error) {
	ccVals := consumerKeeper.GetAllCCValidator(ctx)
	if len(ccVals) != ExpectedValidatorCount {
		return nil, fmt.Errorf(
			"expected %d validators in consumer keeper, got %d",
			ExpectedValidatorCount, len(ccVals),
		)
	}

	poaVals := make([]poatypes.Validator, 0, len(ccVals))
	for _, ccVal := range ccVals {
		consPubKey, err := ccVal.ConsPubKey()
		if err != nil {
			return nil, errorsmod.Wrapf(err,
				"failed to decode cons pubkey for validator %x", ccVal.Address)
		}
		pubKeyAny, err := codectypes.NewAnyWithValue(consPubKey)
		if err != nil {
			return nil, errorsmod.Wrapf(err,
				"failed to wrap cons pubkey for validator %x", ccVal.Address)
		}

		operatorAddr, err := bech32.ConvertAndEncode("stridevaloper", ccVal.Address)
		if err != nil {
			return nil, errorsmod.Wrapf(err,
				"failed to bech32-encode operator address for validator %x", ccVal.Address)
		}

		moniker := ValidatorMonikers[hex.EncodeToString(ccVal.Address)]

		poaVals = append(poaVals, poatypes.Validator{
			PubKey: pubKeyAny,
			Power:  ccVal.Power,
			Metadata: &poatypes.ValidatorMetadata{
				Moniker:         moniker,
				OperatorAddress: operatorAddr,
			},
		})
	}

	return poaVals, nil
}
