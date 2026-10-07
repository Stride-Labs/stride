import json
import unittest

import config
import multisig

ADDRESS = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
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


def sets_by_id(validators_data: dict[str, object] | None) -> dict[str, multisig.TxSet]:
    return {tx_set.id: tx_set for tx_set in multisig.tx_sets(validators_data=validators_data)}


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
                    "Aidan",
                    f"strided tx multisign {stem}.unsigned.json F5 {stem}.FS5.json {stem}.FA5.json "
                    f"--chain-id stride-1 --node {NODE} > {stem}.signed.json\n"
                    f"strided tx broadcast {stem}.signed.json --node {NODE} --broadcast-mode sync",
                ),
            ],
        )
        self.assertTrue(tx.ready)
        self.assertIsNone(tx.reason)
        self.assertEqual(tx.title, "celestia · live test: mhventures (celestiavaloper1q2k…), 15,861,063 utia")
        self.assertEqual(tx.files, [f"{stem}.unsigned.json", f"{stem}.FS5.json", f"{stem}.FA5.json", f"{stem}.FR5.json"])

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
    def test_no_snapshot_leaves_every_tx_not_ready(self) -> None:
        for tx_set in multisig.tx_sets(validators_data=None):
            self.assertEqual([tx.chain_id for tx in tx_set.txs], [zone.chain_id for zone in config.ZONES])
            for tx in tx_set.txs:
                self.assertEqual((tx.ready, tx.reason), (False, "waiting for the Validators snapshot"))
                self.assertEqual(len(tx.commands), 5)

    def test_ids_are_unique_and_tied_to_their_ops_steps(self) -> None:
        tx_sets = multisig.tx_sets(validators_data=None)

        self.assertEqual([(tx_set.id, tx_set.step_id) for tx_set in tx_sets], [("live-test-undelegate", "drain-live-test"), ("full-drain", "drain-rest")])

    def test_payload_is_json_serialisable_with_the_documented_shape(self) -> None:
        payload = multisig.tx_sets(validators_data=fake_snapshot())[0].payload()

        json.dumps(payload)
        self.assertEqual(set(payload), {"id", "step_id", "title", "description", "txs"})
        self.assertEqual(set(payload["txs"][0]), {"chain_id", "title", "ready", "reason", "commands", "files"})
        self.assertEqual(set(payload["txs"][0]["commands"][0]), {"tag", "label", "text"})

    def test_multisig_constants(self) -> None:
        self.assertEqual((multisig.MULTISIG_KEY, multisig.MULTISIG_ADDRESS), ("F5", config.PROTOCOL_ADMIN))
        self.assertEqual([signer.key for signer in multisig.DEFAULT_SIGNERS], ["FS5", "FA5"])
        self.assertEqual(multisig.BROADCASTER.tag, "Aidan")


if __name__ == "__main__":
    unittest.main()
