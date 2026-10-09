import dataclasses
import datetime
import json
import unittest
import urllib.error
from unittest import mock

import chain
import channels
import config

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
            mock.patch.object(chain, "rest_get", side_effect=bad_gateway),
            mock.patch.object(chain, "ibc_client_status", side_effect=bad_gateway),
        ):
            result = channels._zone_channels(
                zone=zone, feed_by_chain={"haqq_11235-1": self.ENTRY}, host_zones=HOST_ZONES, now=NOW
            )

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
        # Every pre-flight lookup hits the dead host: all n/a, so the zone cannot pass.
        preflight = result.preflight
        self.assertEqual(preflight.withdrawal_ica_address, "haqq1withdrawal")
        self.assertEqual(
            [
                preflight.withdraw_address_ok,
                preflight.allow_messages_ok,
                preflight.osmosis_leg_ok,
                preflight.host_client_of_stride_ok,
            ],
            [None, None, None, None],
        )
        self.assertFalse(preflight.all_ok)


HOST_ZONES = {
    "haqq_11235-1": {"delegation_ica_address": "haqq1delegation", "withdrawal_ica_address": "haqq1withdrawal"},
}


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


DELEGATION_ICA = "cosmos1delegation"
WITHDRAWAL_ICA = "cosmos1withdrawal"
ZONE_RECORD = {"delegation_ica_address": DELEGATION_ICA, "withdrawal_ica_address": WITHDRAWAL_ICA}
WITHDRAW_PATH = f"/cosmos/distribution/v1beta1/delegators/{DELEGATION_ICA}/withdraw_address"
PARAMS_PATH = "/ibc/apps/interchain_accounts/host/v1/params"
CLIENT_STATES_PATH = "/ibc/core/client/v1/client_states/07-tendermint-9"


class AllowMessagesTest(unittest.TestCase):
    def missing(self, allow_messages: list[str], chain_id: str = "cosmoshub-4") -> list[str]:
        return channels.allow_messages_missing(allow_messages=allow_messages, chain_id=chain_id)

    def test_msg_transfer_is_required_everywhere(self) -> None:
        self.assertEqual(self.missing([]), [channels.MSG_TRANSFER])
        self.assertEqual(self.missing([channels.MSG_SEND]), [channels.MSG_TRANSFER])
        self.assertEqual(self.missing([channels.MSG_TRANSFER]), [])

    def test_a_wildcard_allows_everything_even_on_osmosis(self) -> None:
        self.assertEqual(self.missing(["*"]), [])
        self.assertEqual(self.missing(["*"], chain_id="osmosis-1"), [])

    def test_osmosis_also_needs_the_bank_send(self) -> None:
        self.assertEqual(self.missing([channels.MSG_TRANSFER], chain_id="osmosis-1"), [channels.MSG_SEND])
        self.assertEqual(self.missing([channels.MSG_SEND], chain_id="osmosis-1"), [channels.MSG_TRANSFER])
        self.assertEqual(self.missing([channels.MSG_TRANSFER, channels.MSG_SEND], chain_id="osmosis-1"), [])

    def test_other_zones_do_not_need_the_bank_send(self) -> None:
        self.assertEqual(self.missing([channels.MSG_TRANSFER, "/cosmos.gov.v1.MsgVote"]), [])


class LegTest(unittest.TestCase):
    def test_open_on_a_client_of_osmosis_passes(self) -> None:
        self.assertTrue(channels.leg_is_ok(state="OPEN", counterparty_chain_id="osmosis-1"))

    def test_closed_or_wrong_counterparty_fails(self) -> None:
        self.assertFalse(channels.leg_is_ok(state="CLOSED", counterparty_chain_id="osmosis-1"))
        self.assertFalse(channels.leg_is_ok(state="OPEN", counterparty_chain_id="osmosis-2"))


class PassesTest(unittest.TestCase):
    def passes(self, chain_id: str = "cosmoshub-4", **overrides: bool | None) -> bool:
        checks = {
            "withdraw_address_ok": True,
            "allow_messages_ok": True,
            "osmosis_leg_ok": True,
            "host_client_of_stride_ok": True,
            **overrides,
        }
        return channels.preflight_passes(chain_id=chain_id, **checks)

    def test_all_true_passes_and_any_false_or_null_fails(self) -> None:
        self.assertTrue(self.passes())
        for check in ("withdraw_address_ok", "allow_messages_ok", "osmosis_leg_ok", "host_client_of_stride_ok"):
            self.assertFalse(self.passes(**{check: False}), check)
            self.assertFalse(self.passes(**{check: None}), check)

    def test_osmosis_has_no_leg_so_a_null_leg_passes_there_only(self) -> None:
        self.assertTrue(self.passes(chain_id="osmosis-1", osmosis_leg_ok=None))
        self.assertFalse(self.passes(chain_id="osmosis-1", allow_messages_ok=False, osmosis_leg_ok=None))


class PreflightTest(unittest.TestCase):
    ENTRY = {"counterparty_client_id": "07-tendermint-5"}

    def run_preflight(
        self,
        chain_id: str = "cosmoshub-4",
        rest: dict[str, object] | None = None,
        client_status: str | Exception = "Active",
        leg_state: str = "OPEN",
        host_zone: dict[str, str] | None = ZONE_RECORD,
    ) -> channels.Preflight:
        zone = config.ZONES_BY_CHAIN_ID[chain_id]
        responses = {
            WITHDRAW_PATH: {"withdraw_address": WITHDRAWAL_ICA},
            PARAMS_PATH: {"params": {"host_enabled": True, "allow_messages": [channels.MSG_TRANSFER]}},
            CLIENT_STATES_PATH: {"client_state": {"chain_id": "osmosis-1"}},
            **(rest or {}),
        }

        def fake_rest_get(chain: chain.Chain, path: str, params: object = None) -> object:
            response = responses[path]
            if isinstance(response, Exception):
                raise response
            return response

        leg_end = chain.ChannelEnd(
            state=leg_state, counterparty_port="transfer", counterparty_channel="channel-1", connection_id="connection-3"
        )
        status_patch = (
            mock.patch.object(chain, "ibc_client_status", side_effect=client_status)
            if isinstance(client_status, Exception)
            else mock.patch.object(chain, "ibc_client_status", return_value=client_status)
        )
        with (
            mock.patch.object(chain, "rest_get", side_effect=fake_rest_get),
            mock.patch.object(chain, "ibc_channel_end", return_value=leg_end),
            mock.patch.object(chain, "ibc_connection_end", return_value=chain.ConnectionEnd("OPEN", "07-tendermint-9", "x", "y")),
            status_patch,
        ):
            return channels._preflight(zone=zone, host=HOST, entry=self.ENTRY, host_zone=host_zone)

    def test_everything_good_passes_with_the_details_filled_in(self) -> None:
        preflight = self.run_preflight()

        self.assertEqual(
            (
                preflight.withdraw_address_ok,
                preflight.allow_messages_ok,
                preflight.osmosis_leg_ok,
                preflight.host_client_of_stride_ok,
                preflight.all_ok,
            ),
            (True, True, True, True, True),
        )
        self.assertEqual((preflight.withdraw_address, preflight.withdrawal_ica_address), (WITHDRAWAL_ICA, WITHDRAWAL_ICA))
        self.assertEqual((preflight.allow_messages_count, preflight.allow_messages_wildcard, preflight.host_enabled), (1, False, True))
        self.assertEqual((preflight.osmosis_leg_state, preflight.osmosis_leg_chain_id), ("OPEN", "osmosis-1"))
        self.assertEqual((preflight.host_client_of_stride_status, preflight.host_client_of_stride_id), ("Active", "07-tendermint-5"))
        self.assertIsNone(preflight.osmosis_leg_note)
        json.dumps(dataclasses.asdict(preflight))

    def test_a_withdraw_address_that_is_not_the_withdrawal_ica_fails_and_carries_both(self) -> None:
        preflight = self.run_preflight(rest={WITHDRAW_PATH: {"withdraw_address": DELEGATION_ICA}})

        self.assertFalse(preflight.withdraw_address_ok)
        self.assertEqual((preflight.withdraw_address, preflight.withdrawal_ica_address), (DELEGATION_ICA, WITHDRAWAL_ICA))
        self.assertFalse(preflight.all_ok)

    def test_the_wildcard_allow_list_passes_and_a_missing_msg_transfer_fails(self) -> None:
        wildcard = self.run_preflight(rest={PARAMS_PATH: {"params": {"host_enabled": True, "allow_messages": ["*"]}}})
        narrow = self.run_preflight(rest={PARAMS_PATH: {"params": {"host_enabled": True, "allow_messages": ["/x.MsgY"]}}})

        self.assertEqual((wildcard.allow_messages_ok, wildcard.allow_messages_wildcard), (True, True))
        self.assertEqual((narrow.allow_messages_ok, narrow.allow_messages_missing), (False, [channels.MSG_TRANSFER]))

    def test_the_leg_fails_when_closed_or_on_the_wrong_chain(self) -> None:
        closed = self.run_preflight(leg_state="CLOSED")
        wrong = self.run_preflight(rest={CLIENT_STATES_PATH: {"client_state": {"chain_id": "other-1"}}})

        self.assertFalse(closed.osmosis_leg_ok)
        self.assertFalse(wrong.osmosis_leg_ok)
        self.assertEqual(wrong.osmosis_leg_chain_id, "other-1")

    def test_an_expired_host_client_of_stride_fails(self) -> None:
        preflight = self.run_preflight(client_status="Expired")

        self.assertFalse(preflight.host_client_of_stride_ok)
        self.assertEqual(preflight.host_client_of_stride_status, "Expired")

    def test_osmosis_has_no_leg_and_needs_the_bank_send(self) -> None:
        params = {PARAMS_PATH: {"params": {"host_enabled": True, "allow_messages": [channels.MSG_TRANSFER]}}}
        without_send = self.run_preflight(chain_id="osmosis-1", rest=params)
        with_send = self.run_preflight(
            chain_id="osmosis-1",
            rest={PARAMS_PATH: {"params": {"host_enabled": True, "allow_messages": [channels.MSG_TRANSFER, channels.MSG_SEND]}}},
        )

        self.assertIsNone(without_send.osmosis_leg_ok)
        self.assertEqual(without_send.osmosis_leg_note, "bank send")
        self.assertFalse(without_send.allow_messages_ok)
        self.assertFalse(without_send.all_ok)
        self.assertTrue(with_send.all_ok)

    def test_one_failed_lookup_is_null_without_hiding_the_others(self) -> None:
        preflight = self.run_preflight(
            rest={PARAMS_PATH: urllib.error.URLError("down")}, client_status=urllib.error.URLError("down")
        )

        self.assertIsNone(preflight.allow_messages_ok)
        self.assertIsNone(preflight.allow_messages_count)
        self.assertIsNone(preflight.host_client_of_stride_ok)
        self.assertTrue(preflight.withdraw_address_ok)
        self.assertTrue(preflight.osmosis_leg_ok)
        self.assertFalse(preflight.all_ok)

    def test_without_stride_host_zone_the_withdraw_check_is_null(self) -> None:
        preflight = self.run_preflight(host_zone=None)

        self.assertIsNone(preflight.withdraw_address_ok)
        self.assertIsNone(preflight.withdrawal_ica_address)
        self.assertFalse(preflight.all_ok)


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


class UnackedIcaPacketsTest(unittest.TestCase):
    def test_sums_ica_rows_and_ignores_the_transfer_row(self) -> None:
        rows = [
            channel_row(name=channels.TRANSFER_NAME, status=channels.Status.PENDING, outbound=flow(None, pending_packets=900, pending_acks=30)),
            channel_row(name="DELEGATION", status=channels.Status.PENDING, outbound=flow(None, pending_packets=14, pending_acks=0)),
            channel_row(name="WITHDRAWAL", status=channels.Status.PENDING, outbound=flow(None, pending_packets=0, pending_acks=2)),
        ]
        self.assertEqual(channels.unacked_ica_packets(rows=rows), 16)

    def test_unknown_count_makes_the_total_unknown(self) -> None:
        rows = [channel_row(name="FEE", status=channels.Status.OK, outbound=flow(None, pending_packets=None, pending_acks=0))]
        self.assertIsNone(channels.unacked_ica_packets(rows=rows))


class IcaChannelsOpenTest(unittest.TestCase):
    def rows(self, **states: tuple[str | None, str | None]) -> list[channels.ChannelRow]:
        transfer_row = channel_row(name=channels.TRANSFER_NAME, status=channels.Status.OK, outbound=flow(None), states=("CLOSED", "CLOSED"))
        return [
            transfer_row,
            *(
                channel_row(name=name, status=channels.Status.OK, outbound=flow(None), states=states.get(name, ("OPEN", "OPEN")))
                for name in ("DELEGATION", "WITHDRAWAL", "FEE", "REDEMPTION")
            ),
        ]

    def test_open_on_both_ends_passes_whatever_the_transfer_channel_does(self) -> None:
        self.assertTrue(channels.ica_channels_open(rows=self.rows()))

    def test_any_ica_end_not_open_fails(self) -> None:
        for name in ("DELEGATION", "WITHDRAWAL", "FEE", "REDEMPTION"):
            for states in (("CLOSED", "OPEN"), ("OPEN", "CLOSED"), ("INIT", "OPEN")):
                self.assertFalse(channels.ica_channels_open(rows=self.rows(**{name: states})))

    def test_an_unknown_state_is_null(self) -> None:
        self.assertIsNone(channels.ica_channels_open(rows=self.rows(FEE=("OPEN", None))))
        self.assertIsNone(channels.ica_channels_open(rows=self.rows(FEE=(None, None))))

    def test_unknown_outranks_a_known_failure(self) -> None:
        self.assertIsNone(channels.ica_channels_open(rows=self.rows(FEE=("OPEN", None), WITHDRAWAL=("CLOSED", "CLOSED"))))

    def test_no_ica_rows_is_null(self) -> None:
        self.assertIsNone(channels.ica_channels_open(rows=self.rows()[:1]))


class HaqqStateTest(unittest.TestCase):
    def state(self, **overrides: object) -> channels.HaqqState:
        fields = {
            "delegation_channel_id": channels.HAQQ_DELEGATION_CHANNEL,
            "delegation_state_stride": channels.STATE_CLOSED,
            "delegation_state_host": channels.STATE_OPEN,
            "delegation_commitments": list(channels.HAQQ_DELEGATION_SEQUENCES),
            "quiet_channel_commitments": {ica: 0 for ica in channels.HAQQ_QUIET_ICA_CHANNELS},
            "ok": False,
        }
        return channels.HaqqState(**{**fields, **overrides})

    def test_the_expected_pre_upgrade_state_passes(self) -> None:
        self.assertTrue(channels.haqq_state_passes(self.state()))

    def test_any_departure_fails(self) -> None:
        self.assertFalse(channels.haqq_state_passes(self.state(delegation_channel_id="channel-900")))
        self.assertFalse(channels.haqq_state_passes(self.state(delegation_state_stride=channels.STATE_OPEN)))
        self.assertFalse(channels.haqq_state_passes(self.state(delegation_state_host=channels.STATE_CLOSED)))
        self.assertFalse(channels.haqq_state_passes(self.state(delegation_state_host=None)))
        self.assertFalse(channels.haqq_state_passes(self.state(delegation_commitments=list(range(85, 98)))))
        self.assertFalse(channels.haqq_state_passes(self.state(quiet_channel_commitments={"FEE": 1})))


if __name__ == "__main__":
    unittest.main()
