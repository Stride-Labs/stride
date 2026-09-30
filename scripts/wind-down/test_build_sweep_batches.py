"""Unit tests for build_sweep_batches over a synthetic export.

    python3 -m unittest scripts/wind-down/test_build_sweep_batches.py
"""

import argparse
import json
import pathlib
import tempfile
import unittest

import bech32_ref
import build_sweep_batches

STRIDE_BASE = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
STRIDE_VESTING = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
STRIDE_MODULE = "stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan"  # the distribution module account: sha256("distribution")[:20]
STRIDE_CONTINUOUS = "stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c"  # ContinuousVestingAccount, 40% through its schedule
STRIDE_ICA = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04"
STRIDE_NO_ACCOUNT = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g"
STRIDE_DUST = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd"
ATOM_VOUCHER = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"


GENESIS_TIME = "2026-09-29T00:00:00Z"
AS_OF = 1790640000  # GENESIS_TIME as unix seconds


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
    ]
    accounts = [
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_BASE},
        {"@type": "/stride.vesting.StridePeriodicVestingAccount", "base_vesting_account": {"base_account": {"address": STRIDE_VESTING}, "original_vesting": [], "delegated_vesting": [], "end_time": "0"}, "vesting_periods": []},
        # 1,000,000 ustrd vesting linearly from AS_OF-400 to AS_OF+600: 400,000 vested, 600,000 locked at the export
        {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount", "base_vesting_account": {"base_account": {"address": STRIDE_CONTINUOUS}, "original_vesting": [{"denom": "ustrd", "amount": "1000000"}], "delegated_vesting": [], "end_time": str(AS_OF + 600)}, "start_time": str(AS_OF - 400)},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": contract},
        {"@type": "/cosmos.auth.v1beta1.ModuleAccount", "base_account": {"address": STRIDE_MODULE}, "name": "distribution"},
        {"@type": "/ibc.applications.interchain_accounts.v1.InterchainAccount", "base_account": {"address": STRIDE_ICA}, "account_owner": "x"},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": STRIDE_DUST},
        {"@type": "/cosmos.auth.v1beta1.BaseAccount", "address": escrow},
    ]
    channels = [{"port_id": "transfer", "channel_id": "channel-5", "state": "STATE_OPEN"}]
    contracts = [{"contract_address": contract, "contract_info": {"code_id": "1"}}]
    return {
        "genesis_time": GENESIS_TIME,
        "app_state": {
            "bank": {"balances": balances},
            "auth": {"accounts": accounts},
            "ibc": {"channel_genesis": {"channels": channels}},
            "wasm": {"contracts": contracts},
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
        self.export = build_sweep_batches.load_export(self.export_path)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_classify_applies_skip_rules_and_floor(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)

        swept = [holder.address for holder in plan.holders]
        self.assertEqual(swept, [STRIDE_BASE, STRIDE_VESTING, STRIDE_CONTINUOUS],
                         "ordered by USD: base ($100.25), stride vesting ($20), continuous ($10.02 on its spendable part)")

        skipped = {entry.address: entry.reason for entry in plan.skipped}
        self.assertEqual(skipped[STRIDE_MODULE], "account type /cosmos.auth.v1beta1.ModuleAccount is not sweepable")
        self.assertEqual(skipped[build_sweep_batches.escrow_address("transfer", "channel-999")], "wasm contract address")
        self.assertEqual(skipped[STRIDE_ICA], "account type /ibc.applications.interchain_accounts.v1.InterchainAccount is not sweepable")
        self.assertEqual(skipped[STRIDE_NO_ACCOUNT], "account not found")
        self.assertEqual(skipped[build_sweep_batches.escrow_address("transfer", "channel-5")], "transfer escrow address")
        self.assertEqual(skipped[STRIDE_DUST], "below floor ($0.01 < $1.00)")

    def test_write_batches_splits_at_batch_size(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)
        out_dir = pathlib.Path(self.tmp.name) / "out"

        files = build_sweep_batches.write_batches(plan=plan, out_dir=out_dir, batch_size=1)

        self.assertEqual([path.name for path in files], ["batch-001.txt", "batch-002.txt", "batch-003.txt"])
        self.assertEqual(files[0].read_text().strip(), STRIDE_BASE)
        summary = json.loads((out_dir / "summary.json").read_text())
        self.assertEqual(summary["denoms"], ["stuatom", "ustrd"])
        self.assertEqual(summary["batches"][0]["num_addresses"], 1)
        self.assertEqual(len(summary["skipped"]), 6)

    def test_vesting_locked_balance_is_not_swept(self) -> None:
        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom", "ustrd"], prices=PRICES, floor_usd=1.0)
        continuous = next(holder for holder in plan.holders if holder.address == STRIDE_CONTINUOUS)

        # 1,000,000 ustrd held, 600,000 still locked at genesis_time: the chain's SpendableCoin is 400,000
        self.assertEqual(continuous.balances, {"stuatom": 1000000, "ustrd": 400000})
        self.assertEqual(f"{continuous.usd:.2f}", "10.02")

        schedule = self.export.vesting[STRIDE_CONTINUOUS]
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF), {"ustrd": 600000})
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF + 600), {})
        self.assertEqual(build_sweep_batches.locked_at(schedule=schedule, as_of=AS_OF - 400), {"ustrd": 1000000})

    def test_extra_denom_parsing_and_whitelist(self) -> None:
        extra = build_sweep_batches.parse_extra_denom("ustrd=0.05:6")
        self.assertEqual((extra.denom, extra.usd_per_token, extra.decimals), ("ustrd", 0.05, 6))

        with self.assertRaises(ValueError):
            build_sweep_batches.parse_extra_denom("ustrd=0.05")

    def test_ibc_denom_requires_whitelisted_outer_hop_however_it_is_passed(self) -> None:
        traces = {ATOM_VOUCHER: ["transfer/channel-0"], "ibc/AAAA": ["transfer/channel-52"]}
        build_sweep_batches.check_denom_destination(denom="stuatom", traces=traces)
        build_sweep_batches.check_denom_destination(denom=ATOM_VOUCHER, traces=traces)

        # The same check guards a voucher given through --denoms, not only --extra-denom
        with self.assertRaises(ValueError) as raised:
            build_sweep_batches.check_denom_destination(denom="ibc/AAAA", traces=traces)
        self.assertIn("ibc/AAAA", str(raised.exception))
        self.assertIn("channel-52", str(raised.exception))

        with self.assertRaises(ValueError):
            build_sweep_batches.check_denom_destination(denom="ibc/BBBB", traces=traces)  # no trace at all

    def test_batch_size_is_bounded_by_the_chain_maximum(self) -> None:
        self.assertEqual(build_sweep_batches.batch_size_arg("100"), 100)
        self.assertEqual(build_sweep_batches.batch_size_arg("1"), 1)
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.batch_size_arg("101")
        with self.assertRaises(argparse.ArgumentTypeError):
            build_sweep_batches.batch_size_arg("0")

        plan = build_sweep_batches.classify_holders(export=self.export, denoms=["stuatom"], prices=PRICES, floor_usd=1.0)
        with self.assertRaises(ValueError):
            build_sweep_batches.write_batches(plan=plan, out_dir=pathlib.Path(self.tmp.name) / "out", batch_size=101)

    def test_bech32_round_trip_matches_the_on_chain_derivation(self) -> None:
        # The stride and osmo forms of the F5 key are the same 20 bytes under two prefixes
        hrp, address = bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        self.assertEqual(hrp, "stride")
        self.assertEqual(len(address), 20)
        self.assertEqual(bech32_ref.encode("osmo", address), "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af")
        self.assertEqual(bech32_ref.encode("stride", address), "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh")
        with self.assertRaises(ValueError):
            bech32_ref.decode("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlx")  # bad checksum


if __name__ == "__main__":
    unittest.main()
