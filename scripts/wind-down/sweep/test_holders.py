"""Holder classification over a synthetic chain: every skip reason, the floor, keyless detection, the exclusions file,
and a denom the chain would refuse. chainio is patched; nothing touches the network."""

import json
import pathlib
import tempfile
import unittest
from decimal import Decimal
from unittest import mock

from sweep import addresses, chainio, config, holders

BASE = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
VESTING = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"  # also the F5 multisig in exclusions.json; the test file below overrides
KEYLESS = "stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c"
ICA = "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04"  # also a protocol address: protocol wins (checked first)
MISSING = "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g"
DUST = "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd"
MODULE = addresses.module_address(name="distribution")
ESCROW = addresses.escrow_address(channel_id="channel-5")
CONTRACT = addresses.escrow_address(channel_id="channel-999")  # any 20-byte address not otherwise used
LOCKED = addresses.module_address(name="locked-vesting-twenty-bytes")
DUSTY = addresses.module_address(name="dusty-vesting-twenty-bytes")
EXCLUDED = addresses.module_address(name="not-a-real-module-just-twenty-bytes")
ATOM = addresses.ibc_denom(path="transfer/channel-0/uatom")

HOST_ZONES = [
    {"chain_id": "cosmoshub-4", "host_denom": "uatom", "redemption_rate": "1.5", "deprecated": False},
    {"chain_id": "celestia", "host_denom": "utia", "redemption_rate": "1.2", "deprecated": False},
    {"chain_id": "evmos_9001-2", "host_denom": "aevmos", "redemption_rate": "1.1", "deprecated": True},
]
OWNERS = {
    "stuatom": [(BASE, "10000000"), (VESTING, "2000000"), (KEYLESS, "1000000"), (ICA, "99000000"), (MISSING, "99000000"),
                (DUST, "1000"), (LOCKED, "3000000"), (DUSTY, "3000000"), (MODULE, "99000000"), (ESCROW, "99000000"), (CONTRACT, "99000000"), (EXCLUDED, "99000000")],
    "ustrd": [(BASE, "5000000"), (VESTING, "100000000")],
    ATOM: [(BASE, "250000")],
}
ACCOUNTS = {
    BASE: {"@type": config.BASE_ACCOUNT, "address": BASE, "pub_key": {"key": "x"}, "sequence": "4"},
    VESTING: {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
              "base_vesting_account": {"base_account": {"address": VESTING, "pub_key": {"key": "x"}, "sequence": "1"}}},
    LOCKED: {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
              "base_vesting_account": {"base_account": {"address": LOCKED, "pub_key": {"key": "x"}, "sequence": "1"}}},
    DUSTY: {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
              "base_vesting_account": {"base_account": {"address": DUSTY, "pub_key": {"key": "x"}, "sequence": "1"}}},
    KEYLESS: {"@type": config.BASE_ACCOUNT, "address": KEYLESS, "pub_key": None, "sequence": "0"},
    ICA: {"@type": config.INTERCHAIN_ACCOUNT, "base_account": {"address": ICA, "pub_key": None, "sequence": "0"}},
    DUST: {"@type": config.BASE_ACCOUNT, "address": DUST, "pub_key": {"key": "x"}, "sequence": "9"},
    MODULE: {"@type": config.MODULE_ACCOUNT, "base_account": {"address": MODULE, "pub_key": None, "sequence": "0"}, "name": "distribution"},
    ESCROW: {"@type": config.BASE_ACCOUNT, "address": ESCROW, "pub_key": None, "sequence": "0"},
    CONTRACT: {"@type": config.BASE_ACCOUNT, "address": CONTRACT, "pub_key": None, "sequence": "0"},
    EXCLUDED: {"@type": config.BASE_ACCOUNT, "address": EXCLUDED, "pub_key": {"key": "x"}, "sequence": "2"},
}
SPENDABLE = {LOCKED: [], DUSTY: [{"denom": "stuatom", "amount": "2000"}], VESTING: [{"denom": "stuatom", "amount": "2000000"}, {"denom": "ustrd", "amount": "40000000"}]}


def fake_rest_get(path: str, params: dict[str, str] | None = None) -> dict:
    params = params or {}
    if path == "/cosmos/bank/v1beta1/denom_owners_by_query":
        owners = OWNERS.get(params["denom"], [])
        return {"denom_owners": [{"address": a, "balance": {"denom": params["denom"], "amount": v}} for a, v in owners],
                "pagination": {"next_key": None}}
    if path.startswith("/cosmos/auth/v1beta1/accounts/"):
        address = path.rsplit("/", 1)[1]
        if address not in ACCOUNTS:
            raise chainio.NotFound(path)
        return {"account": ACCOUNTS[address]}
    if path.startswith("/cosmos/bank/v1beta1/spendable_balances/"):
        return {"balances": SPENDABLE[path.rsplit("/", 1)[1]], "pagination": {"next_key": None}}
    if path == "/Stride-Labs/stride/stakeibc/host_zone":
        return {"host_zone": HOST_ZONES, "pagination": {"next_key": None}}
    if path.startswith("/ibc/apps/transfer/v1/denoms/"):
        if path.endswith(ATOM.removeprefix("ibc/")):
            return {"denom": {"base": "uatom", "trace": [{"port_id": "transfer", "channel_id": "channel-0"}]}}
        raise chainio.NotFound(path)
    if path == "/cosmos/auth/v1beta1/module_accounts":
        return {"accounts": [ACCOUNTS[MODULE]]}
    if path == "/ibc/core/channel/v1/channels":
        return {"channels": [{"port_id": "transfer", "channel_id": "channel-5", "state": "STATE_OPEN"},
                             {"port_id": "icahost", "channel_id": "channel-9", "state": "STATE_OPEN"}], "pagination": {"next_key": None}}
    if path == "/cosmwasm/wasm/v1/code":
        return {"code_infos": [{"code_id": "1"}], "pagination": {"next_key": None}}
    if path == "/cosmwasm/wasm/v1/code/1/contracts":
        return {"contracts": [CONTRACT, "stride1" + "q" * 58], "pagination": {"next_key": None}}
    raise AssertionError(f"unexpected path {path} {params}")


def two_denoms() -> list[holders.SweepDenom]:
    return [
        holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5"),
        holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5"),
        holders.SweepDenom(denom=ATOM, symbol="ATOM", decimals=6, price_usd=Decimal("4"), destination="cosmoshub-4", channel="channel-0"),
    ]


class HolderTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.dir = pathlib.Path(self.tmp.name)
        self.exclusions_path = self.dir / "exclusions.json"
        self.exclusions_path.write_text(json.dumps({"sections": [
            {"name": "team", "reason": "moved by hand", "addresses": [{"address": EXCLUDED, "label": "F5"}]},
        ]}))
        # The synthetic chain knows three denoms; the real config lists twenty
        for name, value in (("NATIVE_SWEEP_DENOMS", (config.NativeDenom("stuatom", "stATOM", 6, "ATOM", "cosmoshub-4"),
                                                      config.NativeDenom("ustrd", "STRD", 6, "STRD", None))),
                            ("VOUCHER_SWEEP_DENOMS", (config.VoucherDenom("uatom", "channel-0", "ATOM", 6, "ATOM", "cosmoshub-4"),))):
            config_patcher = mock.patch.object(config, name, value)
            config_patcher.start()
            self.addCleanup(config_patcher.stop)
        patcher = mock.patch.object(chainio, "rest_get", side_effect=fake_rest_get)
        patcher.start()
        self.addCleanup(patcher.stop)
        height_patcher = mock.patch.object(chainio, "latest_height", return_value=100)
        height_patcher.start()
        self.addCleanup(height_patcher.stop)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_resolve_denoms_prices_sttokens_at_the_rate_and_checks_vouchers(self) -> None:
        with mock.patch.object(config, "NATIVE_SWEEP_DENOMS", (config.NativeDenom("stuatom", "stATOM", 6, "ATOM", "cosmoshub-4"),
                                                                 config.NativeDenom("ustrd", "STRD", 6, "STRD", None))), \
             mock.patch.object(config, "VOUCHER_SWEEP_DENOMS", (config.VoucherDenom("uatom", "channel-0", "ATOM", 6, "ATOM", "cosmoshub-4"),)):
            denoms = holders.resolve_denoms()
        self.assertEqual([(d.denom, d.price_usd, d.destination, d.channel) for d in denoms], [
            ("stuatom", Decimal("6"), "osmosis-1", "channel-5"), ("ustrd", Decimal("0.05"), "osmosis-1", "channel-5"),
            (ATOM, Decimal("4"), "cosmoshub-4", "channel-0"),
        ])

    def test_resolve_denoms_refuses_a_deprecated_zone_and_an_unwhitelisted_voucher(self) -> None:
        with mock.patch.object(config, "NATIVE_SWEEP_DENOMS", (config.NativeDenom("staevmos", "stEVMOS", 18, "ATOM", "evmos_9001-2"),)), \
             mock.patch.object(config, "VOUCHER_SWEEP_DENOMS", ()):
            with self.assertRaisesRegex(holders.DenomError, "staevmos"):
                holders.resolve_denoms()
        with mock.patch.object(config, "NATIVE_SWEEP_DENOMS", ()), \
             mock.patch.object(config, "VOUCHER_SWEEP_DENOMS", (config.VoucherDenom("uatom", "channel-999", "ATOM", 6, "ATOM", "x"),)):
            with self.assertRaisesRegex(holders.DenomError, "channel-999"):
                holders.resolve_denoms()

    def test_classify_applies_every_rule(self) -> None:
        exclusions = holders.load_exclusions(path=self.exclusions_path)
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        holder_set = holders.read_holder_set(floor_usd=Decimal("1"), exclusions=exclusions, accounts=accounts, test_address=None)

        self.assertEqual([h.address for h in holder_set.holders], [BASE, VESTING, KEYLESS])
        by_address = {h.address: h for h in holder_set.holders}
        self.assertEqual(by_address[BASE].usd, Decimal("61.25"))  # 10 stATOM x 6 + 5 STRD x 0.05 + 0.25 ATOM x 4
        self.assertEqual(by_address[VESTING].balances, {"stuatom": 2_000_000, "ustrd": 40_000_000})  # spendable, not total
        self.assertTrue(by_address[KEYLESS].keyless)
        self.assertFalse(by_address[BASE].keyless)
        self.assertEqual(by_address[BASE].transfers, 3)

        skipped = {entry.address: entry.reason for entry in holder_set.skipped}
        self.assertEqual(skipped[ICA], "protocol address")
        self.assertEqual(skipped[MISSING], "protocol address")
        self.assertEqual(skipped[MODULE], "blocked module address")
        self.assertEqual(skipped[ESCROW], "transfer escrow address")
        self.assertEqual(skipped[CONTRACT], "wasm contract address")
        self.assertEqual([(e.address, e.reason) for e in holder_set.excluded], [(EXCLUDED, "excluded: team: F5")])
        self.assertEqual([h.address for h in holder_set.below_floor], [DUSTY, DUST, LOCKED])
        below = {h.address: h for h in holder_set.below_floor}
        self.assertEqual((below[DUSTY].balances, below[DUSTY].usd), ({"stuatom": 2000}, Decimal("0.012")))  # floor applies to spendable
        self.assertEqual((below[LOCKED].balances, below[LOCKED].usd), ({}, Decimal("0")))  # fully locked is recorded, not dropped
        self.assertEqual(holder_set.height, 100)

    def test_skip_reason_order_and_unknown_accounts(self) -> None:
        inputs = holders.SkipInputs(module_addresses=set(), escrows=set(), contracts=set())
        self.assertEqual(holders.skip_reason(address="stride1short", account=None, inputs=inputs, exclusions={}), "address is not 20 bytes")
        self.assertEqual(holders.skip_reason(address=BASE, account=None, inputs=inputs, exclusions={}), "account not found")
        ica = holders.AccountInfo(type=config.INTERCHAIN_ACCOUNT, has_pubkey=False, sequence=0)
        self.assertEqual(holders.skip_reason(address=BASE, account=ica, inputs=inputs, exclusions={}), "interchain account")
        module = holders.AccountInfo(type=config.MODULE_ACCOUNT, has_pubkey=False, sequence=0)
        self.assertEqual(holders.skip_reason(address=BASE, account=module, inputs=inputs, exclusions={}),
                         f"account type {config.MODULE_ACCOUNT} is not sweepable")
        base = holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=1)
        self.assertIsNone(holders.skip_reason(address=BASE, account=base, inputs=inputs, exclusions={}))
        self.assertEqual(holders.skip_reason(address=config.SWEEP_OPERATOR, account=base, inputs=inputs, exclusions={}), "protocol address")

    def test_test_address_ignores_the_floor_and_is_the_only_holder(self) -> None:
        exclusions = holders.load_exclusions(path=self.exclusions_path)
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        holder_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, test_address=BASE)
        self.assertEqual([h.address for h in holder_set.holders], [BASE])

    def test_account_cache_persists_and_rereads_keyless_entries(self) -> None:
        path = self.dir / "accounts.json"
        accounts = holders.AccountCache.load(path=path)
        with mock.patch.object(chainio, "rest_get", side_effect=fake_rest_get) as rest_get:
            accounts.lookup(address=BASE)
            accounts.lookup(address=BASE)
            accounts.lookup(address=KEYLESS)
            accounts.lookup(address=KEYLESS)
            self.assertIsNone(accounts.lookup(address=MISSING))
        self.assertEqual(rest_get.call_count, 4)  # BASE once, KEYLESS twice, MISSING once (not cached)
        accounts.save(path=path)
        reloaded = holders.AccountCache.load(path=path)
        self.assertEqual(reloaded.lookup(address=BASE), holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=4))

    def test_load_exclusions_rejects_bad_addresses_and_duplicates(self) -> None:
        self.exclusions_path.write_text(json.dumps({"sections": [{"name": "x", "reason": "r", "addresses": [{"address": "osmo1bad", "label": "l"}]}]}))
        with self.assertRaises(holders.ExclusionsError):
            holders.load_exclusions(path=self.exclusions_path)
        self.exclusions_path.write_text(json.dumps({"sections": [
            {"name": "x", "reason": "r", "addresses": [{"address": BASE, "label": "l"}, {"address": BASE, "label": "again"}]}]}))
        with self.assertRaises(holders.ExclusionsError):
            holders.load_exclusions(path=self.exclusions_path)

    def test_shipped_exclusions_file_loads(self) -> None:
        exclusions = holders.load_exclusions()
        self.assertIn("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh", exclusions)
        self.assertEqual(exclusions["stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"].section, "team")


    def test_test_address_keeps_a_fully_locked_vesting_account_in_below_floor(self) -> None:
        exclusions = holders.load_exclusions(path=self.exclusions_path)
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        holder_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, test_address=LOCKED)
        self.assertEqual(([h.address for h in holder_set.holders], [h.address for h in holder_set.below_floor]), ([], [LOCKED]))

        dusty_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, test_address=DUSTY)
        self.assertEqual(([h.address for h in dusty_set.holders], dusty_set.below_floor), ([DUSTY], []))  # floor ignored


if __name__ == "__main__":
    unittest.main()
