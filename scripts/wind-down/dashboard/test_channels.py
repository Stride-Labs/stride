import datetime
import unittest
import urllib.error
from unittest import mock

import chain
import channels

NOW = datetime.datetime(2026, 9, 30, 12, 0, 0, tzinfo=datetime.UTC)
STRIDE = chain.Chain(chain_id="stride-1", rest="https://stride.rest", rpc="https://stride.rpc")
HOST = chain.Chain(chain_id="host-1", rest="https://host.rest", rpc="https://host.rpc")


def minutes_ago(minutes: int) -> str:
    return (NOW - datetime.timedelta(minutes=minutes)).isoformat()


def flow(oldest_sent_minutes_ago: int | None, pending_packets: int | None = 1, pending_acks: int | None = 0):
    return channels.PacketFlow(
        pending_packets=pending_packets,
        pending_acks=pending_acks,
        capped=False,
        oldest_sequence=7,
        oldest_sent_at=None if oldest_sent_minutes_ago is None else minutes_ago(oldest_sent_minutes_ago),
        oldest_time_source=None if oldest_sent_minutes_ago is None else channels.TimeSource.SENT,
    )


class SplitPendingTest(unittest.TestCase):
    def test_unreceived_are_packets_and_received_are_acks(self) -> None:
        split = channels.split_pending(commitments=[3, 4, 5, 6, 9], unreceived=[4, 9, 12])

        self.assertEqual(split.packets, [4, 9])
        self.assertEqual(split.acks, [3, 5, 6])

    def test_nothing_unreceived_means_everything_awaits_its_ack(self) -> None:
        split = channels.split_pending(commitments=[1, 2], unreceived=[])

        self.assertEqual(split.packets, [])
        self.assertEqual(split.acks, [1, 2])


class ChannelStatusTest(unittest.TestCase):
    def status(self, end_states: list[str | None], flows: list[channels.PacketFlow]) -> channels.Status:
        return channels.channel_status(end_states=end_states, flows=flows, now=NOW)

    def test_closed_on_either_end(self) -> None:
        self.assertEqual(self.status(["CLOSED", "CLOSED"], []), channels.Status.CLOSED)
        self.assertEqual(self.status(["OPEN", "CLOSED"], []), channels.Status.CLOSED)

    def test_closed_outranks_a_half_finished_handshake(self) -> None:
        self.assertEqual(self.status(["INIT", "CLOSED"], []), channels.Status.CLOSED)

    def test_handshake_stuck_when_an_end_is_not_open(self) -> None:
        self.assertEqual(self.status(["INIT", None], []), channels.Status.HANDSHAKE_STUCK)
        self.assertEqual(self.status(["OPEN", "TRYOPEN"], []), channels.Status.HANDSHAKE_STUCK)

    def test_handshake_stuck_outranks_pending_packets(self) -> None:
        self.assertEqual(self.status(["INIT", None], [flow(120)]), channels.Status.HANDSHAKE_STUCK)

    def test_stuck_when_the_oldest_pending_is_over_thirty_minutes(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [flow(31)]), channels.Status.STUCK)

    def test_stuck_when_only_a_pending_ack_is_old(self) -> None:
        old_ack_only = flow(90, pending_packets=0, pending_acks=2)

        self.assertEqual(self.status(["OPEN", "OPEN"], [old_ack_only]), channels.Status.STUCK)

    def test_old_inbound_flow_makes_the_channel_stuck(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [flow(2), flow(45)]), channels.Status.STUCK)

    def test_pending_when_younger_than_thirty_minutes(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [flow(29)]), channels.Status.PENDING)

    def test_exactly_thirty_minutes_is_still_pending(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [flow(30)]), channels.Status.PENDING)

    def test_pending_when_the_age_is_unknown(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [flow(None)]), channels.Status.PENDING)

    def test_ok_when_nothing_is_pending(self) -> None:
        self.assertEqual(self.status(["OPEN", "OPEN"], [channels.NO_PENDING]), channels.Status.OK)
        self.assertEqual(self.status(["OPEN", "OPEN"], []), channels.Status.OK)

    def test_unreachable_end_says_nothing(self) -> None:
        self.assertEqual(self.status(["OPEN", channels.STATE_UNKNOWN], []), channels.Status.OK)
        self.assertEqual(self.status(["CLOSED", channels.STATE_UNKNOWN], []), channels.Status.CLOSED)
        self.assertEqual(self.status(["INIT", channels.STATE_UNKNOWN], []), channels.Status.HANDSHAKE_STUCK)

    def test_single_known_end_for_holder_routes(self) -> None:
        self.assertEqual(self.status(["OPEN"], []), channels.Status.OK)
        self.assertEqual(self.status(["CLOSED"], []), channels.Status.CLOSED)

    def test_worst_status_picks_the_most_severe(self) -> None:
        statuses = [channels.Status.OK, channels.Status.PENDING, channels.Status.STUCK, channels.Status.PENDING]

        self.assertEqual(channels.worst_status(statuses=statuses), channels.Status.STUCK)
        self.assertEqual(channels.worst_status(statuses=[channels.Status.OK]), channels.Status.OK)
        self.assertEqual(
            channels.worst_status(statuses=[channels.Status.HANDSHAKE_STUCK, channels.Status.CLOSED]),
            channels.Status.CLOSED,
        )


class PacketFlowTest(unittest.TestCase):
    """_packet_flow against a faked chain: the effects are the counts, the oldest sequence, and the null cases."""

    def run_flow(
        self,
        commitments: chain.PacketCommitments,
        unreceived: list[int] | Exception,
        sent_at: str | None,
        received_at: str | None = None,
    ):
        unreceived_patch = (
            mock.patch.object(chain, "ibc_unreceived_packets", side_effect=unreceived)
            if isinstance(unreceived, Exception)
            else mock.patch.object(chain, "ibc_unreceived_packets", return_value=unreceived)
        )
        sent = None if sent_at is None else chain.TxRef(height=1, time=sent_at)
        received = None if received_at is None else chain.TxRef(height=2, time=received_at)
        with (
            mock.patch.object(chain, "ibc_packet_commitments", return_value=commitments),
            unreceived_patch,
            mock.patch.object(chain, "rpc_tx_search_latest", side_effect=[sent, received]) as search,
        ):
            self.search = search
            return channels._packet_flow(
                sender=STRIDE,
                sender_channel="channel-1",
                sender_port="transfer",
                receiver=HOST,
                receiver_channel="channel-2",
                receiver_port="transfer",
            )

    def test_splits_and_reports_the_oldest(self) -> None:
        result = self.run_flow(
            chain.PacketCommitments(sequences=[10, 11, 12, 13], capped=False), unreceived=[12, 13], sent_at=minutes_ago(50)
        )

        self.assertEqual(result.pending_packets, 2)
        self.assertEqual(result.pending_acks, 2)
        self.assertEqual(result.oldest_sequence, 10)
        self.assertEqual(result.oldest_sent_at, minutes_ago(50))
        self.assertEqual(result.oldest_time_source, channels.TimeSource.SENT)
        self.assertFalse(result.capped)
        self.search.assert_called_once()

    def test_no_commitments_is_no_pending(self) -> None:
        result = self.run_flow(chain.PacketCommitments(sequences=[], capped=False), unreceived=[], sent_at=None)

        self.assertEqual(result, channels.NO_PENDING)

    def test_unknown_send_time_leaves_the_age_empty(self) -> None:
        result = self.run_flow(chain.PacketCommitments(sequences=[4], capped=False), unreceived=[4], sent_at=None)

        self.assertEqual(result.pending_packets, 1)
        self.assertIsNone(result.oldest_sent_at)
        self.assertIsNone(result.oldest_time_source)
        self.assertIsNone(result.age_seconds(now=NOW))
        self.search.assert_called_once()  # a pending packet has no receive tx to ask for

    def test_pending_ack_missing_from_the_send_index_ages_from_its_receive(self) -> None:
        result = self.run_flow(
            chain.PacketCommitments(sequences=[5729], capped=False),
            unreceived=[],
            sent_at=None,
            received_at=minutes_ago(90),
        )

        self.assertEqual(result.oldest_sent_at, minutes_ago(90))
        self.assertEqual(result.oldest_time_source, channels.TimeSource.RECEIVED)
        self.assertEqual(channels.channel_status(["OPEN", "OPEN"], [result], NOW), channels.Status.STUCK)
        receive_query = self.search.call_args_list[1].kwargs
        self.assertIs(receive_query["chain"], HOST)
        self.assertEqual(
            receive_query["query"], "recv_packet.packet_dst_channel='channel-2' AND recv_packet.packet_sequence='5729'"
        )

    def test_pending_ack_with_no_receive_tx_either_stays_unknown_age(self) -> None:
        result = self.run_flow(
            chain.PacketCommitments(sequences=[5729], capped=False), unreceived=[], sent_at=None, received_at=None
        )

        self.assertIsNone(result.oldest_sent_at)
        self.assertIsNone(result.oldest_time_source)
        self.assertEqual(channels.channel_status(["OPEN", "OPEN"], [result], NOW), channels.Status.PENDING)

    def test_receive_lookup_failure_is_unknown_age_not_an_error(self) -> None:
        with (
            mock.patch.object(chain, "ibc_packet_commitments", return_value=chain.PacketCommitments([5], False)),
            mock.patch.object(chain, "ibc_unreceived_packets", return_value=[]),
            mock.patch.object(chain, "rpc_tx_search_latest", side_effect=[None, TimeoutError("slow")]),
        ):
            result = channels._packet_flow(
                sender=STRIDE,
                sender_channel="channel-1",
                sender_port="transfer",
                receiver=HOST,
                receiver_channel="channel-2",
                receiver_port="transfer",
            )

        self.assertIsNone(result.oldest_sent_at)
        self.assertEqual(result.pending_acks, 1)

    def test_receiver_failure_leaves_counts_unknown_but_keeps_the_oldest(self) -> None:
        result = self.run_flow(
            chain.PacketCommitments(sequences=[8, 9], capped=True), unreceived=TimeoutError("slow"), sent_at=None
        )

        self.assertIsNone(result.pending_packets)
        self.assertIsNone(result.pending_acks)
        self.assertEqual(result.oldest_sequence, 8)
        self.assertTrue(result.capped)

    def test_age_is_measured_from_the_send_time(self) -> None:
        self.assertEqual(flow(45).age_seconds(now=NOW), 45 * 60)


class IcaSpecTest(unittest.TestCase):
    def test_init_channel_has_no_host_side(self) -> None:
        spec = channels._ica_spec(
            ica={
                "type": "DELEGATION",
                "port_id": "icacontroller-laozi-mainnet.DELEGATION",
                "channel_id": "channel-768",
                "counterparty_channel_id": "",
                "state": "STATE_INIT",
                "counterparty_state": "",
            }
        )

        self.assertEqual((spec.stride_state, spec.host_state, spec.host_channel), ("INIT", None, None))
        self.assertEqual(spec.host_port_id, "icahost")

    def test_closed_channel_keeps_both_states(self) -> None:
        spec = channels._ica_spec(
            ica={
                "type": "DELEGATION",
                "port_id": "icacontroller-celestia.DELEGATION",
                "channel_id": "channel-870",
                "counterparty_channel_id": "channel-704",
                "state": "STATE_CLOSED",
                "counterparty_state": "STATE_CLOSED",
            }
        )

        self.assertEqual((spec.stride_state, spec.host_state, spec.host_channel), ("CLOSED", "CLOSED", "channel-704"))

    def test_init_channel_row_is_handshake_stuck_not_an_error(self) -> None:
        spec = channels.ChannelSpec(
            name="WITHDRAWAL",
            port_id="icacontroller-x.WITHDRAWAL",
            host_port_id="icahost",
            stride_channel="channel-769",
            host_channel=None,
            stride_state="INIT",
            host_state=None,
        )

        with (
            mock.patch.object(chain, "ibc_packet_commitments", return_value=chain.PacketCommitments([], False)),
            mock.patch.object(chain, "rpc_tx_search_latest", return_value=None),
        ):
            row = channels._channel_row(spec=spec, stride=STRIDE, host=HOST, now=NOW)

        self.assertEqual(row.status, channels.Status.HANDSHAKE_STUCK)
        self.assertIsNone(row.inbound)
        self.assertIsNone(row.last_received)


class HostOutageTest(unittest.TestCase):
    ENTRY = {
        "chain_id": "haqq_11235-1",
        "client_id": "07-tendermint-1",
        "counterparty_client_id": "07-tendermint-2",
        "connection_id": "connection-1",
        "counterparty_connection_id": "connection-2",
        "transfer_channel_id": "channel-240",
        "ica_channels": [
            {
                "type": "DELEGATION",
                "port_id": "icacontroller-haqq_11235-1.DELEGATION",
                "channel_id": "channel-869",
                "counterparty_channel_id": "channel-10",
                "state": "STATE_CLOSED",
                "counterparty_state": "STATE_CLOSED",
            }
        ],
    }

    def test_dead_host_rest_keeps_the_zone_and_its_closed_channel(self) -> None:
        zone = channels.config.ZONES_BY_CHAIN_ID["haqq_11235-1"]
        stride_end = chain.ChannelEnd(
            state="OPEN", counterparty_port="transfer", counterparty_channel="channel-7", connection_id="connection-1"
        )
        stride_health = chain.ClientHealth("stride-1", "07-tendermint-1", "Active", None, 100.0)
        bad_gateway = urllib.error.HTTPError("https://host", 502, "Bad Gateway", {}, None)

        def stride_or_fail(on_stride: object):
            def answer(chain: chain.Chain, **_: str) -> object:
                if chain is not STRIDE:
                    raise bad_gateway
                return on_stride

            return answer

        with (
            mock.patch.object(chain, "stride_chain", return_value=STRIDE),
            mock.patch.object(chain, "zone_chain", return_value=HOST),
            mock.patch.object(chain, "ibc_channel_end", side_effect=stride_or_fail(stride_end)),
            mock.patch.object(chain, "ibc_client_health", side_effect=stride_or_fail(stride_health)),
            mock.patch.object(
                chain, "ibc_connection_end", side_effect=stride_or_fail(chain.ConnectionEnd("OPEN", "a", "b", "c"))
            ),
            mock.patch.object(chain, "ibc_packet_commitments", side_effect=stride_or_fail(chain.PacketCommitments([], False))),
            mock.patch.object(chain, "rpc_tx_search_latest", side_effect=TimeoutError("slow")),
        ):
            result = channels._zone_channels(zone=zone, feed_by_chain={"haqq_11235-1": self.ENTRY}, now=NOW)

        by_name = {row.name: row for row in result.channels}
        self.assertEqual(by_name["DELEGATION"].status, channels.Status.CLOSED)
        self.assertEqual(by_name["transfer"].host_state, channels.STATE_UNKNOWN)
        self.assertEqual(by_name["transfer"].status, channels.Status.OK)
        self.assertIsNone(by_name["transfer"].inbound)
        self.assertEqual(result.status, channels.Status.CLOSED)
        self.assertEqual(result.stride_client, stride_health)
        self.assertIsNone(result.host_client)
        self.assertIsNone(result.host_connection_state)
        self.assertEqual(result.host_error, "HTTPError: HTTP Error 502: Bad Gateway")
        self.assertEqual(result.stride_connection_state, "OPEN")


def client(chain_id: str, client_id: str, status: str = "Active", remaining: float | None = 86400.0):
    return chain.ClientHealth(
        chain_id=chain_id, client_id=client_id, status=status, expires_at=None, seconds_remaining=remaining
    )


def channel_row(name: str, status: channels.Status, outbound: channels.PacketFlow, inbound=None, states=("OPEN", "OPEN")):
    return channels.ChannelRow(
        name=name,
        port_id="p",
        stride_channel="channel-1",
        host_channel="channel-2",
        stride_state=states[0],
        host_state=states[1],
        outbound=outbound,
        inbound=inbound,
        last_sent=None,
        last_received=None,
        last_ack=None,
        status=status,
    )


class TilesTest(unittest.TestCase):
    def test_tiles_aggregate_zones_legs_and_routes(self) -> None:
        shared_osmosis_client = client("osmosis-1", "07-tendermint-1", remaining=5 * 86400.0)
        zone = channels.ZoneChannels(
            chain_id="celestia",
            symbol="TIA",
            stride_client=client("stride-1", "07-tendermint-137", remaining=9 * 86400.0),
            host_client=client("celestia", "07-tendermint-0", remaining=2 * 86400.0),
            stride_connection_state="OPEN",
            host_connection_state="OPEN",
            host_error=None,
            status=channels.Status.CLOSED,
            channels=[
                channel_row(
                    "transfer",
                    channels.Status.STUCK,
                    outbound=flow(90, pending_packets=3, pending_acks=1),
                    inbound=flow(10, pending_packets=0, pending_acks=2),
                ),
                channel_row("DELEGATION", channels.Status.CLOSED, channels.NO_PENDING, states=("CLOSED", "CLOSED")),
                channel_row("WITHDRAWAL", channels.Status.HANDSHAKE_STUCK, channels.NO_PENDING, states=("INIT", None)),
                channel_row("FEE", channels.Status.OK, channels.NO_PENDING),
            ],
        )
        leg = channels.LegRow(
            chain_id="celestia",
            symbol="TIA",
            host_channel="channel-2",
            osmosis_channel="channel-6994",
            host_state="OPEN",
            osmosis_state="OPEN",
            host_client=client("celestia", "07-tendermint-0", remaining=2 * 86400.0),  # duplicate of the zone's
            osmosis_client=shared_osmosis_client,
            last_received=None,
            last_ack=None,
            status=channels.Status.OK,
        )
        route = channels.RouteRow(
            chain="Cosmos Hub",
            osmosis_channel="channel-0",
            relayed_by="free",
            state="OPEN",
            counterparty_channel="channel-141",
            client=client("osmosis-1", "07-tendermint-0", status="Expired", remaining=-10.0),
            last_received=None,
            last_ack=None,
            status=channels.Status.OK,
        )

        tiles = channels.build_tiles(zones=[zone], legs=[leg], routes=[route], now=NOW)

        self.assertEqual((tiles.channels_open, tiles.channels_total), (2, 4))
        self.assertEqual((tiles.channels_closed, tiles.channels_handshake_stuck), (1, 1))
        self.assertEqual((tiles.pending_packets, tiles.pending_packet_channels), (3, 1))
        self.assertEqual((tiles.pending_acks, tiles.pending_ack_channels), (3, 2))
        self.assertEqual(tiles.oldest_pending_seconds, 90 * 60)
        self.assertEqual(tiles.oldest_pending_where, "celestia transfer")
        # 4 distinct clients once the duplicate is merged; the expired one is not live and never the soonest.
        self.assertEqual((tiles.clients_live, tiles.clients_total), (3, 4))
        self.assertEqual(tiles.soonest_expiry_seconds, 2 * 86400.0)
        self.assertEqual(tiles.soonest_expiry_where, "celestia 07-tendermint-0")

    def test_empty_tab_has_no_oldest_or_expiry(self) -> None:
        tiles = channels.build_tiles(zones=[], legs=[], routes=[], now=NOW)

        self.assertIsNone(tiles.oldest_pending_seconds)
        self.assertIsNone(tiles.soonest_expiry_seconds)
        self.assertEqual((tiles.channels_open, tiles.channels_total, tiles.clients_total), (0, 0, 0))

    def test_unknown_age_is_not_reported_as_oldest(self) -> None:
        zone = channels.ZoneChannels(
            chain_id="z",
            symbol="Z",
            stride_client=client("stride-1", "a"),
            host_client=client("z", "b"),
            stride_connection_state="OPEN",
            host_connection_state="OPEN",
            host_error=None,
            status=channels.Status.PENDING,
            channels=[channel_row("FEE", channels.Status.PENDING, outbound=flow(None, pending_packets=None, pending_acks=None))],
        )

        tiles = channels.build_tiles(zones=[zone], legs=[], routes=[], now=NOW)

        self.assertIsNone(tiles.oldest_pending_seconds)
        self.assertEqual(tiles.pending_packets, 0)


if __name__ == "__main__":
    unittest.main()
