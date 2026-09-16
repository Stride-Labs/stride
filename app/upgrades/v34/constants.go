package v34

import "time"

const (
	// UpgradeName is the SDK upgrade plan name. Match the binary release tag.
	UpgradeName = "v34"

	// ValidatorPower matches the uniform power of the existing POA set — the
	// value snapshotted from ICS at the v33 migration. POA only weighs relative
	// power, so incoming validators join at the same weight and total power is
	// unchanged by the swap.
	ValidatorPower = int64(274523)

	// Gov params set by this upgrade (previously a 33.4% quorum and 3 day voting period)
	GovQuorum       = "0.250000000000000000"
	GovVotingPeriod = 5 * 24 * time.Hour
)

// IncomingValidator identifies a validator added to the POA set by this
// upgrade. The payout/operator address deliberately does NOT live here — the
// handler joins it from utils.PoaValidatorSet by moniker so the address has a
// single source of truth.
type IncomingValidator struct {
	Moniker          string
	ConsPubKeyBase64 string // base64-encoded ed25519 consensus pubkey
}

// vars rather than consts so tests can substitute generated test keys.
var (
	// Pubkeys confirmed in writing by each validator (2026-09-11) from the
	// `key` field of `strided tendermint show-validator` on their running
	// nodes. Payout addresses live in utils.PoaValidatorSet.
	// REHEARSAL ONLY: the k8s network's val4/val5 nodes (outside the genesis POA set); their
	// consensus pubkeys are generated at network init and pasted here by `rehearsal/poa.sh measure`
	IncomingValidators = []IncomingValidator{
		{Moniker: "val4", ConsPubKeyBase64: "VknSviQ3ja00Ck3Rf19ikQX29Hfk1dVWq/hMx4ikOjk="},
		{Moniker: "val5", ConsPubKeyBase64: "pPrponkDVY72CLiWY9glxtGsNKenX5PAxYS573ajfY0="},
	}

	// OutgoingMonikers are resolved against live POA state at upgrade time —
	// no hand-transcribed consensus addresses to typo.
	// REHEARSAL ONLY: two of the three genesis validators
	OutgoingMonikers = []string{"val2", "val3"}
)

// Stuck ICQ cleanup, measured against mainnet at height 40276875 (2026-09-14).
// Anything that gets stuck after that measurement is left for a later upgrade.
const StuckSlashQueryHostZone = "cosmoshub-4"

var (
	// Validators with SlashQueryInProgress=true but no delegation ICQ in the
	// store. Each has zero delegation, so its delegation ICQ response came back
	// contentless and SubmitQueryResponse deleted the query without invoking
	// the callback that clears the flag.
	StuckSlashQueryValidators = []string{
		"whispernode",
		"kira",
		"onblocnode",
		"pupmos",
		"kraken",
		"gatadao",
		"cosvalidator",
		"klubstaking",
		"huobi",
		"validatornetwork",
		"zoomerlabs",
		"ethicalnode",
	}

	// ICQs past their timeout that never received a response (mostly on halted
	// host zones). Queries that were still within their timeout at measurement
	// time are excluded.
	StuckQueryIds = []string{
		"f0e4f2008f5d83014532568a91549634c8ad54756648ca8370d47d21e4ca0c89", // celestia validator
		"f64cf95c2a063e1ac126ebabdc75fd2f97e1d3bcedc53c34dddf394e92aa3f17", // comdex-1 calibrate
		"8ca0f4abbaf2c794bd4d3be29d71e0a4021bf7d33f6d9cbc8c7dcd6cce4987bf", // comdex-1 calibrate
		"4799692f78acda3ebbedae6189c4a4277bbb0358196af232f2e6850581777321", // comdex-1 calibrate
		"99a956729b8dba88e035dd98232335fcee633db7810ea9ce8529f0f5e6b0aaf7", // comdex-1 calibrate
		"adec3d47d3cfead11309bb18bb9aa24a85cc8b9268c8a04c868b234ef387d3ad", // comdex-1 calibrate
		"f8936e7108f38eb2c6f5939446e89145c475394770aef4905d72053c984f33f5", // comdex-1 calibrate
		"619dfcfa20c14c0e26c81ddd2561e33a11374cc28da4ff7cdb9c73ed2c1c72d7", // evmos_9001-2 validator
		"73c608db2aa4e12a62ef5787a8daf977c05bf0b8c70b2061710ffddf5432dcba", // evmos_9001-2 validator
		"cbfc7d30483175c454d71c838c93a09a8700ec77e7865a37738d966d09698091", // evmos_9001-2 validator
		"38215b337b42fd14fe0bac4640d9905c8a53c1aed78afbc473f2b5adaede8180", // evmos_9001-2 validator
		"96498060cc05a0caef479491e9eef4bfafc04892edcf60b486d94709ca965933", // evmos_9001-2 validator
		"72174a14d19e32cef74f8d67e5d74c960ff87fc62a31a2e4f758ccba2123aa4d", // evmos_9001-2 validator
		"0903699e4e46d035055f1d4526ff7ea3fbba6434a946d2550980c31d26f0ccb3", // evmos_9001-2 validator
		"daf3f0a291a769a4cfbd430f8238bfe9e12481ef499176dbe81acc15c60dd877", // evmos_9001-2 validator
		"5e1978ae4b559a3deaeabd3229fba997eee591819d6bbcf2d567b0046c060c88", // evmos_9001-2 validator
		"a4d834753d822cb49a5c598a30b6dcbd3b9a883b15693055b9901da7a62aa2c2", // evmos_9001-2 validator
		"5370b571e8ef050ec123a06d37706966669ba0b38f39d0d33fdce13c9cb9708b", // evmos_9001-2 validator
		"05debee690941130f3e3e2e3e37beb502339cf7fd5d9916ef1edbc59e09b3d8e", // evmos_9001-2 validator
		"1884dc7763df8f13d40cafd1c5c774085b53f07ac5ecf699f341b63987153bf3", // evmos_9001-2 validator
		"875bdd6669a8eaf27dc9164e6f013dee2270dd80914b2599dcd654478071cdca", // evmos_9001-2 validator
		"56e0505819b5121699c206aa3dfcd72d9ec9b271f2b2ed05d20529db4a81dd55", // evmos_9001-2 validator
		"35a0246ed20d6e65542d15ca128d73a5974ccc75ff591eafcdc6785bed73475c", // evmos_9001-2 validator
		"76d99f16e1551cff451eb31ea9e5078827303f79ad976f7b81e82a678013c474", // evmos_9001-2 validator
		"5eac2ed2840edde75ca6d2bd314fb14a1cc1df6415a92067d094c2e193b2f142", // evmos_9001-2 validator
		"018d930fdb176d8e952493c7e86cb604831e280a0c6eca111dc2530bfb6292ce", // evmos_9001-2 withdrawalbalance
		"310e40e6408431c4b05a9984354ba788d06734c87b673008826d7ea4019b7a22", // stargaze-1 withdrawalbalance
		"b39fbded310381fb2227f18b072cc9faebb0d1d6f4f64f7c8ba67fb20206fad7", // umee-1 calibrate
		"42cb7222cd5cd45cf3e9420036ad999a815a6878e72eecff8e18818669208ae5", // umee-1 validator
		"ff6dc0b79a589e623f31a718c316f594f10607e0e307cb92db2d60024020d7e3", // umee-1 withdrawalbalance
	}
)
