import dataclasses
import json
import unittest
from typing import Any

import chain
import config
import funds

VAULT = "osmo1vault"
DELEGATION_ICA = "cosmos1delegation"
FEE_ICA = "cosmos1fee"
HOST_CHANNEL = "channel-141"
ATOM_ON_OSMOSIS = chain.ibc_denom(path="transfer/channel-0/uatom")
CANONICAL_STATOM = funds.canonical_st_denom(st_denom="stuatom")
HUB_ROUTE_STATOM = chain.ibc_denom(
    path="transfer/channel-0/transfer/channel-391/stuatom"
)
INJ_ON_OSMOSIS = chain.ibc_denom(path="transfer/channel-122/inj")
CANONICAL_STINJ = funds.canonical_st_denom(st_denom="stinj")

ATOM_ZONE = funds.ZoneDenoms(
    chain_id="cosmoshub-4",
    osmosis_denom=ATOM_ON_OSMOSIS,
    st_denom="stuatom",
    canonical_st_denom=CANONICAL_STATOM,
)
INJ_ZONE = funds.ZoneDenoms(
    chain_id="injective-1",
    osmosis_denom=INJ_ON_OSMOSIS,
    st_denom="stinj",
    canonical_st_denom=CANONICAL_STINJ,
)


def position(
    staked: int, unbonding: int = 0, entries: int = 0
) -> funds.StakingPosition:
    return funds.StakingPosition(
        staked=staked,
        validators=3,
        unbonding=funds.UnbondingSummary(
            amount=unbonding, entries=entries, earliest=None, latest=None
        ),
    )


def transfer(
    amount: int, status: funds.TransferStatus, denom: str = "uatom"
) -> funds.Transfer:
    return funds.Transfer(
        time="2026-09-30T12:00:00+00:00",
        height=100,
        ica=funds.IcaType.DELEGATION,
        sequence=1,
        amount=amount,
        denom=denom,
        status=status,
    )


def pool(
    native: int, st_amount: int, kind: funds.PoolKind = funds.PoolKind.CANONICAL
) -> funds.Pool:
    return funds.Pool(
        contract="osmo1pool",
        kind=kind,
        alloyed_denom="factory/osmo1pool/alloyed/x",
        native=native,
        st_denom="",
        st_amount=st_amount,
    )


def ica_balances(
    delegation: int = 0, withdrawal: int = 0, fee: int = 0, redemption: int = 0
) -> dict[funds.IcaType, dict[str, int]]:
    return {
        funds.IcaType.DELEGATION: {"uatom": delegation},
        funds.IcaType.WITHDRAWAL: {"uatom": withdrawal, "ibc/USDC": 5},
        funds.IcaType.FEE: {"uatom": fee},
        funds.IcaType.REDEMPTION: {"uatom": redemption},
    }


def send_packet_tx(
    height: int, sends: list[tuple[str, int, str, str, int, str]]
) -> dict[str, Any]:
    """A tx_search hit with one send_packet (and its ibc_transfer) per (channel, sequence, sender, receiver, amount, denom)."""
    events = []
    for channel, sequence, sender, receiver, amount, denom in sends:
        packet = json.dumps(
            {
                "denom": denom,
                "amount": str(amount),
                "sender": sender,
                "receiver": receiver,
            }
        )
        events.append(
            {
                "type": "send_packet",
                "attributes": [
                    {"key": "packet_data_hex", "value": packet.encode().hex()},
                    {"key": "packet_sequence", "value": str(sequence)},
                    {"key": "packet_src_port", "value": "transfer"},
                    {"key": "packet_src_channel", "value": channel},
                ],
            }
        )
        events.append(
            {"type": "ibc_transfer", "attributes": [{"key": "sender", "value": sender}]}
        )
    return {"height": str(height), "tx_result": {"events": events}}


def raw_pool(
    contract: str,
    assets: dict[str, int],
    base_denoms: dict[str, str],
    configured: list[str] | None = None,
) -> funds.RawPool:
    return funds.RawPool(
        contract=contract,
        alloyed_denom=f"factory/{contract}/alloyed/x",
        assets=configured if configured is not None else list(assets),
        balances=assets,
        base_denoms=base_denoms,
    )


class NeededAndCoverageTest(unittest.TestCase):
    def test_needed_is_supply_times_rate_truncated(self) -> None:
        self.assertEqual(
            funds.needed_amount(
                st_supply=1_000_000, redemption_rate="1.176363576955626723"
            ),
            1_176_363,
        )

    def test_needed_is_exact_at_eighteen_decimals(self) -> None:
        supply = 107_130_613_496_528_606_863_394_583  # the haqq stToken supply order of magnitude
        rate = "1.060496560022296837"

        needed = funds.needed_amount(st_supply=supply, redemption_rate=rate)

        self.assertEqual(needed, supply * 1_060_496_560_022_296_837 // 10**18)

    def test_covered_counts_vault_pool_native_and_swapped_in_sttokens_at_the_rate(
        self,
    ) -> None:
        covered = funds.covered_amount(
            vault=100,
            pools=[pool(native=250, st_amount=1000), pool(native=50, st_amount=0)],
            redemption_rate="2.5",
        )

        self.assertEqual(covered, 100 + 250 + 50 + 2500)

    def test_coverage_ratio_has_six_places(self) -> None:
        self.assertEqual(funds.coverage_ratio(covered=1, needed=3), "0.333333")
        self.assertEqual(funds.coverage_ratio(covered=3010, needed=3000), "1.003333")

    def test_coverage_is_null_when_nothing_is_needed(self) -> None:
        self.assertIsNone(funds.coverage_ratio(covered=5, needed=0))


class StagesTest(unittest.TestCase):
    def stages(self, **overrides: Any) -> funds.Stages:
        arguments: dict[str, Any] = {
            "position": position(staked=1000, unbonding=300),
            "ica_balances": ica_balances(
                delegation=40, withdrawal=8, fee=2, redemption=70
            ),
            "host_denom": "uatom",
            "transfers": [
                transfer(amount=25, status=funds.TransferStatus.IN_FLIGHT),
                transfer(amount=99, status=funds.TransferStatus.SETTLED),
            ],
            "vault": 500,
            "pools": [pool(native=600, st_amount=10), pool(native=100, st_amount=0)],
            "staketia": None,
        }
        return funds.build_stages(**{**arguments, **overrides})

    def test_plain_zone(self) -> None:
        stages = self.stages()

        self.assertEqual(stages.staked, 1000)
        self.assertEqual(stages.unbonding, 300)
        self.assertEqual(stages.liquid, 50)  # the redemption ICA is not part of the bar
        self.assertEqual(stages.in_flight, 25)  # settled transfers are not in flight
        self.assertEqual(stages.vault, 500)
        self.assertEqual(stages.pools, 700)

    def test_in_flight_counts_only_the_host_denom(self) -> None:
        stages = self.stages(
            transfers=[
                transfer(amount=25, status=funds.TransferStatus.IN_FLIGHT),
                transfer(
                    amount=400, status=funds.TransferStatus.IN_FLIGHT, denom="ibc/USDC"
                ),
            ]
        )

        self.assertEqual(stages.in_flight, 25)

    def test_in_flight_is_null_when_the_index_lookup_failed(self) -> None:
        self.assertIsNone(self.stages(transfers=None).in_flight)

    def test_osmosis_side_is_null_when_osmosis_could_not_be_read(self) -> None:
        stages = self.stages(vault=None, pools=None)

        self.assertIsNone(stages.vault)
        self.assertIsNone(stages.pools)

    def test_celestia_adds_the_multisig_and_the_claim_address(self) -> None:
        staketia = funds.Staketia(
            multisig_address="celestia1multisig",
            staked=7,
            unbonding=11,
            unbonding_entries=2,
            unbonding_earliest=None,
            unbonding_latest=None,
            liquid=13,
            claim_address="stride1claim",
            claim_balance=17,
            unbonding_records=None,
        )

        stages = self.stages(staketia=staketia)

        self.assertEqual(stages.staked, 1007)
        self.assertEqual(stages.unbonding, 311)
        self.assertEqual(stages.liquid, 50 + 13 + 17)


class UnbondingSummaryTest(unittest.TestCase):
    def test_totals_and_time_range_across_validators(self) -> None:
        responses = [
            {
                "entries": [
                    {"balance": "10", "completion_time": "2026-10-05T00:00:00Z"},
                    {"balance": "20", "completion_time": "2026-10-01T12:30:00.5Z"},
                ]
            },
            {"entries": [{"balance": "5", "completion_time": "2026-10-03T00:00:00Z"}]},
        ]

        summary = funds.summarize_unbonding(unbonding_responses=responses)

        self.assertEqual(summary.amount, 35)
        self.assertEqual(summary.entries, 3)
        self.assertEqual(summary.earliest, "2026-10-01T12:30:00.500000+00:00")
        self.assertEqual(summary.latest, "2026-10-05T00:00:00+00:00")

    def test_no_entries(self) -> None:
        summary = funds.summarize_unbonding(unbonding_responses=[])

        self.assertEqual(
            summary,
            funds.UnbondingSummary(amount=0, entries=0, earliest=None, latest=None),
        )


class ValidatorPositionsTest(unittest.TestCase):
    def test_rows_join_delegations_and_entries_largest_stake_first(self) -> None:
        delegations = [
            {"delegation": {"validator_address": "valA"}, "balance": {"amount": "100"}},
            {"delegation": {"validator_address": "valB"}, "balance": {"amount": "900"}},
            {"delegation": {"validator_address": "valZero"}, "balance": {"amount": "0"}},
        ]
        unbonding = [
            {
                "validator_address": "valA",
                "entries": [
                    {"balance": "7", "completion_time": "2026-10-09T00:00:00Z"},
                    {"balance": "5", "completion_time": "2026-10-03T00:00:00Z"},
                ],
            },
            {"validator_address": "valOnlyUnbonding", "entries": [{"balance": "2", "completion_time": "2026-10-04T00:00:00Z"}]},
        ]

        rows = funds.validator_positions(
            delegations=delegations,
            unbonding_responses=unbonding,
            names={"valA": "Alpha", "valB": "Beta"},
            source=funds.StakeSource.ICA,
        )

        self.assertEqual([row.address for row in rows], ["valB", "valA", "valOnlyUnbonding"])
        self.assertEqual([row.moniker for row in rows], ["Beta", "Alpha", "valOnlyUnbonding"])
        self.assertEqual([row.staked for row in rows], [900, 100, 0])
        # Entries are ordered by completion, and a validator with no delegation but an entry still appears.
        self.assertEqual(
            rows[1].entries,
            [
                funds.UnbondingEntry(amount=5, completion="2026-10-03T00:00:00+00:00"),
                funds.UnbondingEntry(amount=7, completion="2026-10-09T00:00:00+00:00"),
            ],
        )
        self.assertEqual(rows[2].entries, [funds.UnbondingEntry(amount=2, completion="2026-10-04T00:00:00+00:00")])
        self.assertTrue(all(row.source == funds.StakeSource.ICA for row in rows))

    def test_nothing_staked_and_nothing_unbonding_is_omitted(self) -> None:
        rows = funds.validator_positions(
            delegations=[{"delegation": {"validator_address": "valZero"}, "balance": {"amount": "0"}}],
            unbonding_responses=[{"validator_address": "valEmpty", "entries": []}],
            names={},
            source=funds.StakeSource.MULTISIG,
        )
        self.assertEqual(rows, [])


class TransfersTest(unittest.TestCase):
    def parse(self, txs: list[dict[str, Any]]) -> list[funds.SentTransfer]:
        return funds.parse_sent_transfers(
            txs=txs,
            ica=funds.IcaType.DELEGATION,
            sender=DELEGATION_ICA,
            receiver=VAULT,
            channel=HOST_CHANNEL,
        )

    def test_reads_sequence_amount_and_denom_from_the_packet(self) -> None:
        sent = self.parse(
            [
                send_packet_tx(
                    height=500,
                    sends=[
                        (HOST_CHANNEL, 42, DELEGATION_ICA, VAULT, 1_000_000, "uatom")
                    ],
                )
            ]
        )

        self.assertEqual(
            sent,
            [
                funds.SentTransfer(
                    height=500,
                    ica=funds.IcaType.DELEGATION,
                    sequence=42,
                    amount=1_000_000,
                    denom="uatom",
                )
            ],
        )

    def test_a_relayer_tx_may_carry_several_sends(self) -> None:
        tx = send_packet_tx(
            height=500,
            sends=[
                (HOST_CHANNEL, 42, DELEGATION_ICA, VAULT, 1, "uatom"),
                (HOST_CHANNEL, 43, DELEGATION_ICA, VAULT, 2, "uatom"),
            ],
        )

        self.assertEqual([sent.sequence for sent in self.parse([tx])], [42, 43])

    def test_ignores_other_channels_receivers_and_senders(self) -> None:
        txs = [
            send_packet_tx(
                height=1, sends=[("channel-999", 1, DELEGATION_ICA, VAULT, 1, "uatom")]
            ),
            send_packet_tx(
                height=2,
                sends=[
                    (HOST_CHANNEL, 2, DELEGATION_ICA, "stride1somewhere", 1, "uatom")
                ],
            ),
            send_packet_tx(
                height=3, sends=[(HOST_CHANNEL, 3, FEE_ICA, VAULT, 1, "uatom")]
            ),
        ]

        self.assertEqual(self.parse(txs), [])

    def test_marks_committed_sequences_in_flight_and_orders_newest_first(self) -> None:
        sent = [
            funds.SentTransfer(
                height=100, ica=funds.IcaType.FEE, sequence=7, amount=1, denom="uatom"
            ),
            funds.SentTransfer(
                height=300,
                ica=funds.IcaType.DELEGATION,
                sequence=9,
                amount=2,
                denom="uatom",
            ),
            funds.SentTransfer(
                height=200,
                ica=funds.IcaType.DELEGATION,
                sequence=8,
                amount=3,
                denom="uatom",
            ),
        ]
        times = {100: "t100", 200: "t200", 300: "t300"}

        marked = funds.mark_transfers(sent=sent, committed={8}, times=times)

        self.assertEqual([transfer.sequence for transfer in marked], [9, 8, 7])
        self.assertEqual(
            [transfer.status for transfer in marked],
            [
                funds.TransferStatus.SETTLED,
                funds.TransferStatus.IN_FLIGHT,
                funds.TransferStatus.SETTLED,
            ],
        )
        self.assertEqual(
            [transfer.time for transfer in marked], ["t300", "t200", "t100"]
        )
        self.assertEqual(marked[1].amount, 3)


class PoolsTest(unittest.TestCase):
    def test_discovers_contracts_from_alloyed_receipts_plus_extras(self) -> None:
        balances = {
            "factory/osmo1canon/alloyed/stATOMcanon": 10,
            "factory/osmo1other/nonalloyed": 1,
            ATOM_ON_OSMOSIS: 5,
            "factory/osmo1hub/alloyed/stATOMhub": 3,
        }

        contracts = funds.discover_pool_contracts(
            vault_balances=balances, extra=("osmo1new", "osmo1hub")
        )

        self.assertEqual(contracts, ["osmo1canon", "osmo1hub", "osmo1new"])

    def test_canonical_and_route_pools_go_to_the_zone_whose_native_denom_they_hold(
        self,
    ) -> None:
        pools = [
            raw_pool(
                "osmo1canon",
                {ATOM_ON_OSMOSIS: 1000, CANONICAL_STATOM: 40},
                {CANONICAL_STATOM: "stuatom"},
            ),
            raw_pool(
                "osmo1hub",
                {ATOM_ON_OSMOSIS: 200, HUB_ROUTE_STATOM: 7},
                {HUB_ROUTE_STATOM: "stuatom"},
            ),
            raw_pool(
                "osmo1inj",
                {INJ_ON_OSMOSIS: 5, CANONICAL_STINJ: 0},
                {CANONICAL_STINJ: "stinj"},
            ),
        ]

        assigned = funds.classify_pools(pools=pools, zones=[ATOM_ZONE, INJ_ZONE])

        self.assertEqual(
            [
                (pool.contract, pool.kind, pool.native, pool.st_amount)
                for pool in assigned["cosmoshub-4"]
            ],
            [
                ("osmo1canon", funds.PoolKind.CANONICAL, 1000, 40),
                ("osmo1hub", funds.PoolKind.ROUTE, 200, 7),
            ],
        )
        self.assertEqual(assigned["cosmoshub-4"][1].st_denom, HUB_ROUTE_STATOM)
        self.assertEqual(
            [
                (pool.contract, pool.kind, pool.native)
                for pool in assigned["injective-1"]
            ],
            [("osmo1inj", funds.PoolKind.CANONICAL, 5)],
        )

    def test_a_freshly_funded_pool_is_classified_from_its_configured_assets(
        self,
    ) -> None:
        # Nothing has been swapped in, so the bank holds only the native token; the asset config still names the stToken.
        pools = [
            raw_pool(
                "osmo1canon",
                {ATOM_ON_OSMOSIS: 1000},
                {CANONICAL_STATOM: "stuatom"},
                configured=[ATOM_ON_OSMOSIS, CANONICAL_STATOM],
            )
        ]

        assigned = funds.classify_pools(pools=pools, zones=[ATOM_ZONE])

        self.assertEqual(assigned["cosmoshub-4"][0].kind, funds.PoolKind.CANONICAL)
        self.assertEqual(assigned["cosmoshub-4"][0].st_amount, 0)

    def test_a_drained_pool_still_belongs_to_its_zone_through_its_sttoken(self) -> None:
        # The contract drops the native asset once its balance hits zero; the stTokens swapped in still count.
        pools = [
            raw_pool("osmo1hub", {HUB_ROUTE_STATOM: 900}, {HUB_ROUTE_STATOM: "stuatom"})
        ]

        assigned = funds.classify_pools(pools=pools, zones=[ATOM_ZONE, INJ_ZONE])

        self.assertEqual(
            [
                (pool.kind, pool.native, pool.st_amount)
                for pool in assigned["cosmoshub-4"]
            ],
            [(funds.PoolKind.ROUTE, 0, 900)],
        )
        self.assertEqual(assigned["injective-1"], [])

    def test_a_pool_for_no_zone_is_ignored(self) -> None:
        usdc = chain.ibc_denom(path="transfer/channel-208/uusdc")
        pools = [raw_pool("osmo1allusdc", {usdc: 5}, {usdc: "uusdc"})]

        assigned = funds.classify_pools(pools=pools, zones=[ATOM_ZONE])

        self.assertEqual(assigned, {"cosmoshub-4": []})


class PayloadTest(unittest.TestCase):
    def test_every_int_becomes_a_string_but_bools_and_none_survive(self) -> None:
        payload = funds._stringify_ints(
            {"a": 10**26, "b": [1, {"c": None}], "d": True, "e": "x"}
        )

        self.assertEqual(
            payload, {"a": str(10**26), "b": ["1", {"c": None}], "d": True, "e": "x"}
        )

    def test_zone_denoms_follow_the_osmosis_side_channel(self) -> None:
        side = host_side(
            chain_id="cosmoshub-4", host_denom="uatom", osmosis_channel="channel-0"
        )

        denoms = funds._zone_denoms(host_side=side)

        self.assertEqual(denoms, ATOM_ZONE)

    def test_osmosis_itself_uses_the_bare_denom(self) -> None:
        side = host_side(chain_id="osmosis-1", host_denom="uosmo", osmosis_channel=None)

        self.assertEqual(funds._zone_denoms(host_side=side).osmosis_denom, "uosmo")

    def test_no_denoms_without_the_osmosis_channel(self) -> None:
        side = host_side(
            chain_id="cosmoshub-4", host_denom="uatom", osmosis_channel=None
        )

        self.assertIsNone(funds._zone_denoms(host_side=side))


class ZoneAssemblyTest(unittest.TestCase):
    def test_zone_without_osmosis_has_null_coverage_and_merges_the_staketia_times(self) -> None:
        staketia = funds.Staketia(
            multisig_address="celestia1multisig",
            staked=100,
            unbonding=50,
            unbonding_entries=1,
            unbonding_earliest="2026-10-01T00:00:00+00:00",
            unbonding_latest="2026-10-09T00:00:00+00:00",
            liquid=0,
            claim_address="stride1claim",
            claim_balance=0,
            unbonding_records=None,
        )
        position = funds.StakingPosition(
            staked=10,
            validators=1,
            unbonding=funds.UnbondingSummary(
                amount=5, entries=2, earliest="2026-10-03T00:00:00+00:00", latest="2026-10-05T00:00:00+00:00"
            ),
        )
        side = funds.HostSide(
            zone=config.ZONES_BY_CHAIN_ID["celestia"],
            host_zone={"host_denom": "utia", "redemption_rate": "1.5", "deposit_address": "stride1deposit", **ica_host_zone()},
            st_supply=1000,
            osmosis_channel="channel-6994",
            ica_balances={ica: {"utia": index} for index, ica in enumerate(funds.IcaType, start=1)},
            position=position,
            transfers=[],
            deposit_balance=0,
            staketia=staketia,
            staketia_accounts=[],
        )

        zone = funds._zone_funds(side=side, osmosis=None, pools_by_zone=None, open_records={"celestia": 4})

        self.assertEqual(zone.needed, 1500)
        self.assertIsNone(zone.covered)
        self.assertIsNone(zone.coverage)
        self.assertIsNone(zone.stages.vault)
        self.assertIsNone(zone.pools)
        self.assertEqual(zone.stages.staked, 110)
        self.assertEqual(zone.unbonding_entries, 3)
        self.assertEqual(zone.unbonding_earliest, "2026-10-01T00:00:00+00:00")
        self.assertEqual(zone.unbonding_latest, "2026-10-09T00:00:00+00:00")
        self.assertEqual(zone.ica_liquid[funds.IcaType.REDEMPTION], 4)
        self.assertEqual(zone.redemption_ica_balance, 4)
        self.assertEqual(zone.open_redemption_records, 4)
        self.assertEqual(
            [account.name for account in zone.accounts][:5],
            ["Delegation ICA", "Withdrawal ICA", "Fee ICA", "Redemption ICA", "Deposit address"],
        )
        self.assertEqual(zone.accounts[-1].name, "Osmosis vault")
        self.assertIsNone(zone.accounts[-1].liquid)

    def test_ica_other_balances_keep_non_host_non_zero_denoms(self) -> None:
        side = host_side(chain_id="celestia", host_denom="utia", osmosis_channel=None)
        side = dataclasses.replace(
            side,
            host_zone={"host_denom": "utia", "redemption_rate": "1", "deposit_address": "stride1deposit", **ica_host_zone()},
            ica_balances={
                **side.ica_balances,
                funds.IcaType.WITHDRAWAL: {"utia": 2, "ibc/USDC": 5, "ibc/ZERO": 0},
            },
        )

        zone = funds._zone_funds(side=side, osmosis=None, pools_by_zone=None, open_records={})

        self.assertEqual(zone.accounts[1].other_balances, [funds.Balance(denom="ibc/USDC", amount=5)])
        self.assertEqual(zone.accounts[0].other_balances, [])


def ica_host_zone() -> dict[str, str]:
    return {f"{ica.lower()}_ica_address": f"celestia1{ica.lower()}" for ica in funds.IcaType}


def host_side(
    chain_id: str, host_denom: str, osmosis_channel: str | None
) -> funds.HostSide:
    return funds.HostSide(
        zone=config.ZONES_BY_CHAIN_ID[chain_id],
        host_zone={"host_denom": host_denom},
        st_supply=0,
        osmosis_channel=osmosis_channel,
        ica_balances=ica_balances(),
        position=position(staked=0),
        transfers=[],
        deposit_balance=0,
        staketia=None,
        staketia_accounts=[],
    )


if __name__ == "__main__":
    unittest.main()
