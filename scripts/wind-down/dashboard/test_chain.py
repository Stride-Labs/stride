import datetime
import http.client
import json
import unittest
import urllib.error
import urllib.parse
from unittest import mock

import chain

CHAIN = chain.Chain(chain_id="test-1", rest="https://rest.test", rpc="https://rpc.test")
NOW = datetime.datetime(2026, 9, 30, 12, 0, 0, tzinfo=datetime.UTC)


class IbcDenomTest(unittest.TestCase):
    def test_known_hash(self) -> None:
        self.assertEqual(
            chain.ibc_denom(path="transfer/channel-0/uatom"),
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
        )

    def test_path_is_part_of_the_hash(self) -> None:
        self.assertNotEqual(
            chain.ibc_denom(path="transfer/channel-1/uatom"), chain.ibc_denom(path="transfer/channel-0/uatom")
        )


class ClientExpiryTest(unittest.TestCase):
    def test_expiry_is_consensus_timestamp_plus_trusting_period(self) -> None:
        expiry = chain.client_expiry(consensus_timestamp="2026-09-30T16:39:15.164084974Z", trusting_period_seconds=1036800)

        self.assertEqual(expiry, datetime.datetime(2026, 10, 12, 16, 39, 15, 164084, tzinfo=datetime.UTC))

    def test_timestamp_without_fraction(self) -> None:
        self.assertEqual(
            chain.parse_timestamp(timestamp="2026-09-30T00:00:01Z"),
            datetime.datetime(2026, 9, 30, 0, 0, 1, tzinfo=datetime.UTC),
        )

    def test_duration_parse(self) -> None:
        self.assertEqual(chain.parse_duration_seconds(duration="1036800s"), 1036800.0)

    def test_client_health_reports_seconds_remaining(self) -> None:
        responses = {
            "/ibc/core/client/v1/client_status/07-tendermint-9": {"status": "Active"},
            "/ibc/core/client/v1/client_states/07-tendermint-9": {
                "client_state": {
                    "trusting_period": "86400s",
                    "latest_height": {"revision_number": "1", "revision_height": "500"},
                }
            },
            "/ibc/core/client/v1/consensus_states/07-tendermint-9/revision/1/height/500": {
                "consensus_state": {"timestamp": "2026-09-30T06:00:00.5Z"}
            },
        }

        with mock.patch.object(chain, "rest_get", side_effect=lambda chain, path, params=None: responses[path]):
            health = chain.ibc_client_health(chain=CHAIN, client_id="07-tendermint-9", now=NOW)

        self.assertEqual(health.status, "Active")
        self.assertEqual(health.chain_id, "test-1")
        self.assertEqual(health.expires_at, "2026-10-01T06:00:00.500000+00:00")
        self.assertEqual(health.seconds_remaining, 18 * 3600 + 0.5)

    def test_expired_client_has_negative_remaining(self) -> None:
        responses = {
            "/ibc/core/client/v1/client_status/c": {"status": "Expired"},
            "/ibc/core/client/v1/client_states/c": {
                "client_state": {
                    "trusting_period": "3600s",
                    "latest_height": {"revision_number": "0", "revision_height": "7"},
                }
            },
            "/ibc/core/client/v1/consensus_states/c/revision/0/height/7": {
                "consensus_state": {"timestamp": "2026-09-30T10:00:00Z"}
            },
        }

        with mock.patch.object(chain, "rest_get", side_effect=lambda chain, path, params=None: responses[path]):
            health = chain.ibc_client_health(chain=CHAIN, client_id="c", now=NOW)

        self.assertEqual(health.status, "Expired")
        self.assertEqual(health.seconds_remaining, -3600)


class PaginationTest(unittest.TestCase):
    def test_follows_next_key(self) -> None:
        pages = {
            None: {"items": [1, 2], "pagination": {"next_key": "a+/="}},
            "a+/=": {"items": [3], "pagination": {"next_key": None}},
        }
        calls: list[dict[str, str]] = []

        def fake_rest_get(chain: chain.Chain, path: str, params: dict[str, str]) -> dict:
            calls.append(params)
            return pages[params.get("pagination.key")]

        with mock.patch.object(chain, "rest_get", side_effect=fake_rest_get):
            items = chain.rest_get_all_pages(chain=CHAIN, path="/x", key="items")

        self.assertEqual(items, [1, 2, 3])
        self.assertEqual(calls[1]["pagination.key"], "a+/=")

    def test_stops_at_max_items(self) -> None:
        page = {"items": [1, 2, 3], "pagination": {"next_key": "more"}}

        with mock.patch.object(chain, "rest_get", return_value=page) as rest_get:
            items = chain.rest_get_all_pages(chain=CHAIN, path="/x", key="items", max_items=3)

        self.assertEqual(items, [1, 2, 3])
        self.assertEqual(rest_get.call_count, 1)


class PacketHelpersTest(unittest.TestCase):
    def test_commitments_are_sorted_and_not_capped_at_exactly_the_cap(self) -> None:
        commitments = [{"sequence": str(sequence)} for sequence in (5, 3, 4)]

        with mock.patch.object(chain, "rest_get_all_pages", return_value=commitments):
            result = chain.ibc_packet_commitments(chain=CHAIN, channel_id="channel-1", port_id="transfer", cap=3)

        self.assertEqual(result.sequences, [3, 4, 5])
        self.assertFalse(result.capped)

    def test_commitments_beyond_the_cap_are_truncated_and_flagged(self) -> None:
        commitments = [{"sequence": str(sequence)} for sequence in range(10, 0, -1)]

        with mock.patch.object(chain, "rest_get_all_pages", return_value=commitments) as pages:
            result = chain.ibc_packet_commitments(chain=CHAIN, channel_id="channel-1", port_id="transfer", cap=4)

        self.assertEqual(result.sequences, [1, 2, 3, 4])
        self.assertTrue(result.capped)
        self.assertEqual(pages.call_args.kwargs["max_items"], 5)

    def test_unreceived_batches_the_sequence_list(self) -> None:
        urls: list[str] = []

        def fake_rest_get(chain: chain.Chain, path: str) -> dict:
            urls.append(path)
            batch = path.split("/packet_commitments/")[1].split("/")[0].split(",")
            return {"sequences": [sequence for sequence in batch if int(sequence) % 2 == 0]}

        with mock.patch.object(chain, "rest_get", side_effect=fake_rest_get):
            unreceived = chain.ibc_unreceived_packets(
                chain=CHAIN, channel_id="channel-9", port_id="icahost", sequences=list(range(1, 251))
            )

        self.assertEqual(len(urls), 3)
        self.assertEqual(unreceived, list(range(2, 251, 2)))
        self.assertIn("/channels/channel-9/ports/icahost/packet_commitments/", urls[0])

    def test_unreceived_acks_batches_the_sequence_list(self) -> None:
        urls: list[str] = []

        def fake_rest_get(chain: chain.Chain, path: str) -> dict:
            urls.append(path)
            batch = path.split("/packet_commitments/")[1].split("/")[0].split(",")
            return {"sequences": [sequence for sequence in batch if int(sequence) % 2 == 0]}

        with mock.patch.object(chain, "rest_get", side_effect=fake_rest_get):
            committed = chain.ibc_unreceived_acks(
                chain=CHAIN, channel_id="channel-9", port_id="transfer", sequences=list(range(1, 251))
            )

        self.assertEqual(len(urls), 3)
        self.assertEqual(committed, list(range(2, 251, 2)))
        self.assertIn("/channels/channel-9/ports/transfer/packet_commitments/", urls[0])
        self.assertTrue(all(url.endswith("/unreceived_acks") for url in urls))

    def test_unreceived_acks_with_nothing_to_check_makes_no_request(self) -> None:
        with mock.patch.object(chain, "rest_get") as rest_get:
            result = chain.ibc_unreceived_acks(chain=CHAIN, channel_id="channel-9", port_id="transfer", sequences=[])

        self.assertEqual(result, [])
        rest_get.assert_not_called()

    def test_unreceived_with_nothing_to_check_makes_no_request(self) -> None:
        with mock.patch.object(chain, "rest_get") as rest_get:
            result = chain.ibc_unreceived_packets(chain=CHAIN, channel_id="channel-9", port_id="icahost", sequences=[])

        self.assertEqual(result, [])
        rest_get.assert_not_called()

    def test_channel_end_with_no_counterparty_channel_yet(self) -> None:
        response = {
            "channel": {
                "state": "STATE_INIT",
                "counterparty": {"port_id": "icahost", "channel_id": ""},
                "connection_hops": ["connection-146"],
            }
        }

        with mock.patch.object(chain, "rest_get", return_value=response):
            end = chain.ibc_channel_end(chain=CHAIN, channel_id="channel-768", port_id="icacontroller-x.DELEGATION")

        self.assertEqual(end.state, "INIT")
        self.assertIsNone(end.counterparty_channel)
        self.assertEqual(end.connection_id, "connection-146")


class RpcTest(unittest.TestCase):
    def setUp(self) -> None:
        chain._latest_height.cache_clear()
        chain.block_time.cache_clear()

    def test_latest_tx_searches_the_recent_window_and_returns_block_time(self) -> None:
        responses = [
            {"result": {"sync_info": {"latest_block_height": "500000"}}},
            {"result": {"txs": [{"height": "424242"}]}},
            {"result": {"header": {"time": "2026-09-30T11:00:00Z"}}},
        ]

        with mock.patch.object(chain, "get_json", side_effect=responses) as get_json:
            result = chain.rpc_tx_search_latest(chain=CHAIN, query="recv_packet.packet_dst_channel='channel-1'")

        self.assertEqual(result, chain.TxRef(height=424242, time="2026-09-30T11:00:00Z"))
        search_url = urllib.parse.unquote_plus(get_json.call_args_list[1].kwargs["url"])
        self.assertIn("AND tx.height>=400000", search_url)
        self.assertIn("per_page=1", search_url)
        self.assertIn("order_by=\"desc\"", search_url)
        self.assertIn("/header?height=424242", get_json.call_args_list[2].kwargs["url"])

    def test_falls_back_to_the_whole_index_when_the_window_is_empty(self) -> None:
        responses = [
            {"result": {"sync_info": {"latest_block_height": "500000"}}},
            {"result": {"txs": []}},
            {"result": {"txs": [{"height": "12"}]}},
            {"result": {"header": {"time": "2026-01-01T00:00:00Z"}}},
        ]

        with mock.patch.object(chain, "get_json", side_effect=responses) as get_json:
            result = chain.rpc_tx_search_latest(chain=CHAIN, query="x='y'")

        self.assertEqual(result, chain.TxRef(height=12, time="2026-01-01T00:00:00Z"))
        self.assertNotIn("tx.height", urllib.parse.unquote_plus(get_json.call_args_list[2].kwargs["url"]))

    def test_latest_tx_is_none_when_the_index_holds_nothing(self) -> None:
        responses = [
            {"result": {"sync_info": {"latest_block_height": "500000"}}},
            {"result": {"txs": []}},
            {"result": {"txs": []}},
        ]

        with mock.patch.object(chain, "get_json", side_effect=responses):
            self.assertIsNone(chain.rpc_tx_search_latest(chain=CHAIN, query="x='y'"))

    def test_block_time_is_cached_per_height(self) -> None:
        response = {"result": {"header": {"time": "2026-09-30T11:00:00Z"}}}

        with mock.patch.object(chain, "get_json", return_value=response) as get_json:
            first = chain.block_time(rpc=CHAIN.rpc, height=7)
            second = chain.block_time(rpc=CHAIN.rpc, height=7)

        self.assertEqual((first, second), ("2026-09-30T11:00:00Z", "2026-09-30T11:00:00Z"))
        self.assertEqual(get_json.call_count, 1)

    def test_latest_height_is_reused_within_the_cache_window(self) -> None:
        with mock.patch.object(chain, "get_json", return_value={"result": {"sync_info": {"latest_block_height": "9"}}}) as get_json:
            with mock.patch.object(chain.time, "time", return_value=1000.0):
                first = chain.rpc_latest_height(chain=CHAIN)
                second = chain.rpc_latest_height(chain=CHAIN)

        self.assertEqual((first, second), (9, 9))
        self.assertEqual(get_json.call_count, 1)


class GetJsonTest(unittest.TestCase):
    def test_retries_once_after_a_network_error(self) -> None:
        with mock.patch.object(chain, "_fetch_json", side_effect=[urllib.error.URLError("boom"), {"ok": True}]) as fetch:
            self.assertEqual(chain.get_json(url="https://x"), {"ok": True})

        self.assertEqual(fetch.call_count, 2)

    def test_retries_once_after_an_error_urllib_does_not_wrap(self) -> None:
        with mock.patch.object(
            chain, "_fetch_json", side_effect=[http.client.RemoteDisconnected(), {"ok": True}]
        ) as fetch:
            self.assertEqual(chain.get_json(url="https://x"), {"ok": True})

        self.assertEqual(fetch.call_count, 2)

    def test_second_failure_propagates(self) -> None:
        with mock.patch.object(chain, "_fetch_json", side_effect=TimeoutError("slow")) as fetch:
            with self.assertRaises(TimeoutError):
                chain.get_json(url="https://x")

        self.assertEqual(fetch.call_count, 2)


class ErrorBoundaryTest(unittest.TestCase):
    def test_network_and_decode_errors_become_zone_errors(self) -> None:
        for error in (
            urllib.error.URLError("refused"),
            TimeoutError("slow"),
            http.client.IncompleteRead(b""),
            http.client.RemoteDisconnected(),
            ConnectionResetError(),
            json.JSONDecodeError("bad", "x", 0),
            KeyError("status"),
        ):
            with self.subTest(error=type(error).__name__):
                result = chain.within_error_boundary(chain_id="zone-1", work=self.raiser(error))

                self.assertIsInstance(result, chain.ZoneError)
                self.assertEqual(result.chain_id, "zone-1")
                self.assertIn(type(error).__name__, result.error)

    def test_successful_work_passes_through(self) -> None:
        self.assertEqual(chain.within_error_boundary(chain_id="zone-1", work=lambda: 5), 5)

    def test_other_errors_propagate(self) -> None:
        with self.assertRaises(ZeroDivisionError):
            chain.within_error_boundary(chain_id="zone-1", work=self.raiser(ZeroDivisionError()))

    def test_optional_lookup_failure_is_none_and_success_passes_through(self) -> None:
        self.assertIsNone(chain.optional(self.raiser(http.client.IncompleteRead(b""))))
        self.assertEqual(chain.optional(lambda: 5), 5)

    def test_optional_lets_other_errors_propagate(self) -> None:
        with self.assertRaises(ZeroDivisionError):
            chain.optional(self.raiser(ZeroDivisionError()))

    def test_succeeded_keeps_only_the_requested_type(self) -> None:
        error = chain.ZoneError(chain_id="zone-1", error="boom")

        self.assertEqual(chain.succeeded([1, error, 2], int), [1, 2])
        self.assertEqual(chain.succeeded([1, error], chain.ZoneError), [error])

    @staticmethod
    def raiser(error: Exception):
        def raise_error() -> None:
            raise error

        return raise_error


if __name__ == "__main__":
    unittest.main()
