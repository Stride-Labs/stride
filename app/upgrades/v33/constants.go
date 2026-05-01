package v33

import (
	_ "embed"
	"encoding/json"
)

// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
const UpgradeName = "v33"

// REHEARSAL ONLY — DO NOT MERGE
// AdminMultisigAddress points at the integration-tests `admin` account so the
// rehearsal can drive POA admin behavior with a known signer. The mainnet
// multisig address lives on the `main` branch.
const AdminMultisigAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"

// ExpectedValidatorCount is enforced by the upgrade handler. Panics if
// consumerKeeper.GetAllCCValidator returns a different count.
const ExpectedValidatorCount = 8

//go:embed validators.json
var validatorsJSON []byte

// ValidatorMonikers maps lowercase hex of the consensus address (the raw 20-byte
// address returned by ccv consumer GetAllCCValidator) to the validator's moniker.
// ICS does not store monikers on the consumer chain — they live on the Hub.
// We pre-bake them so they appear correctly in POA's validator records.
//
// Populated from validators.json at init. Regenerate with scripts/fetch_validator_monikers.sh.
var ValidatorMonikers = func() map[string]string {
	m := make(map[string]string)
	if err := json.Unmarshal(validatorsJSON, &m); err != nil {
		panic("v33: failed to parse embedded validators.json: " + err.Error())
	}
	if len(m) != ExpectedValidatorCount {
		panic("v33: validators.json has wrong validator count")
	}
	return m
}()
