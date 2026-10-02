"""Unit tests for build_sweep_batches over a synthetic export.

    python3 -m unittest scripts/wind-down/test_build_sweep_batches.py
"""

import argparse
import contextlib
import io
import json
import pathlib
import re
import tempfile
import unittest
from unittest import mock

import bech32_ref
import build_sweep_batches

STRIDE_BASE = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
STRIDE_VESTING = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
STRIDE_MODULE = "stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan"  # the distribution module account: sha256("distribution")[:20]
STRIDE_CONTINUOUS = "stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c"  # ContinuousVestingAccount, 40% through its schedule
STRIDE_BLOCKED_BASE = "stride1j4yzhgjm00ch3h0p9kel7g8sp6g045qfcgk6ex"  # sha256("auction")[:20] held by a plain BaseAccount
STRIDE_ICA = "stride1zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zkh7xr"
STRIDE_NO_ACCOUNT = "stride1yg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zdtjcw5"
STRIDE_DUST = "stride1xvenxvenxvenxvenxvenxvenxvenxvenl499mx"
ATOM_VOUCHER = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"


# The export's genesis_time is the chain's original genesis, years before the export height: the
# builder must never use it for vesting, only --as-of
GENESIS_TIME = "2020-01-01T00:00:00Z"
AS_OF = 1790640000  # 2026-09-29T00:00:00Z, the block time at the export height
SWEEP_OPERATOR = "stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy"
PROTOCOL_ADDRESSES = [
    "stride1ju3xt2f8xuhzxqg6590sazctlz6l4md0wc5w6c",  # staketia deposit
    "stride19ksqv50zmntzjfflfmnegj75tdfkk89vl2q5yu",  # staketia redemption
    "stride1pjw24gg0fm26758hxee3wta35kq9jpszcslm6z",  # staketia claim
    "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj",  # stakedym deposit
    "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n",  # stakedym redemption
    "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras",  # stakedym claim
    "stride1tpzfseenwg4kq54sf9hdp3mkra652fvqtsuclq",  # staketia safe
    "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7",  # stakedym safe
    "stride19xm04qaah8t2eupyeglz63vkaxzytpyc8m7kk4",  # staketia operator
]
REPO_ROOT = pathlib.Path(__file__).resolve().parents[2]


def synthetic_export() -> dict:
    escrow = build_sweep_batches.escrow_address("transfer", "channel-5")
    contract = build_sweep_batches.escrow_address("transfer", "channel-999")  # any 20-byte address not otherwise used
    balances = [
        {"address": STRIDE_BASE, "coins": [{"denom": "stuatom", "amount": "10000000"}, {"denom": "ustrd", "amount": "5000000"}]},
        {"address": STRIDE_VESTING, "coins": [{"denom": "stuatom", "amount": "2000000"}]},
        {"address": STRIDE_CONTINUOUS, "coins": [{"denom": "stuatom", "amount": "1000000"}, {"denom": "ustrd", "amount": "1000000"}]},
        {"address": contract, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_MODULE, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_ICA, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_NO_ACCOUNT, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": escrow, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": STRIDE_DUST, "coins": [{"denom": "stuatom", "amount": "1000"}]},
        {"address": STRIDE_BLOCKED_BASE, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
        {"address": SWEEP_OPERATOR, "coins": [{"denom": "stuatom", "amount": "99000000"}]},
    ] + [{"address": address, "coins": [{"denom": "stuatom", "amount": "99000000"}]} for address in PROTOCOL_ADDRESSES]
    accounts = [
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_BASE, "pub_key": {"@type": "/cosmos.crypto.secp256k1.PubKey", "key": "AAAA"}, "sequence": "4"},
        {"@type": "/stride.vesting.StridePeriodicVestingAccount", "base_vesting_account": {"base_account": {"address": STRIDE_VESTING}, "original_vesting": [], "delegated_vesting": [], "end_time": "0"}, "vesting_periods": []},
        # 1,000,000 ustrd vesting linearly from AS_OF-400 to AS_OF+600: 400,000 vested, 600,000 locked at the export
        {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount", "base_vesting_account": {"base_account": {"address": STRIDE_CONTINUOUS, "sequence": "3"}, "original_vesting": [{"denom": "ustrd", "amount": "1000000"}], "delegated_vesting": [], "end_time": str(AS_OF + 600)}, "start_time": str(AS_OF - 400)},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": contract},
        {"@type": "/cosmos.auth.v1beta1.ModuleAccount", "base_account": {"address": STRIDE_MODULE}, "name": "distribution"},
        {"@type": "/ibc.applications.interchain_accounts.v1.InterchainAccount", "base_account": {"address": STRIDE_ICA}, "account_owner": "x"},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_DUST},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": escrow},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_BLOCKED_BASE},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": SWEEP_OPERATOR},
    ] + [{"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": address} for address in PROTOCOL_ADDRESSES]
    channels = [{"port_id": "transfer", "channel_id": "channel-5", "state": "STATE_OPEN"}]
    contracts = [{"contract_address": contract, "contract_info": {"code_id": "1"}}]
    return {
        "genesis_time": GENESIS_TIME,
        "app_state": {
            "bank": {"balances": balances},
            "auth": {"accounts": accounts},
            "ibc": {"channel_genesis": {"channels": channels}},
            "wasm": {"contracts": contracts},
            "stakeibc": {"host_zone_list": [
                {"chain_id": "cosmoshub-4", "host_denom": "uatom"},
                {"chain_id": "osmosis-1", "host_denom": "uosmo", "deprecated": True},
            ]},
        }
    }


PRICES = {
    "stuatom": {"usd_per_token": 10.0, "decimals": 6},
    "ustrd": {"usd_per_token": 0.05, "decimals": 6},
}


class BuildSweepBatchesTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.export_path = pathlib.Path(self.tmp.name) / "export.json"
        self.export_path.write_text(json.dumps(synthetic_export()))
        self.export = build_sweep_batches.load_export(self.export_path, as_of=AS_OF)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_classify_applies_skip_rules_and_floor(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0, sweep_operator=SWEEP_OPERATOR)

        swept = [holder.address for holder in plan.holders]
        self.assertEqual(swept, [STRIDE_BASE, STRIDE_VESTING, STRIDE_CONTINUOUS],
                         "ordered by USD: base ($100.25), stride vesting ($20), continuous ($10.02 on its spendable part)")

        skipped = {entry.address: entry.reason for entry in plan.skipped}
        self.assertEqual(skipped[STRIDE_MODULE], "blocked module address")
        self.assertEqual(skipped[build_sweep_batches.escrow_address("transfer", "channel-999")], "wasm contract address")
        self.assertEqual(skipped[STRIDE_ICA], "interchain account")
        self.assertEqual(skipped[STRIDE_NO_ACCOUNT], "account not found")
        self.assertEqual(skipped[build_sweep_batches.escrow_address("transfer", "channel-5")], "transfer escrow address")
        self.assertEqual(skipped[STRIDE_DUST], "below floor ($0.01 < $1.00)")

    def test_write_batches_splits_at_batch_size(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0, sweep_operator=SWEEP_OPERATOR)
        out_dir = pathlib.Path(self.tmp.name) / "out"

        files = build_sweep_batches.write_batches(plan=plan, out_dir=out_dir, batch_size=1)

        self.assertEqual([path.name for path in files], ["batch-001.txt", "batch-002.txt", "batch-003.txt"])
        self.assertEqual(files[0].read_text().strip(), STRIDE_BASE)
        summary = json.loads((out_dir / "summary.json").read_text())
        self.assertEqual(summary["denoms"], ["stuatom", "ustrd"])
        self.assertEqual(summary["batches"][0]["num_addresses"], 1)
        self.assertEqual(len(summary["skipped"]), 6 + 2 + len(PROTOCOL_ADDRESSES))

    def test_vesting_locked_balance_is_not_swept(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)
        continuous = next(holder for holder in plan.holders if holder.address == STRIDE_CONTINUOUS)

        # 1,000,000 ustrd held, 600,000 still locked at --as-of: the chain's SpendableCoin is 400,000
        self.assertEqual(continuous.balances, {"stuatom": 1000000, "ustrd": 400000})
        self.assertEqual(f"{continuous.usd:.2f}", "10.02")

        schedule = self.export.vesting[STRIDE_CONTINUOUS]
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF), {"ustrd": 600000})
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF + 600), {})
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF - 400), {"ustrd": 1000000})

    def test_genesis_time_years_before_as_of_does_not_affect_vesting(self) -> None:
        # The fixture's genesis_time is 2020; the continuous account is 40% vested at AS_OF only if
        # the builder measures from --as-of. Measured from genesis_time everything would be vested
        self.assertEqual(self.export.as_of, AS_OF)
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["ustrd"], prices=PRICES, floor_usd=0.0)
        continuous = next(holder for holder in plan.holders if holder.address == STRIDE_CONTINUOUS)
        self.assertEqual(continuous.balances, {"ustrd": 400000})

    def test_as_of_accepts_iso_and_unix_seconds(self) -> None:
        self.assertEqual(build_sweep_batches.as_of_arg("2026-09-29T00:00:00Z"), AS_OF)
        self.assertEqual(build_sweep_batches.as_of_arg("2026-09-29T00:00:00.123456789Z"), AS_OF)
        self.assertEqual(build_sweep_batches.as_of_arg("2026-09-29T02:00:00+02:00"), AS_OF)
        self.assertEqual(build_sweep_batches.as_of_arg(str(AS_OF)), AS_OF)
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.as_of_arg("yesterday")

    def test_as_of_is_required(self) -> None:
        with self.assertRaises(SystemExit), contextlib.redirect_stderr(io.StringIO()):
            with mock.patch("sys.argv", ["x", "--export", "e", "--prices", "p", "--denoms", "ustrd", "--floor-usd", "1", "--out-dir", "o"]):
                build_sweep_batches.parse_args()

    def test_every_protocol_address_and_the_operator_is_excluded(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0, sweep_operator=SWEEP_OPERATOR)
        skipped = {entry.address: entry.reason for entry in plan.skipped}
        swept = {holder.address for holder in plan.holders}
        for address in PROTOCOL_ADDRESSES + [SWEEP_OPERATOR]:
            self.assertEqual(skipped[address], "protocol address", address)
            self.assertNotIn(address, swept)
        self.assertIn(STRIDE_BASE, swept, "a normal holder in the same run is still swept")

        # Without --sweep-operator the operator is an ordinary holder; the multisigs stay excluded
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0)
        self.assertIn(SWEEP_OPERATOR, {holder.address for holder in plan.holders})
        self.assertEqual(len([e for e in plan.skipped if e.reason == "protocol address"]), len(PROTOCOL_ADDRESSES))

    def test_blocked_module_addresses_are_excluded(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0)
        skipped = {entry.address: entry.reason for entry in plan.skipped}
        # A blocked name from the app's list, held by a BaseAccount and absent from the export's module accounts
        self.assertEqual(skipped[STRIDE_BLOCKED_BASE], "blocked module address")
        # A module account present in the export
        self.assertEqual(skipped[STRIDE_MODULE], "blocked module address")

        # A module account whose name is not in the constant list is still excluded via the export
        custom = build_sweep_batches.module_address("custom_module")
        self.export.module_addresses.add(custom)
        self.export.balances[custom] = {"stuatom": 99000000}
        self.export.account_types[custom] = "/cosmos.auth.v1beta1.BaseAccount"
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0)
        self.assertIn((custom, "blocked module address"), [(e.address, e.reason) for e in plan.skipped])

    def test_extra_denom_parsing_and_whitelist(self) -> None:
        extra = build_sweep_batches.parse_extra_denom("ustrd=0.05:6")
        self.assertEqual((extra.denom, extra.usd_per_token, extra.decimals), ("ustrd", 0.05, 6))

        with self.assertRaises(ValueError):
            build_sweep_batches.parse_extra_denom("ustrd=0.05")

    def test_ibc_denom_requires_whitelisted_outer_hop_however_it_is_passed(self) -> None:
        traces = {ATOM_VOUCHER: ["transfer/channel-0"], "ibc/AAAA": ["transfer/channel-52"]}
        allowed = self.export.allowed_native_denoms
        build_sweep_batches.check_denom_destination(denom="stuatom", traces=traces, allowed_native=allowed)
        build_sweep_batches.check_denom_destination(denom=ATOM_VOUCHER, traces=traces, allowed_native=allowed)

        # The same check guards a voucher given through --denoms, not only --extra-denom
        with self.assertRaises(ValueError) as raised:
            build_sweep_batches.check_denom_destination(denom="ibc/AAAA", traces=traces, allowed_native=allowed)
        self.assertIn("ibc/AAAA", str(raised.exception))
        self.assertIn("channel-52", str(raised.exception))

        with self.assertRaises(ValueError):
            build_sweep_batches.check_denom_destination(denom="ibc/BBBB", traces=traces, allowed_native=allowed)  # no trace at all

        # The outermost hop must be on the transfer port
        with self.assertRaises(ValueError) as raised:
            build_sweep_batches.check_denom_destination(denom="ibc/CCCC", traces={"ibc/CCCC": ["wasm.contract/channel-0"]}, allowed_native=allowed)
        self.assertIn("wasm.contract", str(raised.exception))

    def test_native_denoms_are_limited_to_the_allow_list(self) -> None:
        # ustrd, stutia and the stToken of each non-deprecated host zone; not stadym, a deprecated
        # zone's stToken, or a host zone's own denom
        self.assertEqual(self.export.allowed_native_denoms, {"ustrd", "stutia", "stuatom"})
        for denom in ["ustrd", "stutia", "stuatom"]:
            build_sweep_batches.check_denom_destination(denom=denom, traces={}, allowed_native=self.export.allowed_native_denoms)
        for denom in ["stadym", "stuosmo", "uatom", "ufake"]:
            with self.assertRaises(ValueError, msg=denom) as raised:
                build_sweep_batches.check_denom_destination(denom=denom, traces={}, allowed_native=self.export.allowed_native_denoms)
            self.assertIn(denom, str(raised.exception))

    def test_sweep_operator_defaults_to_the_spec_address(self) -> None:
        argv = ["x", "--export", "e", "--prices", "p", "--denoms", "ustrd", "--floor-usd", "1", "--out-dir", "o", "--as-of", "0"]
        with mock.patch("sys.argv", argv):
            self.assertEqual(build_sweep_batches.parse_args().sweep_operator, "stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy")

    def test_keyless_candidates_are_listed_for_review(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0, sweep_operator=SWEEP_OPERATOR)
        # Only the stride vesting account passes every rule with no pubkey and sequence 0: the base
        # account has a key and the continuous one has signed. Skipped holders are never listed
        self.assertEqual([holder.address for holder in plan.keyless_candidates], [STRIDE_VESTING])

        out_dir = pathlib.Path(self.tmp.name) / "out"
        build_sweep_batches.write_batches(plan=plan, out_dir=out_dir, batch_size=100)
        self.assertEqual((out_dir / "keyless_candidates.txt").read_text(), f"{STRIDE_VESTING} $20.00\n")
        self.assertEqual(json.loads((out_dir / "summary.json").read_text())["num_keyless_candidates"], 1)

    def test_batch_size_has_no_upper_bound(self) -> None:
        self.assertEqual(build_sweep_batches.batch_size_arg("100"), 100)
        self.assertEqual(build_sweep_batches.batch_size_arg("1"), 1)
        self.assertEqual(build_sweep_batches.batch_size_arg("101"), 101)
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.batch_size_arg("0")

    def test_protocol_addresses_mirror_the_go_constants(self) -> None:
        source = (REPO_ROOT / "x/stakeibc/types/wind_down.go").read_text()
        constants = re.findall(r'^\s*(?:Staketia|Stakedym)\w+Address\s*=\s*"(stride1\w+)"', source, re.MULTILINE)
        self.assertEqual(len(constants), len(set(constants)), "each Go constant is a distinct address")
        self.assertEqual(set(constants), build_sweep_batches.PROTOCOL_ADDRESSES)

    def test_blocked_module_names_mirror_app_go(self) -> None:
        source = (REPO_ROOT / "app/app.go").read_text()
        macc = re.search(r"maccPerms = map\[string\]\[\]string\{(.*?)\n\t\}", source, re.DOTALL).group(1)
        identifiers = re.findall(r"^\s*([\w.]+):", macc, re.MULTILINE)
        blacklist = re.search(r"func \(app \*StrideApp\) BlacklistedModuleAccountAddrs\(\).*?\n\}\n", source, re.DOTALL).group(0)
        excluded = set(re.findall(r"acc == ([\w.]+)", blacklist))
        blocked = [identifier for identifier in identifiers if identifier not in excluded]

        # The Go identifiers resolve to module names through this table; a maccPerms key added
        # without a row here fails the test, which is the prompt to update the builder
        names = {
            "authtypes.FeeCollectorName": "fee_collector",
            "distrtypes.ModuleName": "distribution",
            "ccvconsumertypes.ConsumerRedistributeName": "cons_redistribute",
            "minttypes.ModuleName": "mint",
            "stakingtypes.BondedPoolName": "bonded_tokens_pool",
            "stakingtypes.NotBondedPoolName": "not_bonded_tokens_pool",
            "govtypes.ModuleName": "gov",
            "ibctransfertypes.ModuleName": "transfer",
            "claimtypes.ModuleName": "claim",
            "interchainquerytypes.ModuleName": "interchainquery",
            "icatypes.ModuleName": "interchainaccounts",
            "wasmtypes.ModuleName": "wasm",
            "icqoracletypes.ModuleName": "icqoracle",
            "auctiontypes.ModuleName": "auction",
            "strdburnertypes.ModuleName": "strdburner",
            "poatypes.ModuleName": "poa",
        }
        self.assertEqual(len(excluded), 7)
        self.assertEqual(sorted(names[identifier] for identifier in blocked), sorted(build_sweep_batches.BLOCKED_MODULE_NAMES))

    def test_bech32_round_trip_matches_the_on_chain_derivation(self) -> None:
        # The stride and osmo forms of the F5 key are the same 20 bytes under two prefixes
        hrp, address = bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        self.assertEqual(hrp, "stride")
        self.assertEqual(len(address), 20)
        self.assertEqual(bech32_ref.encode("osmo", address), "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af")
        self.assertEqual(bech32_ref.encode("stride", address), "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        with self.assertRaises(ValueError):
            bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlx")  # bad checksum

    def test_real_mainnet_address_pair_and_channel_5_escrow(self) -> None:
        # Values verified against Stride mainnet REST on 2026-09-30; the Go keeper test pins the same pair
        hrp, address_bytes = bech32_ref.decode("stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c")
        self.assertEqual(hrp, "stride")
        self.assertEqual(
            bech32_ref.encode("osmo", address_bytes), "osmo1am99pcvynqqhyrwqfvfmnvxjk96rn46lj4pkkx"
        )
        self.assertEqual(
            build_sweep_batches.escrow_address("transfer", "channel-5"),
            "stride16h2ynrzwhxgjnd0hswkvdvq9nav9kklq08fhf4",
        )


if __name__ == "__main__":
    unittest.main()
