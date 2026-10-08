import json
import unittest
from unittest import mock

import config
import multisig
import server

ADDRESS = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
VAULT = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"
MODERATOR = "osmo1ugrn8qgsvyr8zwrv8h2g4r8ascngxk7qeaz7e0htjq3znswkh4cqhjdpgy"
OSMOSIS_NODE = "https://osmosis-strd-rpc.polkachu.com:443"
CELESTIA_VALOPER = "celestiavaloper1q2kaajedxm0r5xc0twdqz6atap96502d67yjyj"
NODE = "https://stride-strd-rpc.polkachu.com:443"
CELESTIA, COSMOS, DYDX, *_ = (zone.chain_id for zone in config.ZONES)


def zone_entry(chain_id: str, **overrides: object) -> dict[str, object]:
    return {"kind": "zone", "chain_id": chain_id, "live_test_pick": None, "live_test_reason": "no pick today", **overrides}


def pick(address: str, moniker: str, recorded: int) -> dict[str, str]:
    return {"address": address, "moniker": moniker, "recorded": str(recorded)}


def fake_snapshot() -> dict[str, object]:
    return {
        "zones": [
            zone_entry(CELESTIA, live_test_pick=pick(CELESTIA_VALOPER, "mhventures", 15_861_063), live_test_reason=None),
            zone_entry(COSMOS),
            {"chain_id": DYDX, "error": "URLError: boom"},
        ]
    }


def sets_by_id(
    validators_data: dict[str, object] | None,
    funds_data: dict[str, object] | None = None,
    pools_data: dict[str, object] | None = None,
) -> dict[str, multisig.TxSet]:
    tx_sets = multisig.tx_sets(validators_data=validators_data, funds_data=funds_data, pools_data=pools_data)
    return {tx_set.id: tx_set for tx_set in tx_sets}


def tx_for(tx_set: multisig.TxSet, chain_id: str) -> multisig.MultisigTx:
    return next(tx for tx in tx_set.txs if tx.chain_id == chain_id)


class LiveTestCommandsTest(unittest.TestCase):
    def test_celestia_commands_are_exact(self) -> None:
        tx = tx_for(sets_by_id(fake_snapshot())["live-test-undelegate"], CELESTIA)
        stem = "/tmp/wind-down/live-test-celestia"

        self.assertEqual(
            [(command.tag, command.text) for command in tx.commands],
            [
                (
                    "anyone",
                    "mkdir -p /tmp/wind-down\n"
                    f"echo '[{{\"address\": \"{CELESTIA_VALOPER}\", \"offset\": \"0\"}}]' > {stem}.json\n"
                    f"strided tx stakeibc undelegate-from-validators celestia {stem}.json --from {ADDRESS} --generate-only \\\n"
                    f"  --chain-id stride-1 --node {NODE} --gas 12000000 --fees 60000ustrd > {stem}.unsigned.json",
                ),
                *(
                    (
                        tag,
                        f"strided tx sign {stem}.unsigned.json --multisig {ADDRESS} --from {key} "
                        f"--chain-id stride-1 --node {NODE} \\\n  --output-document {stem}.{key}.json",
                    )
                    for tag, key in (("Sam", "FS5"), ("Aidan", "FA5"), ("Riley", "FR5"))
                ),
                (
                    "Sam",
                    f"strided tx multisign {stem}.unsigned.json F5 {stem}.FS5.json ~/Downloads/{stem.rsplit('/', 1)[-1]}.FA5.json "
                    f"--chain-id stride-1 --node {NODE} --output-document {stem}.signed.json\n"
                    f"strided tx broadcast {stem}.signed.json --node {NODE} --broadcast-mode sync",
                ),
            ],
        )
        self.assertTrue(tx.ready)
        self.assertIsNone(tx.reason)
        self.assertEqual(tx.title, "celestia · live test: mhventures (celestiavaloper1q2k…), 15,861,063 utia")
        self.assertEqual(tx.files, [f"{stem}.unsigned.json", f"{stem}.FS5.json", f"{stem}.FA5.json", f"{stem}.FR5.json"])

    def test_every_sign_command_says_the_signer_needs_the_multisig_key_in_their_keyring(self) -> None:
        all_sets = sets_by_id(fake_snapshot(), funds_data=fake_funds(), pools_data=fake_pools())
        for tx_set in all_sets.values():
            binary, address = ("osmosisd", VAULT) if tx_set.id in ("pool-funding", "pool-creation") else ("strided", ADDRESS)
            for tx in tx_set.txs:
                sign_commands = [command for command in tx.commands if "tx sign" in command.text]

                self.assertEqual(len(sign_commands), 0 if not tx.commands else 3)
                for command in sign_commands:
                    self.assertIn(f"needs the F5 multisig key in your keyring: {binary} keys show {address}", command.label)

    def test_riley_is_labelled_as_the_backup(self) -> None:
        commands = tx_for(sets_by_id(fake_snapshot())["live-test-undelegate"], CELESTIA).commands

        riley = next(command for command in commands if command.tag == "Riley")
        self.assertIn("Backup", riley.label)
        self.assertNotIn("Backup", next(command for command in commands if command.tag == "Aidan" and "sign " in command.text).label)

    def test_zone_without_a_pick_is_not_ready_but_keeps_the_placeholder(self) -> None:
        tx = tx_for(sets_by_id(fake_snapshot())["live-test-undelegate"], COSMOS)

        self.assertEqual((tx.ready, tx.reason), (False, "no pick today"))
        self.assertIn('"address": "<LIVE_TEST_VALOPER>"', tx.commands[0].text)
        self.assertIn("--gas 12000000 --fees 60000ustrd", tx.commands[0].text)

    def test_zone_with_an_error_is_not_ready_with_that_error(self) -> None:
        tx = tx_for(sets_by_id(fake_snapshot())["live-test-undelegate"], DYDX)

        self.assertEqual((tx.ready, tx.reason), (False, "URLError: boom"))

    def test_zone_missing_from_the_snapshot_is_not_ready(self) -> None:
        tx = tx_for(sets_by_id({"zones": []})["live-test-undelegate"], CELESTIA)

        self.assertFalse(tx.ready)
        self.assertEqual(tx.reason, "zone missing from the Validators snapshot")

    def test_eighteen_decimal_amount_is_shown_in_whole_tokens(self) -> None:
        injective = next(zone for zone in config.ZONES if zone.decimals == 18)
        data = {"zones": [zone_entry(injective.chain_id, live_test_pick=pick("injvaloper1abcdef", "m", 2_500_000_000_000_000_000))]}

        tx = tx_for(sets_by_id(data)["live-test-undelegate"], injective.chain_id)

        self.assertEqual(tx.title, f"{injective.chain_id} · live test: m (injvaloper1abc…), 2.5 {injective.symbol}")


class FullDrainCommandsTest(unittest.TestCase):
    def test_celestia_generate_command_is_exact(self) -> None:
        tx = tx_for(sets_by_id(fake_snapshot())["full-drain"], CELESTIA)
        stem = "/tmp/wind-down/full-drain-celestia"

        self.assertEqual(
            tx.commands[0].text,
            "mkdir -p /tmp/wind-down\n"
            f"strided tx stakeibc undelegate-from-validators celestia --all --from {ADDRESS} --generate-only \\\n"
            f"  --chain-id stride-1 --node {NODE} --gas 13000000 --fees 65000ustrd > {stem}.unsigned.json",
        )
        self.assertIn(f"{stem}.signed.json", tx.commands[-1].text)
        self.assertTrue(tx.ready)

    def test_gas_and_fees_per_zone(self) -> None:
        full_drain = sets_by_id(fake_snapshot())["full-drain"]

        flags = {tx.chain_id: tx.commands[0].text.split("--gas ")[1].split(" >")[0] for tx in full_drain.txs}

        self.assertEqual(flags["cosmoshub-4"], "25000000 --fees 125000ustrd")
        self.assertEqual(flags["osmosis-1"], "15000000 --fees 75000ustrd")
        self.assertEqual(flags["ssc-1"], "15000000 --fees 75000ustrd")
        others = {chain_id: gas for chain_id, gas in flags.items() if chain_id not in ("cosmoshub-4", "osmosis-1", "ssc-1")}
        self.assertEqual(set(others.values()), {"13000000 --fees 65000ustrd"})

    def test_a_zone_error_blocks_the_full_drain_but_a_missing_pick_does_not(self) -> None:
        full_drain = sets_by_id(fake_snapshot())["full-drain"]

        self.assertEqual(tx_for(full_drain, DYDX).reason, "URLError: boom")
        self.assertTrue(tx_for(full_drain, COSMOS).ready)


class SetsTest(unittest.TestCase):
    def test_no_snapshot_leaves_every_drain_tx_not_ready(self) -> None:
        for tx_set in multisig.tx_sets(validators_data=None, funds_data=None, pools_data=None)[1:3]:
            self.assertEqual([tx.chain_id for tx in tx_set.txs], [zone.chain_id for zone in config.ZONES])
            for tx in tx_set.txs:
                self.assertEqual((tx.ready, tx.reason), (False, "waiting for the Validators snapshot"))
                self.assertEqual(len(tx.commands), 5)

    def test_ids_are_unique_and_tied_to_their_ops_steps(self) -> None:
        tx_sets = multisig.tx_sets(validators_data=None, funds_data=None, pools_data=None)

        self.assertEqual(
            [(tx_set.id, tx_set.step_id) for tx_set in tx_sets],
            [
                ("pool-creation", "vote-pools-create"),
                ("live-test-undelegate", "drain-live-test"),
                ("full-drain", "drain-rest"),
                ("ica-transfers", "transfers"),
                ("staketia-claim-balance", "tia-claim-balance"),
                ("pool-funding", "join-pools"),
            ],
        )

    def test_payload_is_json_serialisable_with_the_documented_shape(self) -> None:
        payload = multisig.tx_sets(validators_data=fake_snapshot(), funds_data=None, pools_data=None)[1].payload()

        json.dumps(payload)
        self.assertEqual(set(payload), {"id", "step_id", "title", "description", "notes", "txs"})
        self.assertEqual(set(payload["txs"][0]), {"chain_id", "title", "ready", "reason", "commands", "files", "members", "done", "blocked", "seeded"})
        self.assertEqual(set(payload["txs"][0]["commands"][0]), {"tag", "label", "text"})

    def test_multisig_constants(self) -> None:
        self.assertEqual((multisig.MULTISIG_KEY, multisig.MULTISIG_ADDRESS), ("F5", config.PROTOCOL_ADMIN))
        self.assertEqual([signer.key for signer in multisig.DEFAULT_SIGNERS], ["FS5", "FA5"])
        self.assertEqual(multisig.BROADCASTER.tag, "Sam")


# ---- transfer-day sets

STRIDE_NODE = NODE
FOREIGN_DENOM = "ibc/ABCDEF1A2B3C"
OSMOSIS_DENOM = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
CANONICAL_CONTRACT = "osmo1canonicalcontract"
ROUTE_CONTRACT = "osmo1routecontractabc123"


def fake_funds() -> dict[str, object]:
    return {
        "zones": [
            {"chain_id": CELESTIA, "error": "URLError: celestia down"},
            {
                "chain_id": COSMOS,
                "host_denom": "uatom",
                "ica_balances": {
                    "FEE": {"uatom": "2500000"},
                    "WITHDRAWAL": {"uatom": "400000"},  # below the one-token test amount
                    "DELEGATION": {"uatom": "90000000", "ibc/ZEROED": "0"},
                    "REDEMPTION": {},
                },
            },
            {
                "chain_id": DYDX,
                "host_denom": "adydx",
                "ica_balances": {
                    "FEE": {"adydx": "3000000000000000000"},
                    "WITHDRAWAL": {"adydx": "2000000000000000000", FOREIGN_DENOM: "12345678", "ibc/DUST": "0"},
                    "DELEGATION": {"adydx": "1000000000000000000"},
                    "REDEMPTION": {"adydx": "1000000000000000000"},
                },
            },
        ]
    }


def pool_entry(
    contract: str, kind: str, pool_id: str | None, allocation: str | None, funded_exactly: bool | None, vault_shares: str = "0"
) -> dict[str, object]:
    return {
        "contract": contract,
        "kind": kind,
        "pool_id": pool_id,
        "allocation": allocation,
        "funded_exactly": funded_exactly,
        "vault_shares": vault_shares,
    }


def fake_pools(**overrides: object) -> dict[str, object]:
    # The test joins have landed: the vault holds one token of shares in each fundable pool, so the rest joins are ready.
    pools = [
        # Route first on purpose: the set puts the canonical pool first whatever the payload order.
        pool_entry(ROUTE_CONTRACT, "route", None, "3000000", False, vault_shares="1000000"),
        pool_entry("osmo1unrecognisedcontract", "unrecognised", "9", None, None),
        pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "5000000", False, vault_shares="1000000"),
    ]
    return {
        "zones": [
            {
                "chain_id": COSMOS,
                "osmosis_denom": OSMOSIS_DENOM,
                "pools": pools,
                "planned": fake_planned(),
                "creation_fee": CREATION_FEE,
                "vault_fee_balance": "40000000",
                "creation_fee_short": False,
                **overrides,
            },
            {"chain_id": DYDX, "error": "URLError: osmosis down"},
            {"chain_id": "juno-1", "osmosis_denom": OSMOSIS_DENOM, "pools": []},
        ]
    }


def expected_commands(
    generate: str, stem: str, binary: str = "strided", address: str = ADDRESS, chain_id: str = "stride-1", node: str = STRIDE_NODE
) -> list[tuple[str, str]]:
    return [
        ("anyone", generate),
        *(
            (
                tag,
                f"{binary} tx sign {stem}.unsigned.json --multisig {address} --from {key} "
                f"--chain-id {chain_id} --node {node} \\\n  --output-document {stem}.{key}.json",
            )
            for tag, key in (("Sam", "FS5"), ("Aidan", "FA5"), ("Riley", "FR5"))
        ),
        (
            "Sam",
            f"{binary} tx multisign {stem}.unsigned.json F5 {stem}.FS5.json ~/Downloads/{stem.rsplit('/', 1)[-1]}.FA5.json "
            f"--chain-id {chain_id} --node {node} --output-document {stem}.signed.json\n"
            f"{binary} tx broadcast {stem}.signed.json --node {node} --broadcast-mode sync",
        ),
    ]


def command_pairs(tx: multisig.MultisigTx) -> list[tuple[str, str]]:
    return [(command.tag, command.text) for command in tx.commands]


def transfer_txs(chain_id: str, funds_data: dict[str, object] | None = None) -> list[multisig.MultisigTx]:
    tx_set = sets_by_id(fake_snapshot(), funds_data=fake_funds() if funds_data is None else funds_data)["ica-transfers"]
    return [tx for tx in tx_set.txs if tx.chain_id == chain_id]


def transfer_generate(zone: str, ica: str, amount: str, label: str) -> str:
    stem = f"/tmp/wind-down/transfer-{zone}-{ica.lower()}-{label}"
    return (
        "mkdir -p /tmp/wind-down\n"
        f"strided tx stakeibc transfer-from-ica {zone} {ica} {amount} --from {ADDRESS} --generate-only "
        f"--chain-id stride-1 --node {NODE} --gas 600000 --fees 3000ustrd > {stem}.unsigned.json"
    )


class IcaTransfersTest(unittest.TestCase):
    def test_test_and_rest_pair_are_exact(self) -> None:
        test_tx, rest_tx = transfer_txs(COSMOS)[:2]
        test_stem = "/tmp/wind-down/transfer-cosmoshub-4-fee-test"
        rest_stem = "/tmp/wind-down/transfer-cosmoshub-4-fee-rest"

        self.assertEqual(
            command_pairs(test_tx),
            expected_commands(transfer_generate(COSMOS, "FEE", "1000000uatom", "test"), stem=test_stem),
        )
        self.assertEqual(
            command_pairs(rest_tx),
            expected_commands(transfer_generate(COSMOS, "FEE", "2500000uatom", "rest"), stem=rest_stem),
        )
        self.assertEqual(test_tx.title, "cosmoshub-4 · FEE ICA · test: 1000000uatom")
        self.assertEqual(rest_tx.title, "cosmoshub-4 · FEE ICA · rest: 2500000uatom (live balance, copy after the test has settled)")
        self.assertEqual(test_tx.files, [f"{test_stem}.unsigned.json", f"{test_stem}.FS5.json", f"{test_stem}.FA5.json", f"{test_stem}.FR5.json"])
        self.assertEqual([(tx.ready, tx.reason) for tx in (test_tx, rest_tx)], [(True, None)] * 2)

    def test_eighteen_decimal_test_amount_is_one_whole_token(self) -> None:
        test_tx = transfer_txs(DYDX)[0]

        self.assertIn("transfer-from-ica dydx-mainnet-1 FEE 1000000000000000000adydx ", test_tx.commands[0].text)

    def test_foreign_denom_gets_its_own_tx_with_the_full_balance_after_its_ica(self) -> None:
        txs = transfer_txs(DYDX)
        withdrawal = [tx for tx in txs if "WITHDRAWAL ICA" in tx.title]
        stem = "/tmp/wind-down/transfer-dydx-mainnet-1-withdrawal-denom-d3cc7f0d"

        self.assertEqual(len(withdrawal), 3)  # test, rest, the one non-zero foreign denom
        foreign = withdrawal[-1]
        self.assertEqual(
            command_pairs(foreign),
            expected_commands(transfer_generate(DYDX, "WITHDRAWAL", f"12345678{FOREIGN_DENOM}", "denom-d3cc7f0d"), stem=stem),
        )
        self.assertEqual(foreign.title, f"dydx-mainnet-1 · WITHDRAWAL ICA · foreign denom: 12345678{FOREIGN_DENOM} (full balance)")
        self.assertTrue(foreign.ready)
        self.assertFalse(any("ibc/DUST" in command.text for tx in txs for command in tx.commands))

    def test_zero_foreign_balance_and_the_host_denom_do_not_make_foreign_txs(self) -> None:
        delegation = [tx for tx in transfer_txs(COSMOS) if "DELEGATION ICA" in tx.title]

        self.assertEqual(len(delegation), 2)

    def test_per_zone_order_is_fee_withdrawal_delegation_redemption_test_then_rest(self) -> None:
        titles = [tx.title.removeprefix("cosmoshub-4 · ").split(":")[0] for tx in transfer_txs(COSMOS)]

        self.assertEqual(
            titles,
            [f"{ica} ICA · {kind}" for ica in ("FEE", "WITHDRAWAL", "DELEGATION", "REDEMPTION") for kind in ("test", "rest")],
        )

    def test_balance_below_the_test_amount_is_not_ready_for_both_txs(self) -> None:
        withdrawal = [tx for tx in transfer_txs(COSMOS) if "WITHDRAWAL ICA" in tx.title]
        reason = "the WITHDRAWAL ICA holds 400000uatom, below the 1000000uatom test amount: nothing to send"

        self.assertEqual([(tx.ready, tx.reason) for tx in withdrawal], [(False, reason)] * 2)
        self.assertIn("transfer-from-ica cosmoshub-4 WITHDRAWAL 400000uatom ", withdrawal[1].commands[0].text)

    def test_missing_ica_entry_counts_as_an_empty_balance(self) -> None:
        redemption = [tx for tx in transfer_txs(COSMOS) if "REDEMPTION ICA" in tx.title]

        self.assertEqual(
            {tx.reason for tx in redemption},
            {"the REDEMPTION ICA holds 0uatom, below the 1000000uatom test amount: nothing to send"},
        )

    def test_exactly_the_test_amount_is_ready(self) -> None:
        data = {"zones": [{"chain_id": COSMOS, "host_denom": "uatom", "ica_balances": {ica: {"uatom": "1000000"} for ica in ("FEE", "WITHDRAWAL", "DELEGATION", "REDEMPTION")}}]}

        self.assertTrue(all(tx.ready for tx in transfer_txs(COSMOS, funds_data=data)))

    def test_an_in_flight_transfer_blocks_the_rest_but_not_the_test(self) -> None:
        def flight(ica: str, status: str) -> dict[str, object]:
            return {"time": "t", "height": "1", "ica": ica, "sequence": "7", "amount": "1000000", "denom": "uatom", "status": status}

        data = fake_funds()
        data["zones"][1]["transfers"] = [flight("FEE", "in flight"), flight("DELEGATION", "settled")]

        txs = transfer_txs(COSMOS, funds_data=data)
        fee_test, fee_rest = [tx for tx in txs if "FEE ICA" in tx.title]
        delegation = [tx for tx in txs if "DELEGATION ICA" in tx.title]

        self.assertEqual((fee_test.ready, fee_test.reason), (True, None))
        self.assertEqual(
            (fee_rest.ready, fee_rest.reason),
            (False, "a transfer from the FEE ICA is in flight: copy once it settles and the Funds snapshot refreshes"),
        )
        self.assertIn("FEE 2500000uatom ", fee_rest.commands[0].text)  # the stale amount stays visible
        self.assertEqual([(tx.ready, tx.reason) for tx in delegation], [(True, None)] * 2)

    def test_an_unknown_transfer_index_does_not_block_the_rest(self) -> None:
        data = fake_funds()
        data["zones"][1]["transfers"] = None

        fee = [tx for tx in transfer_txs(COSMOS, funds_data=data) if "FEE ICA" in tx.title]

        self.assertEqual([(tx.ready, tx.reason) for tx in fee], [(True, None)] * 2)

    def test_zone_error_missing_zone_and_missing_snapshot_are_not_ready_with_placeholders(self) -> None:
        cases = [
            (transfer_txs(CELESTIA), "URLError: celestia down"),
            (transfer_txs("juno-1"), "zone missing from the Funds snapshot"),
            (
                [tx for tx in sets_by_id(None)["ica-transfers"].txs if tx.chain_id == COSMOS],
                "waiting for the Funds snapshot",
            ),
        ]
        for txs, reason in cases:
            with self.subTest(reason=reason):
                self.assertEqual(len(txs), 8)
                self.assertEqual({(tx.ready, tx.reason) for tx in txs}, {(False, reason)})
                self.assertIn("transfer-from-ica", txs[1].commands[0].text)
                self.assertIn("<BALANCE><HOST_DENOM> ", txs[1].commands[0].text)
                self.assertIn("FEE 1000000<HOST_DENOM> ", txs[0].commands[0].text)

    def test_every_zone_has_its_txs_contiguous_and_in_zone_order(self) -> None:
        all_sets = sets_by_id(fake_snapshot(), funds_data=fake_funds(), pools_data=fake_pools())
        for tx_set in all_sets.values():
            with self.subTest(tx_set=tx_set.id):
                chain_ids = [tx.chain_id for tx in tx_set.txs]
                if tx_set.id == "pool-creation":  # its bundles span every zone, under one heading
                    self.assertEqual(set(chain_ids), {"all zones"})
                    continue
                runs = [chain_id for index, chain_id in enumerate(chain_ids) if index == 0 or chain_ids[index - 1] != chain_id]

                self.assertEqual(runs, [zone.chain_id for zone in config.ZONES if zone.chain_id in chain_ids])


class StaketiaClaimBalanceTest(unittest.TestCase):
    def test_test_and_rest_are_exact(self) -> None:
        tx_set = sets_by_id(None)["staketia-claim-balance"]
        generate = (
            "mkdir -p /tmp/wind-down\n"
            "strided tx stakeibc transfer-staketia-claim-balance {amount} "
            f"--from {ADDRESS} --generate-only --chain-id stride-1 --node {NODE} --gas 600000 --fees 3000ustrd "
            "> /tmp/wind-down/staketia-claim-balance-{label}.unsigned.json"
        )

        self.assertEqual([tx.chain_id for tx in tx_set.txs], ["celestia", "celestia"])
        self.assertEqual(
            command_pairs(tx_set.txs[0]),
            expected_commands(generate.format(amount=1000000, label="test"), stem="/tmp/wind-down/staketia-claim-balance-test"),
        )
        self.assertEqual(
            command_pairs(tx_set.txs[1]),
            expected_commands(generate.format(amount=0, label="rest"), stem="/tmp/wind-down/staketia-claim-balance-rest"),
        )
        self.assertEqual(tx_set.txs[0].title, "celestia · staketia claim balance · test: 1 TIA")
        self.assertEqual(tx_set.txs[1].title, "celestia · staketia claim balance · rest: the whole remainder (amount 0)")
        self.assertEqual([(tx.ready, tx.reason) for tx in tx_set.txs], [(True, None)] * 2)


def pool_set(pools_data: dict[str, object] | None) -> multisig.TxSet:
    return sets_by_id(None, pools_data=pools_data)["pool-funding"]


def pool_txs(chain_id: str = COSMOS, pools_data: dict[str, object] | None = None) -> list[multisig.MultisigTx]:
    return [tx for tx in pool_set(fake_pools() if pools_data is None else pools_data).txs if tx.chain_id == chain_id]


def pool_generate(contract: str, message: str, amount: str | None, label: str) -> str:
    stem = f"/tmp/wind-down/pool-cosmoshub-4-{label}"
    amount_flag = f" --amount {amount}" if amount else ""
    return (
        "mkdir -p /tmp/wind-down\n"
        f"osmosisd tx wasm execute {contract} '{message}'{amount_flag} --from {VAULT} --generate-only "
        f"--chain-id osmosis-1 --node {OSMOSIS_NODE} --gas 1500000 --gas-prices 0.1uosmo > {stem}.unsigned.json"
    )


def pool_commands(generate: str, label: str) -> list[tuple[str, str]]:
    return expected_commands(
        generate, stem=f"/tmp/wind-down/pool-cosmoshub-4-{label}", binary="osmosisd", address=VAULT, chain_id="osmosis-1", node=OSMOSIS_NODE
    )


CANONICAL_STATOM = "ibc/C140AFD542AE77BD7DCC83F13FDD8C5E5BB8C4929785E6EC2F4C636F98F17901"
AXELAR_STATOM = "ibc/A5F0C3C4D7E8F9A1B2C3D4E5F60718293A4B5C6D7E8F9A0B1C2D3E4F5A6B7C8D"
AXELAR_STATOM_ON_AXELAR = "ibc/0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF"
ALLUSDC = "factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC"
CREATION_FEE = {"denom": ALLUSDC, "amount": "20000000"}


def planned_entry(
    kind: str,
    subdenom: str,
    denom_on_osmosis: str | None,
    stride_channel: str | None = None,
    holder_name: str | None = "stride",
    holder_chain_id: str | None = "stride-1",
    seeded: bool | None = True,
    live_contract: str | None = None,
    error: str | None = None,
    blocked: str | None = None,
) -> dict[str, object]:
    """A planned pool as the Pools snapshot serialises it; the message carries the Hub's factors."""
    message = None if error else {
        "pool_asset_configs": [
            {"denom": denom_on_osmosis, "normalization_factor": "1000000000000000000"},
            {"denom": OSMOSIS_DENOM, "normalization_factor": "2013525450106978250"},
        ],
        "alloyed_asset_subdenom": subdenom,
        "alloyed_asset_normalization_factor": "2013525450106978250",
        "admin": VAULT,
        "moderator": MODERATOR,
    }
    return {
        "kind": kind,
        "stride_channel": stride_channel,
        "holder_chain_id": holder_chain_id,
        "holder_name": holder_name,
        "holder_binary": "strided",
        "holder_node": NODE,
        "counterparty_channel": None,
        "osmosis_channel": "channel-326",
        "holder_to_osmosis_channel": "channel-5",
        "denom_on_holder": None if kind == "canonical" else AXELAR_STATOM_ON_AXELAR,
        "denom_on_osmosis": denom_on_osmosis,
        "seeded": seeded,
        "test_wallet_balance": None if seeded is None else ("5" if seeded else "0"),
        "st_factor": "1000000000000000000",
        "native_factor": "2013525450106978250",
        "alloyed_subdenom": subdenom,
        "instantiate_msg": message,
        "seed_command": None,
        "live_contract": live_contract,
        "error": error,
        "blocked": blocked,
    }


def fake_planned() -> list[dict[str, object]]:
    return [
        planned_entry("canonical", "stATOM", CANONICAL_STATOM),
        planned_entry(
            "route", "stATOM.axelar.channel11", AXELAR_STATOM, stride_channel="channel-11", holder_name="axelar", holder_chain_id="axelar-dojo-1"
        ),
    ]


def creation_txs(pools_data: dict[str, object] | None = None) -> list[multisig.MultisigTx]:
    return sets_by_id(None, pools_data=fake_pools() if pools_data is None else pools_data)["pool-creation"].txs


def creation_line(message_json: str, gas: int, part: str) -> str:
    return (
        f"osmosisd tx cosmwasmpool create-pool 996 '{message_json}' --from {VAULT} --generate-only "
        f"--chain-id osmosis-1 --node {OSMOSIS_NODE} --gas {gas} --gas-prices 0.1uosmo > {part}.unsigned.json"
    )


def merge_line(stem: str) -> str:
    return (
        "jq -s '(map(.body.messages) | add) as $msgs | .[0] | .body.messages = $msgs | .signatures = []' "
        f"{stem}-*.unsigned.json > {stem}.unsigned.json"
    )


def creation_commands(generate: str, stem: str) -> list[tuple[str, str]]:
    return expected_commands(generate, stem=stem, binary="osmosisd", address=VAULT, chain_id="osmosis-1", node=OSMOSIS_NODE)


def hub_message(denom: str, subdenom: str) -> str:
    return (
        '{"pool_asset_configs":[{"denom":"' + denom + '","normalization_factor":"1000000000000000000"},'
        '{"denom":"' + OSMOSIS_DENOM + '","normalization_factor":"2013525450106978250"}],'
        '"alloyed_asset_subdenom":"' + subdenom + '","alloyed_asset_normalization_factor":"2013525450106978250",'
        '"admin":"' + VAULT + '","moderator":"' + MODERATOR + '"}'
    )


def many_planned(count: int) -> list[dict[str, object]]:
    return [planned_entry("route", f"stATOM.r{index}", AXELAR_STATOM, stride_channel=f"channel-{index}", holder_name="axelar") for index in range(count)]


WAITING_FOR_OTHER_ZONES = (
    "celestia pool creation: zone missing from the Pools snapshot; "
    "dydx-mainnet-1 pool creation: URLError: osmosis down; "
    "haqq_11235-1 pool creation: zone missing from the Pools snapshot; "
    "injective-1 pool creation: zone missing from the Pools snapshot; "
    "juno-1 pool creation: no planned pool in the Pools snapshot; "
    "laozi-mainnet pool creation: zone missing from the Pools snapshot; "
    "osmosis-1 pool creation: zone missing from the Pools snapshot; "
    "phoenix-1 pool creation: zone missing from the Pools snapshot; … and 2 more"
)


class PoolCreationTest(unittest.TestCase):
    def test_the_set_comes_first(self) -> None:
        tx_sets = multisig.tx_sets(validators_data=None, funds_data=None, pools_data=fake_pools())

        self.assertEqual((tx_sets[0].id, tx_sets[0].step_id), ("pool-creation", "vote-pools-create"))

    def test_the_ready_pools_form_one_bundle_with_exact_commands(self) -> None:
        bundle, waiting = creation_txs()
        stem = "/tmp/wind-down/create-b1"
        generate = "\n".join(
            [
                "mkdir -p /tmp/wind-down",
                creation_line(hub_message(CANONICAL_STATOM, "stATOM"), 3_100_000, f"{stem}-01-statom"),
                creation_line(hub_message(AXELAR_STATOM, "stATOM.axelar.channel11"), 3_100_000, f"{stem}-02-statom-axelar-channel11"),
                merge_line(stem),
            ]
        )

        self.assertEqual(bundle.chain_id, "all zones")
        self.assertEqual(bundle.title, "bundle 1 of 1 · 2 pools")
        self.assertEqual(bundle.members, ["cosmoshub-4 stATOM", "cosmoshub-4 stATOM.axelar.channel11"])
        self.assertEqual(command_pairs(bundle), creation_commands(generate, stem=stem))
        self.assertEqual((bundle.ready, bundle.reason), (True, None))
        self.assertEqual(bundle.files, [f"{stem}.unsigned.json", f"{stem}.FS5.json", f"{stem}.FA5.json", f"{stem}.FR5.json"])
        self.assertEqual(bundle.commands[0].label, "Write each pool's unsigned tx and merge them into one (2 messages, --gas 3100000)")
        self.assertEqual((waiting.title, waiting.ready, waiting.commands), ("not in a bundle: 10 pools without a message yet", False, []))
        self.assertEqual(waiting.reason, WAITING_FOR_OTHER_ZONES)

    def test_bundles_split_at_eighteen_messages_and_number_their_parts_per_bundle(self) -> None:
        with mock.patch.object(config, "CREATE_POOL_BUNDLE_ZONES", ()):
            txs = creation_txs(pools_data=fake_pools(planned=many_planned(20), vault_fee_balance="400000000"))
        first, second = txs[0], txs[1]

        self.assertEqual((first.title, first.members[0], first.members[-1]), ("bundle 1 of 2 · 18 pools", "cosmoshub-4 stATOM.r0", "cosmoshub-4 stATOM.r17"))
        self.assertEqual((second.title, second.members), ("bundle 2 of 2 · 2 pools", ["cosmoshub-4 stATOM.r18", "cosmoshub-4 stATOM.r19"]))
        self.assertIn("--gas 27100000 --gas-prices 0.1uosmo > /tmp/wind-down/create-b1-18-statom-r17.unsigned.json", first.commands[0].text)
        self.assertIn("--gas 3100000 --gas-prices 0.1uosmo > /tmp/wind-down/create-b2-01-statom-r18.unsigned.json", second.commands[0].text)
        self.assertIn("create-b2-*.unsigned.json > /tmp/wind-down/create-b2.unsigned.json", second.commands[0].text)
        self.assertEqual(len(txs), 3)  # plus the waiting tx for the other zones

    def test_unseeded_pools_sit_in_their_bundle_not_ready_while_unresolved_ones_wait_outside(self) -> None:
        planned = [
            planned_entry("canonical", "stATOM", CANONICAL_STATOM),
            planned_entry("route", "stATOM.channel47", None, stride_channel="channel-47", holder_name=None, holder_chain_id="carbon-1", seeded=None, error="carbon-1 is not in config.HOLDER_CHAINS"),
            planned_entry("route", "stATOM.axelar", AXELAR_STATOM, stride_channel="channel-11", holder_name="axelar", seeded=False),
            planned_entry("route", "stATOM.secret", AXELAR_STATOM, stride_channel="channel-40", holder_name="secret", seeded=None),
            planned_entry("route", "stATOM.cosmoshub", AXELAR_STATOM, stride_channel="channel-0", holder_name="cosmoshub", live_contract="osmo1hub"),
        ]

        done, bundle, waiting = creation_txs(pools_data=fake_pools(planned=planned, vault_fee_balance="80000000"))

        # The created pool is bundle 1, done; the unseeded pools are in bundle 2, with the bundle not ready.
        self.assertEqual((done.title, done.done, done.members), ("bundle 1 of 2 · 1 pools · done", True, ["cosmoshub-4 stATOM.cosmoshub"]))
        self.assertEqual((bundle.title, bundle.members), ("bundle 2 of 2 · 3 pools", ["cosmoshub-4 stATOM", "cosmoshub-4 stATOM.axelar", "cosmoshub-4 stATOM.secret"]))
        self.assertEqual((bundle.ready, bundle.reason), (False, "2 of 3 pools waiting on the test wallet: cosmoshub-4 stATOM.axelar, cosmoshub-4 stATOM.secret"))
        self.assertEqual(len(bundle.commands), 5)
        self.assertIn("create-b2-03-statom-secret.unsigned.json", bundle.commands[0].text)
        # A route with no instantiate message cannot be bundled; a created pool is simply done.
        self.assertTrue(waiting.reason.startswith("celestia pool creation: zone missing from the Pools snapshot; cosmoshub-4 stATOM.channel47: carbon-1 is not in config.HOLDER_CHAINS; "))
        self.assertNotIn("stATOM.cosmoshub", waiting.reason)
        self.assertEqual(waiting.title, "not in a bundle: 11 pools without a message yet")

    def test_created_pools_are_a_done_bundle_and_the_rest_start_at_the_next_number(self) -> None:
        created = [planned_entry("canonical", "stATOM", CANONICAL_STATOM, live_contract="osmo1canon"), planned_entry("route", "stATOM.secret", AXELAR_STATOM, stride_channel="channel-40", holder_name="secret", live_contract="osmo1secret")]
        (done, _) = creation_txs(pools_data=fake_pools(planned=created))
        mixed = created[:1] + [planned_entry("route", "stATOM.neutron", AXELAR_STATOM, stride_channel="channel-123", holder_name="neutron")]
        (first, second, _) = creation_txs(pools_data=fake_pools(planned=mixed))

        self.assertEqual((done.title, done.done, done.ready, done.reason, done.commands), ("bundle 1 of 1 · 2 pools · done", True, False, None, []))
        self.assertEqual(done.members, ["cosmoshub-4 stATOM", "cosmoshub-4 stATOM.secret"])
        self.assertEqual((first.title, first.done, first.members), ("bundle 1 of 2 · 1 pools · done", True, ["cosmoshub-4 stATOM"]))
        self.assertEqual((second.title, second.done, second.ready, second.members), ("bundle 2 of 2 · 1 pools", False, True, ["cosmoshub-4 stATOM.neutron"]))
        self.assertIn("create-b2-01-statom-neutron.unsigned.json", second.commands[0].text)
        self.assertIn("(1 messages, --gas 1600000)", second.commands[0].label)

    def test_created_pools_keep_bundle_one_and_the_rest_follow_the_zone_groups(self) -> None:
        def zone(chain_id: str, *plans: dict[str, object]) -> dict[str, object]:
            return {"chain_id": chain_id, "pools": [], "planned": list(plans), "creation_fee": CREATION_FEE, "vault_fee_balance": "400000000", "creation_fee_short": False}

        data = {"zones": [
            zone(COSMOS, planned_entry("canonical", "stATOM", CANONICAL_STATOM, live_contract="osmo1a"), planned_entry("route", "stATOM.secret", AXELAR_STATOM, stride_channel="channel-40", holder_name="secret")),
            zone("osmosis-1", planned_entry("canonical", "stOSMO", CANONICAL_STATOM)),
            zone("juno-1", planned_entry("canonical", "stJUNO", CANONICAL_STATOM, live_contract="osmo1b")),
            zone("ssc-1", planned_entry("canonical", "stSAGA", CANONICAL_STATOM)),
        ]}
        with mock.patch.object(config, "CREATE_POOL_BUNDLE_ZONES", (("osmosis-1",), ("ssc-1", COSMOS))):
            txs = creation_txs(pools_data=data)

        self.assertEqual([(tx.title, tx.members) for tx in txs[:3]], [
            ("bundle 1 of 3 · 2 pools · done", ["cosmoshub-4 stATOM", "juno-1 stJUNO"]),
            ("bundle 2 of 3 · 1 pools", ["osmosis-1 stOSMO"]),
            ("bundle 3 of 3 · 2 pools", ["ssc-1 stSAGA", "cosmoshub-4 stATOM.secret"]),
        ])

    def test_blocked_pools_form_the_last_bundle_and_are_never_ready(self) -> None:
        planned = [
            planned_entry("canonical", "stATOM", CANONICAL_STATOM),
            planned_entry("route", "stATOM.injective", AXELAR_STATOM, stride_channel="channel-6", holder_name="injective", blocked="rate limiter"),
            planned_entry("route", "stATOM.secret", AXELAR_STATOM, stride_channel="channel-40", holder_name="secret"),
        ]

        open_bundle, blocked_bundle, _ = creation_txs(pools_data=fake_pools(planned=planned, vault_fee_balance="80000000"))

        self.assertEqual((open_bundle.title, open_bundle.members), ("bundle 1 of 2 · 2 pools", ["cosmoshub-4 stATOM", "cosmoshub-4 stATOM.secret"]))
        self.assertEqual((open_bundle.ready, open_bundle.reason), (True, None))
        self.assertEqual((blocked_bundle.title, blocked_bundle.members), ("bundle 2 of 2 · 1 pools (blocked)", ["cosmoshub-4 stATOM.injective"]))
        self.assertEqual((blocked_bundle.ready, blocked_bundle.blocked, blocked_bundle.reason), (False, True, "blocked: rate limiter"))
        self.assertFalse(open_bundle.blocked)
        self.assertIn("create-b2-01-statom-injective.unsigned.json", blocked_bundle.commands[0].text)

    def test_a_fee_shortfall_or_unknown_fee_blocks_every_bundle_but_keeps_its_commands(self) -> None:
        short = creation_txs(pools_data=fake_pools(creation_fee_short=True, vault_fee_balance="0"))[0]
        unknown = creation_txs(pools_data=fake_pools(creation_fee_short=None, creation_fee=None, vault_fee_balance=None))[0]

        self.assertEqual((short.ready, short.reason), (False, "vault holds 0 allUSDC, needs 20 × 2 = 40 allUSDC: top it up first"))
        self.assertEqual((unknown.ready, unknown.reason), (False, "the pool creation fee or the vault's balance of it is not known yet (see the Pools tab)"))
        self.assertTrue(all(len(tx.commands) == 5 for tx in (short, unknown)))

    def test_the_fee_gate_counts_pools_across_zones_against_the_one_vault_balance(self) -> None:
        def two_zone_pools(balance: str) -> dict[str, object]:
            def zone(chain_id: str, subdenom: str) -> dict[str, object]:
                return {
                    "chain_id": chain_id,
                    "pools": [],
                    "planned": [planned_entry("canonical", subdenom, CANONICAL_STATOM)],
                    "creation_fee": CREATION_FEE,
                    "vault_fee_balance": balance,
                    "creation_fee_short": False,
                }

            return {"zones": [zone(COSMOS, "stATOM"), zone("juno-1", "stJUNO")]}

        with mock.patch.object(config, "CREATE_POOL_BUNDLE_ZONES", ()):
            short = creation_txs(two_zone_pools("20000000"))[0]
            enough = creation_txs(two_zone_pools("40000000"))[0]

        self.assertEqual((short.title, short.members), ("bundle 1 of 1 · 2 pools", ["cosmoshub-4 stATOM", "juno-1 stJUNO"]))
        self.assertEqual((short.ready, short.reason), (False, "vault holds 20 allUSDC, needs 20 × 2 = 40 allUSDC: top it up first"))
        self.assertEqual((enough.ready, enough.reason), (True, None))

    def test_an_unseeded_pool_still_gets_its_bundle_and_a_missing_snapshot_only_the_waiting_tx(self) -> None:
        bundle, waiting = creation_txs(pools_data=fake_pools(planned=[planned_entry("canonical", "stATOM", CANONICAL_STATOM, seeded=False)]))
        (missing,) = sets_by_id(None)["pool-creation"].txs

        self.assertEqual((bundle.title, bundle.ready, bundle.reason), ("bundle 1 of 1 · 1 pools", False, "1 of 1 pools waiting on the test wallet: cosmoshub-4 stATOM"))
        self.assertEqual(waiting.title, "not in a bundle: 10 pools without a message yet")
        self.assertEqual((missing.title, missing.ready, missing.commands, missing.files), ("not in a bundle: 11 pools without a message yet", False, [], []))
        self.assertTrue(missing.reason.startswith("celestia pool creation: waiting for the Pools snapshot; "))

    def test_description_states_the_bundle_rules_the_fee_and_the_factor_rule(self) -> None:
        with_fee = sets_by_id(None, pools_data=fake_pools())["pool-creation"]
        without = sets_by_id(None, pools_data=fake_pools(creation_fee=None))["pool-creation"]
        notes = "\n".join(with_fee.notes)

        self.assertEqual(with_fee.description, "Create every planned pool on Osmosis from the vault, 18 pools per multisig tx.")
        self.assertIn("Creation fee: 20 allUSDC per pool, paid by the vault; 2 pools still to create.", with_fee.notes)
        self.assertIn("Creation fee: not known yet per pool, paid by the vault; 2 pools still to create.", without.notes)
        self.assertIn("Gas 1,500,000 per message at 0.1uosmo/gas, about 3x the base fee", notes)
        self.assertIn("1e6 for haqq, dYdX and Injective", notes)
        self.assertIn("generate and sign a bundle the same day", notes)
        self.assertEqual(with_fee.notes[-1], multisig.ONE_AT_A_TIME)


class PoolFundingTest(unittest.TestCase):
    def test_canonical_test_join_rest_join_and_mark_are_exact(self) -> None:
        txs = pool_txs()
        by_title = {tx.title: tx for tx in txs}
        join = '{"join_pool":{}}'
        mark = '{"mark_corrupted_assets":{"denoms":["' + OSMOSIS_DENOM + '"]}}'

        test_join = by_title[f"cosmoshub-4 · canonical pool 1234 · test join, 1000000{OSMOSIS_DENOM}"]
        rest_join = by_title[f"cosmoshub-4 · canonical pool 1234 · join rest, 4000000{OSMOSIS_DENOM}"]
        mark_tx = by_title["cosmoshub-4 · canonical pool 1234 · mark corrupted"]

        self.assertEqual(
            command_pairs(test_join),
            pool_commands(pool_generate(CANONICAL_CONTRACT, join, f"1000000{OSMOSIS_DENOM}", "1234-test"), label="1234-test"),
        )
        self.assertEqual(
            command_pairs(rest_join),
            pool_commands(pool_generate(CANONICAL_CONTRACT, join, f"4000000{OSMOSIS_DENOM}", "1234-rest"), label="1234-rest"),
        )
        self.assertEqual(
            command_pairs(mark_tx), pool_commands(pool_generate(CANONICAL_CONTRACT, mark, None, "1234-mark"), label="1234-mark")
        )
        self.assertEqual([(tx.ready, tx.reason) for tx in (test_join, rest_join, mark_tx)], [(True, None)] * 3)
        self.assertEqual(test_join.files[0], "/tmp/wind-down/pool-cosmoshub-4-1234-test.unsigned.json")

    def test_order_is_every_test_join_then_each_pools_rest_and_mark_canonical_first(self) -> None:
        steps = [tx.title.removeprefix("cosmoshub-4 · ").split(" · ")[:2] for tx in pool_txs()]

        self.assertEqual(
            [f"{pool} {step.split(',')[0]}" for pool, step in steps],
            [
                "canonical pool 1234 test join",
                "route pool abc123 test join",
                "canonical pool 1234 join rest",
                "canonical pool 1234 mark corrupted",
                "route pool abc123 join rest",
                "route pool abc123 mark corrupted",
            ],
        )

    def test_a_pool_without_an_id_is_named_by_the_last_six_characters_of_its_contract(self) -> None:
        route_test = next(tx for tx in pool_txs() if "route pool abc123 · test join" in tx.title)

        self.assertIn("/tmp/wind-down/pool-cosmoshub-4-abc123-test.unsigned.json", route_test.commands[0].text)
        self.assertIn(f"wasm execute {ROUTE_CONTRACT} ", route_test.commands[0].text)

    def test_unrecognised_pools_are_not_funded(self) -> None:
        self.assertFalse(any("unrecognised" in tx.title or "osmo1unrecognisedcontract" in tx.commands[0].text for tx in pool_txs()))

    def test_neither_join_is_ready_until_the_allocation_is_known(self) -> None:
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", None, None)]
        reason = "the pool's allocation is not known yet (see the Pools tab)"

        txs = pool_txs(pools_data=fake_pools(pools=pools))
        test_join, rest, mark = txs

        self.assertEqual([(tx.ready, tx.reason) for tx in (test_join, rest)], [(False, reason)] * 2)
        self.assertIn(f"--amount 1000000{OSMOSIS_DENOM} ", test_join.commands[0].text)
        self.assertIn("--amount <ALLOCATION_MINUS_VAULT_SHARES>" + OSMOSIS_DENOM, rest.commands[0].text)
        self.assertTrue(mark.ready)

    def test_rest_join_is_not_ready_when_the_pool_is_funded_exactly(self) -> None:
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "5000000", True, vault_shares="5000000")]

        txs = pool_txs(pools_data=fake_pools(pools=pools))
        rest = next(tx for tx in txs if "join rest" in tx.title)

        self.assertEqual((rest.ready, rest.reason), (False, "the pool is already funded exactly"))
        self.assertIn(f"--amount 0{OSMOSIS_DENOM} ", rest.commands[0].text)
        self.assertTrue(all(tx.ready for tx in txs if "join rest" not in tx.title))

    def test_an_allocation_below_the_test_amount_is_joined_whole_by_the_test_and_leaves_no_rest(self) -> None:
        # Live: the stuatom channel-11 route holds ~20,136 uatom of escrow, far below the one-token test join.
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "20136", False)]

        test_join, rest, _ = pool_txs(pools_data=fake_pools(pools=pools))

        self.assertEqual((test_join.ready, test_join.reason), (True, None))
        self.assertEqual(test_join.title, f"cosmoshub-4 · canonical pool 1234 · test join, 20136{OSMOSIS_DENOM}")
        self.assertIn(f"--amount 20136{OSMOSIS_DENOM} ", test_join.commands[0].text)
        self.assertEqual((rest.ready, rest.reason), (False, "the allocation (20136) does not exceed the test join: nothing left to join"))

    def test_a_zero_allocation_readies_neither_join(self) -> None:
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "0", False)]

        test_join, rest, _ = pool_txs(pools_data=fake_pools(pools=pools))

        self.assertEqual((test_join.ready, test_join.reason), (False, "the allocation is 0: nothing to join"))
        self.assertEqual((rest.ready, rest.reason), (False, "the allocation is 0: nothing to join"))
        self.assertIn(f"--amount 0{OSMOSIS_DENOM} ", test_join.commands[0].text)

    def test_an_allocation_equal_to_the_test_amount_leaves_no_rest(self) -> None:
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "1000000", False)]

        test_join, rest, _ = pool_txs(pools_data=fake_pools(pools=pools))

        self.assertIn(f"--amount 1000000{OSMOSIS_DENOM} ", test_join.commands[0].text)
        self.assertTrue(test_join.ready)
        self.assertEqual(rest.reason, "the allocation (1000000) does not exceed the test join: nothing left to join")
        self.assertFalse(rest.ready)

    def test_rest_join_waits_until_the_snapshot_shows_exactly_the_test_join(self) -> None:
        cases = [("0", "5000000"), ("999999", "4000001"), ("2000000", "3000000")]
        for vault_shares, remaining in cases:
            with self.subTest(vault_shares=vault_shares):
                pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "5000000", False, vault_shares=vault_shares)]

                test_join, rest, _ = pool_txs(pools_data=fake_pools(pools=pools))

                self.assertTrue(test_join.ready)
                self.assertEqual(
                    (rest.ready, rest.reason),
                    (False, f"the Pools snapshot does not show the test join yet (vault shares {vault_shares}, expected 1000000)"),
                )
                self.assertIn(f"--amount {remaining}{OSMOSIS_DENOM} ", rest.commands[0].text)
                self.assertEqual(rest.title, f"cosmoshub-4 · canonical pool 1234 · join rest, {remaining}{OSMOSIS_DENOM}")

    def test_allocation_one_above_the_test_is_ready_for_a_single_unit(self) -> None:
        pools = [pool_entry(CANONICAL_CONTRACT, "canonical", "1234", "1000001", False, vault_shares="1000000")]

        rest = next(tx for tx in pool_txs(pools_data=fake_pools(pools=pools)) if "join rest" in tx.title)

        self.assertTrue(rest.ready)
        self.assertIn(f"--amount 1{OSMOSIS_DENOM} ", rest.commands[0].text)

    def test_unknown_osmosis_denom_blocks_every_tx_of_the_pool_but_keeps_the_commands(self) -> None:
        txs = pool_txs(pools_data=fake_pools(osmosis_denom=None))

        self.assertEqual(len(txs), 6)
        self.assertEqual({tx.ready for tx in txs}, {False})
        self.assertTrue(all("native denom on Osmosis" in tx.reason for tx in txs))
        self.assertIn("--amount 1000000<OSMOSIS_DENOM>", txs[0].commands[0].text)
        self.assertIn('"denoms":["<OSMOSIS_DENOM>"]', next(tx for tx in txs if "mark" in tx.title).commands[0].text)

    def test_zone_error_missing_zone_empty_zone_and_missing_snapshot_get_one_not_ready_tx(self) -> None:
        cases = [
            (pool_txs(DYDX), "URLError: osmosis down"),
            (pool_txs("juno-1"), "no pool to fund in the Pools snapshot"),
            (pool_txs("ssc-1"), "zone missing from the Pools snapshot"),
            ([tx for tx in sets_by_id(None)["pool-funding"].txs if tx.chain_id == COSMOS], "waiting for the Pools snapshot"),
        ]
        for txs, reason in cases:
            with self.subTest(reason=reason):
                self.assertEqual([(tx.ready, tx.reason, tx.commands, tx.files) for tx in txs], [(False, reason, [], [])])

    def test_description_states_the_order_and_the_osmo_fee_note(self) -> None:
        tx_set = pool_set(fake_pools())
        notes = "\n".join(tx_set.notes)

        self.assertEqual(tx_set.description, "Fund each pool with a test join, then the rest of its allocation, then mark the native token corrupted.")
        self.assertIn("every pool's test join first", notes)
        self.assertIn("the rest join and the mark back to back", notes)
        self.assertIn("about 0.15 OSMO per tx, three txs per pool", notes)
        self.assertIn("once the Pools snapshot shows the test join as vault shares", notes)

    def test_transfers_description_states_the_in_flight_gate(self) -> None:
        notes = "\n".join(sets_by_id(None)["ica-transfers"].notes)

        self.assertIn("stays not ready while a transfer from that ICA is in flight", notes)


class MultisigBodyTest(unittest.TestCase):
    def test_absent_pools_cache_is_no_data_and_fetched_at_is_the_oldest_that_exists(self) -> None:
        views = {
            "validators": {"data": fake_snapshot(), "fetched_at": "2026-10-07T12:05:00+00:00"},
            "funds": {"data": fake_funds(), "fetched_at": "2026-10-07T12:01:00+00:00"},
            "pools": {"loading": True},
        }

        body = server.multisig_body(views=views)

        self.assertEqual(body["fetched_at"], "2026-10-07T12:01:00+00:00")
        sets = {tx_set["id"]: tx_set for tx_set in body["data"]["sets"]}
        self.assertEqual(
            set(sets),
            {"pool-creation", "live-test-undelegate", "full-drain", "ica-transfers", "staketia-claim-balance", "pool-funding"},
        )
        self.assertEqual(sets["pool-funding"]["txs"][0]["reason"], "waiting for the Pools snapshot")
        self.assertTrue(sets["pool-creation"]["txs"][0]["reason"].startswith("celestia pool creation: waiting for the Pools snapshot; "))
        self.assertTrue(sets["ica-transfers"]["txs"][8]["ready"])

    def test_nothing_loaded_yet_has_no_fetched_at(self) -> None:
        body = server.multisig_body(views={name: {"loading": True} for name in server.MULTISIG_SOURCES})

        self.assertIsNone(body["fetched_at"])
        json.dumps(body)

    def test_an_unregistered_cache_is_treated_as_still_loading(self) -> None:
        registered = {name: cache for name, cache in server.CACHES.items() if name != "pools"}

        with mock.patch.dict(server.CACHES, registered, clear=True):
            views = server.multisig_views()

        self.assertEqual(views["pools"], {"loading": True})
        self.assertEqual(set(views), {"validators", "funds", "pools"})


if __name__ == "__main__":
    unittest.main()
