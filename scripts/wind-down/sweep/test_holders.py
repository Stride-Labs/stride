"""Holder classification over a synthetic chain: every skip reason, the floor, keyless detection, the exclusions file,
and a denom the chain would refuse. chainio is patched; nothing touches the network."""

import collections
import datetime
import json
import os
import threading
import pathlib
import tempfile
import time
import unittest
from decimal import Decimal
from unittest import mock

from sweep import addresses, chainio, config, holders

BASE = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
VESTING = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"  # also the F5 multisig in the shipped exclusions.json; setUp's exclusions file lists EXCLUDED instead
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


NOW = datetime.datetime(2026, 10, 14, 12, 0, tzinfo=datetime.UTC)
COMPLETES = "2026-10-26T08:55:25.123456789Z"
# address -> the unbonding_responses the fake chain answers with; empty means no unbonding
UNBONDING: dict[str, list[dict]] = {}
STAKING_READS: list[str] = []


def NO_STAKING() -> holders.StakingCache:
    return holders.StakingCache(entries={})


def unbonding(address: str, *entries: tuple[str, str]) -> dict:
    return {"delegator_address": address,
            "entries": [{"balance": balance, "completion_time": completion} for balance, completion in entries]}


def fake_rest_get(path: str, params: dict[str, str] | None = None, height: int | None = None,
                  base: str = config.REST) -> dict:
    params = params or {}
    if path.startswith("/cosmos/staking/v1beta1/delegators/"):
        address = path.split("/")[5]
        STAKING_READS.append(address)
        return {"unbonding_responses": UNBONDING.get(address, []), "pagination": {"next_key": None}}
    if path == "/cosmos/staking/v1beta1/pool":
        return {"pool": {"not_bonded_tokens": "700", "bonded_tokens": "50"}}
    if path == "/cosmos/bank/v1beta1/denom_owners_by_query":
        owners = OWNERS.get(params["denom"], [])
        return {"denom_owners": [{"address": a, "balance": {"denom": params["denom"], "amount": v}} for a, v in owners],
                "pagination": {"next_key": None, "total": str(len(owners))}}
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


class HolderTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.dir = pathlib.Path(self.tmp.name)
        # Caches built without a path save to these, never to the real state directory
        for name in ("ACCOUNTS_CACHE_PATH", "STAKING_CACHE_PATH"):
            path_patcher = mock.patch.object(config, name, self.dir / f"default-{name.lower()}.json")
            path_patcher.start()
            self.addCleanup(path_patcher.stop)
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
        UNBONDING.clear()
        STAKING_READS.clear()
        height_patcher = mock.patch.object(chainio, "latest_height", return_value=100)
        height_patcher.start()
        self.addCleanup(height_patcher.stop)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_resolve_denoms_prices_sttokens_at_the_rate_and_checks_vouchers(self) -> None:
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
        holder_set = holders.read_holder_set(floor_usd=Decimal("1"), exclusions=exclusions, accounts=accounts, staking=NO_STAKING(), test_address=None)

        self.assertEqual([h.address for h in holder_set.holders], [BASE, VESTING, KEYLESS])
        by_address = {h.address: h for h in holder_set.holders}
        self.assertEqual(by_address[BASE].usd, Decimal("61.25"))  # 10 stATOM x 6 + 5 STRD x 0.05 + 0.25 ATOM x 4
        self.assertEqual(by_address[VESTING].balances, {"stuatom": 2_000_000, "ustrd": 40_000_000})  # spendable, not total
        self.assertEqual(by_address[VESTING].locked, {"ustrd": 60_000_000})  # total minus spendable; stuatom is fully spendable
        self.assertEqual((by_address[BASE].locked, by_address[KEYLESS].locked), ({}, {}))
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
        self.assertEqual((below[LOCKED].locked, below[DUSTY].locked), ({"stuatom": 3_000_000}, {"stuatom": 2_998_000}))
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
        holder_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, staking=NO_STAKING(), test_address=BASE)
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
            self.assertIsNone(accounts.lookup(address=MISSING))
            self.assertEqual(rest_get.call_count, 3)  # each read once in this process, keyless and missing included
            accounts.save(path=path)

            # A later process re-reads the keyless one (it may have signed since) but not the final one
            reloaded = holders.AccountCache.load(path=path)
            self.assertEqual(reloaded.lookup(address=BASE),
                             holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=4))
            self.assertEqual(rest_get.call_count, 3)
            reloaded.lookup(address=KEYLESS)
            self.assertEqual(rest_get.call_count, 4)

    def test_load_exclusions_rejects_bad_addresses_and_duplicates(self) -> None:
        self.exclusions_path.write_text(json.dumps({"sections": [{"name": "x", "reason": "r", "addresses": [{"address": "osmo1bad", "label": "l"}]}]}))
        with self.assertRaises(holders.ExclusionsError):
            holders.load_exclusions(path=self.exclusions_path)
        self.exclusions_path.write_text(json.dumps({"sections": [
            {"name": "x", "reason": "r", "addresses": [{"address": BASE, "label": "l"}, {"address": BASE, "label": "again"}]}]}))
        with self.assertRaises(holders.ExclusionsError):
            holders.load_exclusions(path=self.exclusions_path)

    def test_skip_reason_contract_and_exclusion_rules_come_last(self) -> None:
        base = holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=1)
        exclusion = holders.Exclusion(address=BASE, section="team", label="F5", reason="moved by hand")
        inputs = holders.SkipInputs(module_addresses=set(), escrows=set(), contracts={BASE})
        self.assertEqual(holders.skip_reason(address=BASE, account=base, inputs=inputs, exclusions={BASE: exclusion}),
                         "wasm contract address")  # a contract wins over an exclusion
        no_contracts = holders.SkipInputs(module_addresses=set(), escrows=set(), contracts=set())
        self.assertEqual(holders.skip_reason(address=BASE, account=base, inputs=no_contracts, exclusions={BASE: exclusion}),
                         "excluded: team: F5")

        earlier_rule = holders.SkipInputs(module_addresses={BASE}, escrows=set(), contracts={BASE})
        self.assertEqual(holders.skip_reason(address=BASE, account=base, inputs=earlier_rule, exclusions={BASE: exclusion}),
                         "blocked module address")
        self.assertEqual(holders.skip_reason(address=BASE, account=None, inputs=inputs, exclusions={BASE: exclusion}),
                         "account not found")

    def test_classify_skips_a_non_20_byte_holder_and_an_unsweepable_account_type(self) -> None:
        short = "stride1short"
        denoms = [holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal(1),
                                     destination="osmosis-1", channel="channel-5")]
        balances = {short: {"ustrd": 5_000_000}, BASE: {"ustrd": 5_000_000}}
        module_type = holders.AccountInfo(type=config.MODULE_ACCOUNT, has_pubkey=True, sequence=1)
        accounts = holders.AccountCache(entries={BASE: module_type})
        inputs = holders.SkipInputs(module_addresses=set(), escrows=set(), contracts=set())

        holder_set = holders.classify(denoms=denoms, balances=balances, floor_usd=Decimal(1), exclusions={}, inputs=inputs,
                                      accounts=accounts, staking=NO_STAKING(), now=NOW, test_address=None, height=1)

        self.assertEqual(holder_set.holders, [])
        self.assertEqual({entry.address: entry.reason for entry in holder_set.skipped},
                         {short: "address is not 20 bytes", BASE: f"account type {config.MODULE_ACCOUNT} is not sweepable"})

    def test_cached_keyless_entry_is_reread_until_it_has_a_pubkey_then_final(self) -> None:
        keyless = {"@type": config.BASE_ACCOUNT, "address": KEYLESS, "pub_key": None, "sequence": "0"}
        signed = {"@type": config.BASE_ACCOUNT, "address": KEYLESS, "pub_key": {"key": "x"}, "sequence": "1"}
        accounts = holders.AccountCache(entries={KEYLESS: holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=False, sequence=0)})
        with mock.patch.object(chainio, "rest_get", side_effect=[{"account": keyless}, {"account": signed}]) as rest_get:
            self.assertFalse(accounts.lookup(address=KEYLESS).has_pubkey)
            self.assertFalse(accounts.lookup(address=KEYLESS).has_pubkey)  # read this process: not read again
            self.assertEqual(rest_get.call_count, 1)

            # A later process (same entries, nothing fresh) reads it again and finds the pubkey
            later = holders.AccountCache(entries=dict(accounts._entries))
            self.assertTrue(later.lookup(address=KEYLESS).has_pubkey)
            self.assertTrue(later.lookup(address=KEYLESS).has_pubkey)  # final now: no third read
        self.assertEqual(rest_get.call_count, 2)

    def test_test_address_absent_from_the_balances_yields_an_empty_set(self) -> None:
        exclusions = holders.load_exclusions(path=self.exclusions_path)
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        holder_set = holders.read_holder_set(floor_usd=Decimal("1"), exclusions=exclusions, accounts=accounts, staking=NO_STAKING(),
                                             test_address="stride1" + "p" * 38)
        self.assertEqual((holder_set.holders, holder_set.excluded, holder_set.skipped, holder_set.below_floor), ([], [], [], []))

    def read_set(self, floor: str = "1", test_address: str | None = None,
                 staking: holders.StakingCache | None = None) -> holders.HolderSet:
        return holders.read_holder_set(
            floor_usd=Decimal(floor), exclusions=holders.load_exclusions(path=self.exclusions_path),
            accounts=holders.AccountCache.load(path=self.dir / "accounts.json"),
            staking=staking or NO_STAKING(), test_address=test_address)

    def test_a_holder_with_unbonding_stake_is_deferred_with_its_balances_usd_and_locked_part(self) -> None:
        UNBONDING[VESTING] = [unbonding(VESTING, ("5", "2026-10-26T08:55:25Z"))]
        UNBONDING[BASE] = [unbonding(BASE, ("7", COMPLETES))]

        holder_set = self.read_set()

        self.assertEqual([h.address for h in holder_set.holders], [KEYLESS])
        self.assertEqual([entry.holder.address for entry in holder_set.deferred], [BASE, VESTING])  # USD descending
        by_address = {entry.holder.address: entry for entry in holder_set.deferred}
        vesting = by_address[VESTING]
        self.assertEqual(vesting.holder.balances, {"stuatom": 2_000_000, "ustrd": 40_000_000})
        self.assertEqual(vesting.holder.usd, Decimal("14"))
        self.assertEqual(vesting.holder.locked, {"ustrd": 60_000_000})
        self.assertEqual(vesting.staked, holders.Staked(ustrd=5, completes_at="2026-10-26T08:55:25Z"))
        self.assertEqual(by_address[BASE].holder.usd, Decimal("61.25"))
        self.assertEqual(by_address[BASE].staked, holders.Staked(ustrd=7, completes_at=COMPLETES))

    def test_the_test_plan_never_defers_and_reads_no_staking(self) -> None:
        UNBONDING[BASE] = [unbonding(BASE, ("7", COMPLETES))]
        holder_set = self.read_set(test_address=BASE)
        self.assertEqual(([h.address for h in holder_set.holders], holder_set.deferred), ([BASE], []))
        self.assertEqual(STAKING_READS, [])

    def test_a_below_floor_staker_stays_below_floor_and_is_not_deferred(self) -> None:
        UNBONDING[DUST] = [unbonding(DUST, ("7", COMPLETES))]
        UNBONDING[DUSTY] = [unbonding(DUSTY, ("7", COMPLETES))]  # vesting, spendable value under the floor

        holder_set = self.read_set()

        self.assertEqual(holder_set.deferred, [])
        self.assertIn(DUST, [h.address for h in holder_set.below_floor])
        self.assertIn(DUSTY, [h.address for h in holder_set.below_floor])
        self.assertNotIn(DUST, STAKING_READS)
        # Prefetch reads by balance value, before the spendable part is known: DUSTY costs one extra read, no more
        self.assertEqual(STAKING_READS.count(DUSTY), 1)
        self.assertNotIn(ICA, STAKING_READS)  # skipped holders are not read either

    def test_a_keyless_staker_is_deferred_too(self) -> None:
        UNBONDING[KEYLESS] = [unbonding(KEYLESS, ("1", COMPLETES))]
        holder_set = self.read_set()
        self.assertEqual([(e.holder.address, e.holder.keyless) for e in holder_set.deferred], [(KEYLESS, True)])

    def test_read_staked_sums_entries_across_validators_and_keeps_the_latest_completion(self) -> None:
        UNBONDING[BASE] = [unbonding(BASE, ("10", "2026-10-26T08:55:25.5Z"), ("20", "2026-10-26T08:55:25.123456789Z")),
                           unbonding(BASE, ("30", "2026-10-26T08:55:24Z"))]
        self.assertEqual(holders.read_staked(address=BASE), holders.Staked(ustrd=60, completes_at="2026-10-26T08:55:25.5Z"))
        self.assertEqual(holders.read_staked(address=MISSING), holders.Staked(ustrd=0, completes_at=None))

    def test_read_staking_pool_adds_not_bonded_and_bonded(self) -> None:
        self.assertEqual(holders.read_staking_pool(), 750)

    def test_parse_chain_time_truncates_nine_fractional_digits(self) -> None:
        parsed = holders.parse_chain_time(text="2026-10-26T08:55:25.987654321Z")
        self.assertEqual(parsed, datetime.datetime(2026, 10, 26, 8, 55, 25, 987654, tzinfo=datetime.UTC))
        self.assertEqual(holders.parse_chain_time(text="2026-10-26T08:55:25Z"),
                         datetime.datetime(2026, 10, 26, 8, 55, 25, tzinfo=datetime.UTC))

    def test_staking_cache_serves_zero_forever_and_positive_until_it_completes(self) -> None:
        zero = holders.Staked(ustrd=0, completes_at=None)
        positive = holders.Staked(ustrd=9, completes_at="2026-10-26T08:55:25Z")
        cache = holders.StakingCache(entries={BASE: zero, VESTING: positive})
        UNBONDING[VESTING] = []  # by the re-read the stake has matured

        far_future = datetime.datetime(2030, 1, 1, tzinfo=datetime.UTC)
        self.assertEqual(cache.lookup(address=BASE, now=far_future), zero)
        self.assertEqual(cache.lookup(address=VESTING, now=NOW), positive)
        self.assertEqual(STAKING_READS, [])

        completed = datetime.datetime(2026, 10, 26, 8, 55, 25, tzinfo=datetime.UTC)
        self.assertEqual(cache.lookup(address=VESTING, now=completed), zero)  # equal to completes_at counts as completed
        self.assertEqual(STAKING_READS, [VESTING])
        self.assertEqual(cache.lookup(address=VESTING, now=far_future), zero)  # now a final zero
        self.assertEqual(STAKING_READS, [VESTING])

    def test_staking_cache_reads_a_new_address_once_and_round_trips_its_file(self) -> None:
        UNBONDING[BASE] = [unbonding(BASE, ("12", COMPLETES))]
        path = self.dir / "staking.json"
        cache = holders.StakingCache.load(path=path)
        cache.lookup(address=BASE, now=NOW)
        cache.lookup(address=BASE, now=NOW)
        cache.lookup(address=KEYLESS, now=NOW)
        self.assertEqual(STAKING_READS, [BASE, KEYLESS])

        cache.save(path=path)

        self.assertEqual(json.loads(path.read_text()), {
            BASE: {"ustrd": "12", "completes_at": COMPLETES}, KEYLESS: {"ustrd": "0", "completes_at": None}})
        self.assertEqual(list(json.loads(path.read_text())), sorted([BASE, KEYLESS]))
        self.assertFalse(path.with_suffix(".tmp").exists())
        STAKING_READS.clear()
        reloaded = holders.StakingCache.load(path=path)
        self.assertEqual(reloaded.lookup(address=BASE, now=NOW), holders.Staked(ustrd=12, completes_at=COMPLETES))
        self.assertEqual(STAKING_READS, [])

    def test_load_exclusions_rejects_missing_keys(self) -> None:
        bad_files = [
            {"sections": [{"reason": "r", "addresses": []}]},
            {"sections": [{"name": "x", "addresses": []}]},
            {"sections": [{"name": "x", "reason": "r"}]},
            {"sections": [{"name": "x", "reason": "r", "addresses": [{"label": "l"}]}]},
            {"sections": [{"name": "x", "reason": "r", "addresses": [{"address": BASE}]}]},
        ]
        for contents in bad_files:
            with self.subTest(contents=contents):
                self.exclusions_path.write_text(json.dumps(contents))
                with self.assertRaises(holders.ExclusionsError):
                    holders.load_exclusions(path=self.exclusions_path)

    def test_resolve_denoms_refuses_an_empty_denom_trace(self) -> None:
        def empty_trace(path: str, params: dict[str, str] | None = None, base: str = config.REST) -> dict:
            if path.startswith("/ibc/apps/transfer/v1/denoms/"):
                return {"denom": {"base": "uatom", "trace": []}}
            return fake_rest_get(path=path, params=params)

        with mock.patch.object(chainio, "rest_get", side_effect=empty_trace):
            with self.assertRaisesRegex(holders.DenomError, "uatom"):
                holders.resolve_denoms()

    def test_shipped_exclusions_file_loads(self) -> None:
        exclusions = holders.load_exclusions()
        self.assertIn("stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh", exclusions)
        self.assertEqual(exclusions["stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"].section, "team")


    def test_test_address_keeps_a_fully_locked_vesting_account_in_below_floor(self) -> None:
        exclusions = holders.load_exclusions(path=self.exclusions_path)
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        holder_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, staking=NO_STAKING(), test_address=LOCKED)
        self.assertEqual(([h.address for h in holder_set.holders], [h.address for h in holder_set.below_floor]), ([], [LOCKED]))

        dusty_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, staking=NO_STAKING(), test_address=DUSTY)
        self.assertEqual(([h.address for h in dusty_set.holders], dusty_set.below_floor), ([DUSTY], []))  # floor ignored


    def test_account_cache_save_keeps_another_writers_entries(self) -> None:
        path = self.dir / "accounts.json"
        mine = holders.AccountCache.load(path=path)
        other = holders.AccountCache.load(path=path)
        other.lookup(address=BASE)
        other.save()
        mine.lookup(address=KEYLESS)
        mine.save()  # started empty, saved after the other writer: must not drop BASE

        raw = json.loads(path.read_text())
        self.assertEqual(sorted(raw), sorted([BASE, KEYLESS]))
        self.assertEqual(holders.AccountCache.load(path=path).lookup(address=BASE).sequence, 4)

    def test_account_cache_save_prefers_this_processes_entry_for_the_same_address(self) -> None:
        path = self.dir / "accounts.json"
        path.write_text(json.dumps({KEYLESS: {"type": config.BASE_ACCOUNT, "has_pubkey": False, "sequence": "0"}}))
        mine = holders.AccountCache(entries={KEYLESS: holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=3)})
        mine.save(path=path)
        self.assertEqual(json.loads(path.read_text())[KEYLESS], {"type": config.BASE_ACCOUNT, "has_pubkey": True, "sequence": "3"})

    def test_staking_cache_save_keeps_another_writers_entries(self) -> None:
        path = self.dir / "staking.json"
        mine = holders.StakingCache.load(path=path)
        other = holders.StakingCache.load(path=path)
        other.store(address=BASE, staked=holders.Staked(ustrd=0, completes_at=None))
        other.save()
        mine.store(address=VESTING, staked=holders.Staked(ustrd=7, completes_at=COMPLETES))
        mine.save()

        self.assertEqual(json.loads(path.read_text()), {
            BASE: {"ustrd": "0", "completes_at": None},
            VESTING: {"ustrd": "7", "completes_at": COMPLETES},
        })

    def test_prefetch_skips_final_entries_reads_the_rest_and_stores_them(self) -> None:
        accounts = holders.AccountCache(entries={
            BASE: holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=4),
            KEYLESS: holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=False, sequence=0),  # not final: re-read
        })
        staking = holders.StakingCache(entries={
            BASE: holders.Staked(ustrd=0, completes_at=None),
            KEYLESS: holders.Staked(ustrd=9, completes_at="2026-10-13T00:00:00Z"),  # matured: not final
        })
        UNBONDING[KEYLESS] = [unbonding(KEYLESS, ("5", COMPLETES))]

        result = holders.prefetch(addresses=[BASE, KEYLESS, VESTING, MISSING], accounts=accounts, staking=staking,
                                  now=NOW, workers=2)

        self.assertEqual(result, holders.PrefetchResult(read=3, cached=1))
        self.assertEqual(sorted(STAKING_READS), sorted([KEYLESS, VESTING, MISSING]))  # BASE was final
        self.assertTrue(accounts.is_final(address=VESTING))
        self.assertTrue(accounts.is_final(address=MISSING))  # read this process: 404 stores no entry, lookup says None
        self.assertIsNone(accounts.lookup(address=MISSING))
        self.assertEqual(staking.lookup(address=KEYLESS, now=NOW), holders.Staked(ustrd=5, completes_at=COMPLETES))
        self.assertEqual(staking.lookup(address=MISSING, now=NOW), holders.Staked(ustrd=0, completes_at=None))
        self.assertIn(VESTING, json.loads((self.dir / "default-accounts_cache_path.json").read_text()))
        self.assertIn(MISSING, json.loads((self.dir / "default-staking_cache_path.json").read_text()))

    def test_prefetch_reads_only_what_is_not_final_per_cache(self) -> None:
        accounts = holders.AccountCache(entries={BASE: holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=4)})
        staking = holders.StakingCache(entries={})
        with mock.patch.object(holders, "read_account") as read_account:
            holders.prefetch(addresses=[BASE], accounts=accounts, staking=staking, now=NOW, workers=1)
        read_account.assert_not_called()
        self.assertEqual(STAKING_READS, [BASE])

    def test_prefetch_saves_every_n_addresses_and_at_the_end(self) -> None:
        path = self.dir / "staking.json"
        staking = holders.StakingCache(entries={}, path=path)
        accounts = holders.AccountCache(entries={}, path=self.dir / "accounts.json")
        on_disk: dict[int, int] = {}

        def record(done: int, total: int) -> None:
            on_disk[done] = len(json.loads(path.read_text())) if path.exists() else 0

        with mock.patch.object(config, "PREFETCH_SAVE_EVERY", 2):
            holders.prefetch(addresses=[BASE, VESTING, KEYLESS, MISSING, LOCKED], accounts=accounts, staking=staking,
                             now=NOW, workers=1, on_progress=record)

        # Progress fires before the save of the same address: after 3 done the file holds the first 2
        self.assertEqual(on_disk, {1: 0, 2: 0, 3: 2, 4: 2, 5: 4})
        self.assertEqual(len(json.loads(path.read_text())), 5)

    def test_prefetch_saves_what_it_read_when_a_read_fails(self) -> None:
        path = self.dir / "staking.json"
        staking = holders.StakingCache(entries={}, path=path)
        accounts = holders.AccountCache(entries={}, path=self.dir / "accounts.json")

        def read_staked(address: str) -> holders.Staked:
            if address == VESTING:
                raise chainio.ChainError("boom")
            return holders.Staked(ustrd=0, completes_at=None)

        with mock.patch.object(holders, "read_staked", side_effect=read_staked), self.assertRaises(chainio.ChainError):
            holders.prefetch(addresses=[BASE, VESTING], accounts=accounts, staking=staking, now=NOW, workers=1)
        self.assertEqual(list(json.loads(path.read_text())), [BASE])

    def test_read_holder_set_prefetches_only_candidates_at_or_above_the_floor(self) -> None:
        with mock.patch.object(holders, "prefetch", wraps=holders.prefetch) as prefetch:
            self.read_set(floor="1")

        kwargs = prefetch.call_args.kwargs
        # Not DUST (under the floor), not ICA/MISSING (protocol), MODULE or ESCROW (rules before the account), not EXCLUDED
        self.assertEqual(sorted(kwargs["addresses"]), sorted([BASE, VESTING, KEYLESS, LOCKED, DUSTY, CONTRACT]))
        self.assertEqual(kwargs["workers"], config.READ_WORKERS)

    def test_read_holder_set_prefetches_only_the_test_address_and_takes_workers(self) -> None:
        with mock.patch.object(holders, "prefetch", wraps=holders.prefetch) as prefetch:
            holders.read_holder_set(
                floor_usd=Decimal("1000"), exclusions=holders.load_exclusions(path=self.exclusions_path),
                accounts=holders.AccountCache.load(path=self.dir / "accounts.json"), staking=NO_STAKING(),
                test_address=BASE, workers=3)

        self.assertEqual(prefetch.call_args.kwargs["addresses"], [BASE])
        self.assertEqual(prefetch.call_args.kwargs["workers"], 3)

    def test_classify_after_prefetch_makes_no_further_account_or_staking_reads(self) -> None:
        accounts = holders.AccountCache.load(path=self.dir / "accounts.json")
        staking = NO_STAKING()
        with mock.patch.object(holders, "read_account", wraps=holders.read_account) as read_account:
            holders.read_holder_set(floor_usd=Decimal("1"), exclusions=holders.load_exclusions(path=self.exclusions_path),
                                    accounts=accounts, staking=staking, test_address=None)
        account_reads = [call.kwargs["address"] for call in read_account.call_args_list]
        # Every candidate, KEYLESS included, is read once across prefetch and classify
        counts = collections.Counter(account_reads)
        self.assertIn(KEYLESS, counts)
        self.assertEqual({address: count for address, count in counts.items() if count != 1}, {})
        self.assertEqual(len(STAKING_READS), len(set(STAKING_READS)))


class CacheSaveTests(unittest.TestCase):
    def test_a_save_inside_another_savers_write_to_replace_window_loses_nothing(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "accounts.json"
            info = holders.AccountInfo(type=config.BASE_ACCOUNT, has_pubkey=True, sequence=1)
            cache_a = holders.AccountCache(entries={"stride1a": info})
            cache_b = holders.AccountCache(entries={"stride1b": info})
            first_in_replace = threading.Event()
            real_replace = os.replace

            def replace(source: str, destination: pathlib.Path) -> None:
                # The first saver pauses between writing its temp file and renaming it; the second saves meanwhile
                if not first_in_replace.is_set():
                    first_in_replace.set()
                    second.join(timeout=0.5)
                real_replace(source, destination)

            second = threading.Thread(target=lambda: (first_in_replace.wait(), cache_b.save(path=path)))
            with mock.patch.object(os, "replace", side_effect=replace):
                second.start()
                cache_a.save(path=path)
                second.join(timeout=5)

            self.assertFalse(second.is_alive())
            self.assertEqual(sorted(json.loads(path.read_text())), ["stride1a", "stride1b"])
            self.assertEqual(sorted(entry.name for entry in path.parent.iterdir()), ["accounts.json", "accounts.json.lock"])

    def test_a_failed_write_leaves_no_temp_file_and_the_old_file_intact(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "staking.json"
            path.write_text("{}")
            cache = holders.StakingCache(entries={"stride1a": holders.Staked(ustrd=0, completes_at=None)})

            with mock.patch.object(os, "replace", side_effect=OSError("disk")), self.assertRaises(OSError):
                cache.save(path=path)

            self.assertEqual(path.read_text(), "{}")
            self.assertEqual(sorted(entry.name for entry in path.parent.iterdir()), ["staking.json", "staking.json.lock"])

    def test_staking_save_merges_with_the_files_entries(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "staking.json"
            holders.StakingCache(entries={"stride1a": holders.Staked(ustrd=0, completes_at=None)}).save(path=path)
            holders.StakingCache(entries={"stride1b": holders.Staked(ustrd=0, completes_at=None)}).save(path=path)
            self.assertEqual(sorted(json.loads(path.read_text())), ["stride1a", "stride1b"])


class ReadBalancesTests(unittest.TestCase):
    ATOM_DENOM = holders.SweepDenom(denom=ATOM, symbol="ATOM", decimals=6, price_usd=Decimal(4),
                                    destination="cosmoshub-4", channel="channel-0")
    STRD_DENOM = holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal(1),
                                    destination="osmosis-1", channel="channel-5")

    def run_read(self, owners: dict[str, list[tuple[str, str]]], page_size: int, refuse_pinned: bool = False, workers: int = 3,
                 duplicate_first_of_page: bool = False, short_page_at: str | None = None,
                 raise_at: str | None = None) -> tuple[dict, list[dict]]:
        calls: list[dict] = []
        self.calls = calls
        self.key_reads: list[str] = []

        def rest_get_all_pages(path: str, key: str, params: dict[str, str] | None = None,
                               base: str = config.REST) -> list:
            self.assertEqual((path, key), ("/cosmos/bank/v1beta1/denom_owners_by_query", "denom_owners"))
            self.key_reads.append(params["denom"])
            return [{"address": a, "balance": {"denom": params["denom"], "amount": v}} for a, v in owners[params["denom"]]]

        def rest_get(path: str, params: dict[str, str] | None = None, height: int | None = None,
                     base: str = config.REST) -> dict:
            calls.append({"params": dict(params), "height": height, "base": base})
            if raise_at is not None and params["pagination.offset"] == raise_at:
                raise chainio.ChainError("page failed")
            if raise_at is not None and params["pagination.offset"] != "0":
                time.sleep(0.05)  # a page after the failure is still queued or running when the error surfaces
            if refuse_pinned and height is not None:
                raise chainio.ChainError("height not available")
            if short_page_at is not None and params["pagination.offset"] == short_page_at:
                window = owners[params["denom"]][int(short_page_at):int(short_page_at) + page_size - 1]
                return {"denom_owners": [{"address": a, "balance": {"denom": params["denom"], "amount": v}}
                                         for a, v in window], "pagination": {"next_key": None}}
            offset = int(params["pagination.offset"])
            window = owners[params["denom"]][offset:offset + page_size]
            if duplicate_first_of_page and offset > 0:
                window = [owners[params["denom"]][offset - 1], *window]
            body = {"denom_owners": [{"address": address, "balance": {"denom": params["denom"], "amount": amount}}
                                     for address, amount in window],
                    "pagination": {"next_key": None}}
            if "pagination.count_total" in params:
                body["pagination"]["total"] = str(len(owners[params["denom"]]))
            return body

        with mock.patch.object(config, "PAGE_SIZE", page_size), \
                mock.patch.object(chainio, "latest_height", return_value=777), \
                mock.patch.object(chainio, "rest_get", side_effect=rest_get), \
                mock.patch.object(chainio, "rest_get_all_pages", side_effect=rest_get_all_pages), \
                mock.patch("builtins.print"):
            balances = holders.read_balances(denoms=[self.STRD_DENOM, self.ATOM_DENOM], workers=workers)
        return balances, calls

    def test_pages_are_assembled_across_denoms_pinned_to_one_height(self) -> None:
        owners = {"ustrd": [(f"a{index}", "5") for index in range(5)] + [("zero", "0")],
                  ATOM: [("a0", "9"), ("b", "2")]}
        balances, calls = self.run_read(owners=owners, page_size=2)

        self.assertEqual(balances["a4"], {"ustrd": 5})
        self.assertEqual(balances["a0"], {"ustrd": 5, ATOM: 9})
        self.assertEqual(balances["b"], {ATOM: 2})
        self.assertNotIn("zero", balances)
        self.assertEqual(len(balances), 6)
        self.assertEqual({call["height"] for call in calls}, {777})
        self.assertEqual({call["base"] for call in calls}, {config.BALANCES_REST})
        self.assertEqual(sorted(call["params"]["pagination.offset"] for call in calls if call["params"]["denom"] == "ustrd"),
                         ["0", "2", "4"])
        self.assertEqual(sum("pagination.count_total" in call["params"] for call in calls), 2)
        self.assertEqual(self.key_reads, [])  # complete pages: no key-paginated read

    def test_a_holder_repeated_across_pages_triggers_a_key_read_that_counts_once(self) -> None:
        owners = {"ustrd": [("a", "5"), ("b", "6"), ("c", "7")], ATOM: []}
        balances, _ = self.run_read(owners=owners, page_size=1, duplicate_first_of_page=True)
        self.assertEqual(balances, {"a": {"ustrd": 5}, "b": {"ustrd": 6}, "c": {"ustrd": 7}})
        self.assertEqual(self.key_reads, ["ustrd"])  # 5 rows for a total of 3

    def test_a_short_middle_page_rereads_only_that_denom_by_key(self) -> None:
        owners = {"ustrd": [(f"a{index}", "5") for index in range(6)], ATOM: [("b0", "2"), ("b1", "3")]}
        balances, _ = self.run_read(owners=owners, page_size=2, short_page_at="2")
        self.assertEqual(self.key_reads, ["ustrd"])
        self.assertEqual(len(balances), 8)
        self.assertEqual(balances["a2"], {"ustrd": 5})
        self.assertEqual(balances["b1"], {ATOM: 3})

    def test_a_repeated_address_with_a_matching_count_rereads_by_key(self) -> None:
        # Offsets drifted between pages: page 2 repeats a0 and the last page lost a5, so the count still equals the total
        drifted = {"0": [("a0", "5"), ("a1", "5")], "2": [("a0", "5"), ("a2", "5")], "4": [("a3", "5"), ("a4", "5")]}
        owners = {"ustrd": [(f"a{index}", "5") for index in range(6)], ATOM: []}

        def rest_get(path: str, params: dict[str, str] | None = None, height: int | None = None,
                     base: str = config.REST) -> dict:
            window = drifted[params["pagination.offset"]] if params["denom"] == "ustrd" else []
            body = {"denom_owners": [{"address": address, "balance": {"denom": params["denom"], "amount": amount}}
                                     for address, amount in window], "pagination": {"next_key": None}}
            if "pagination.count_total" in params:
                body["pagination"]["total"] = str(len(owners[params["denom"]]))
            return body

        def rest_get_all_pages(path: str, key: str, params: dict[str, str] | None = None,
                               base: str = config.REST) -> list:
            return [{"address": a, "balance": {"denom": params["denom"], "amount": v}} for a, v in owners[params["denom"]]]

        with mock.patch.object(config, "PAGE_SIZE", 2), mock.patch.object(chainio, "latest_height", return_value=777), \
                mock.patch.object(chainio, "rest_get", side_effect=rest_get), \
                mock.patch.object(chainio, "rest_get_all_pages", side_effect=rest_get_all_pages) as key_read, \
                mock.patch("builtins.print") as printed:
            balances = holders.read_balances(denoms=[self.STRD_DENOM, self.ATOM_DENOM], workers=2)

        self.assertEqual(key_read.call_args.kwargs["params"], {"denom": "ustrd"})
        self.assertEqual(sorted(balances), [f"a{index}" for index in range(6)])
        printed.assert_called_once_with("balances: ustrd: 6 rows over 3 pages for a total of 6 with a repeated address; "
                                        "re-reading by key")

    def test_a_refused_height_reads_every_denom_by_key(self) -> None:
        owners = {"ustrd": [("a", "5"), ("b", "6"), ("c", "7")], ATOM: [("a", "1")]}
        balances, calls = self.run_read(owners=owners, page_size=2, refuse_pinned=True)

        self.assertEqual(balances, {"a": {"ustrd": 5, ATOM: 1}, "b": {"ustrd": 6}, "c": {"ustrd": 7}})
        self.assertEqual(len(calls), 1)  # only the refused pinned first page
        self.assertEqual(sorted(self.key_reads), sorted(["ustrd", ATOM]))

    def test_a_refused_height_prints_one_line(self) -> None:
        owners = {"ustrd": [("a", "5")], ATOM: []}
        with mock.patch.object(chainio, "latest_height", return_value=777), \
                mock.patch.object(chainio, "rest_get", side_effect=chainio.ChainError("no")), \
                mock.patch.object(chainio, "rest_get_all_pages", return_value=[]), \
                mock.patch("builtins.print") as printed:
            holders.read_balances(denoms=[self.STRD_DENOM, self.ATOM_DENOM])
        printed.assert_called_once_with("balances: node refused height 777; reading every denom by key")

    def test_an_incomplete_denom_names_itself_in_one_line(self) -> None:
        owners = {"ustrd": [("a", "5"), ("b", "6"), ("c", "7")], ATOM: []}
        pages = [{"denom_owners": [{"address": "a", "balance": {"amount": "5"}}], "pagination": {"total": "3"}}]
        with mock.patch("builtins.print") as printed, \
                mock.patch.object(chainio, "rest_get_all_pages", return_value=[]):
            holders._checked_owners(sweep_denom=self.STRD_DENOM, pages=pages)
        printed.assert_called_once()
        self.assertIn("ustrd", printed.call_args.args[0])

    def test_a_failed_page_rereads_that_denom_by_key_and_cancels_its_queued_pages(self) -> None:
        owners = {"ustrd": [(f"a{index}", "5") for index in range(20)], ATOM: [("b0", "1"), ("b1", "3")]}
        balances, _ = self.run_read(owners=owners, page_size=2, workers=1, raise_at="2")

        self.assertEqual(self.key_reads, ["ustrd"])
        self.assertEqual(len(balances), 22)
        self.assertEqual(balances["a19"], {"ustrd": 5})
        self.assertEqual(balances["b1"], {ATOM: 3})
        ustrd_offsets = [call["params"]["pagination.offset"] for call in self.calls if call["params"]["denom"] == "ustrd"]
        self.assertIn("2", ustrd_offsets)
        self.assertNotIn("18", ustrd_offsets)


if __name__ == "__main__":
    unittest.main()
