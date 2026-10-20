"""Command sets for the Multisig tab: the exact generate / sign / multisign+broadcast commands per zone, built from
the Validators, Funds and Pools snapshots. Pure: no chain calls.

The Stride sets (slash queries, drains, ICA transfers, staketia claim balance) are signed by the F5 protocol-admin
multisig with `strided`; the pool-funding set is signed by the Osmosis vault (a multisig with the same member key
names) with `osmosisd`, as is the pool-creation set (one create-pool per planned pool from the Pools snapshot). Each
tx is done end to end, one at a time, online (no pre-assigned sequences): every signer's `tx sign` looks the
multisig's account number and sequence up itself. A set may hold several txs per zone; they are contiguous per
`chain_id` in `txs`, in the order they go out.
"""

import dataclasses
import hashlib
import json
from dataclasses import dataclass
from decimal import Decimal
from typing import Any

import config
import funds

MULTISIG_KEY = "F5"  # the multisig's name in the keyring; `multisign` needs it (the Osmosis vault uses the same name)
MULTISIG_ADDRESS = config.PROTOCOL_ADMIN
NODE = "https://stride-strd-rpc.polkachu.com:443"
CHAIN_ID = "stride-1"
OSMOSIS_NODE = "https://osmosis-strd-rpc.polkachu.com:443"
OSMOSIS_VAULT_ADDRESS = config.OSMOSIS_VAULT

WORKDIR = "/tmp/wind-down"
PLACEHOLDER_VALOPER = "<LIVE_TEST_VALOPER>"
PLACEHOLDER_DENOM = "<HOST_DENOM>"
PLACEHOLDER_BALANCE = "<BALANCE>"
PLACEHOLDER_OSMOSIS_DENOM = "<OSMOSIS_DENOM>"
PLACEHOLDER_REST_AMOUNT = "<ALLOCATION_MINUS_VAULT_SHARES>"
ANYONE = "anyone"
NO_SNAPSHOT_REASON = "waiting for the Validators snapshot"
ALLOCATION_UNKNOWN_REASON = "the pool's allocation is not known yet (see the Pools tab)"
ALLOCATION_EMPTY_REASON = "the allocation is 0: nothing to join"

# The order the four admin transfers go out (spec §3): fees first, delegations after the withdrawals that feed them.
ICA_ORDER = (funds.IcaType.FEE, funds.IcaType.WITHDRAWAL, funds.IcaType.DELEGATION, funds.IcaType.REDEMPTION)
TRANSFER_GAS = 600_000
TRANSFER_FEES = "3000ustrd"
STAKETIA_TEST_AMOUNT = 1_000_000  # utia
SLASH_QUERY_GAS = 300_000  # an ICQ submission, no ICA: the signing runbook's figure
SLASH_QUERY_FEES = "1500ustrd"
STAKETIA_CHAIN_ID = "celestia"
POOL_GAS = 1_500_000
# About 3x Osmosis's EIP-1559 base fee (0.03 uosmo/gas on 2026-10-08, floor 0.01): CheckTx rejects a fee below the
# base fee at broadcast time, and a multisig tx cannot be re-priced without a new signing round.
OSMOSIS_GAS_PRICE = "0.1uosmo"
CREATE_POOL_CODE_ID = "996"  # the transmuter v3.2.0 code on osmosis-1 (funds.TRANSMUTER_CODE_ID)
# The 2026-09-25 creations used 1.44M gas each (1M of it the tokenfactory denom creation), flat per message.
CREATE_POOL_MSG_GAS = 1_500_000
CREATE_POOL_TX_OVERHEAD_GAS = 100_000
# 18 messages is ~27M gas, under half the 60M mempool cap per tx (31 would fit): a failing message reverts the whole
# bundle and burns its fee, so the blast radius stays at half the pools.
CREATE_POOL_BUNDLE_SIZE = 18
CREATION_GROUP = "all zones"  # the bundles span zones, so the tab shows them under one heading
WAITING_SHOWN = 8  # pools listed by name in the waiting tx before "and N more"
ONE_AT_A_TIME = "One tx at a time: generate, two signatures, multisign, broadcast, then confirm it landed before the next."

POOL_KIND_UNRECOGNISED = "unrecognised"  # a pool no Stride channel backs: reported on the Pools tab, never funded
POOL_KIND_CANONICAL = "canonical"


@dataclass(frozen=True)
class SlashQuery:
    """A validator whose shares-to-tokens rate moved after the eve's hot-key refresh: update-delegation re-queries it
    and applies the slash. Admin-gated from v35, so the multisig signs."""

    chain_id: str
    valoper: str
    moniker: str
    done: bool = False  # the callback has landed: the card stays for the record, not to re-run


SLASH_QUERIES = (
    SlashQuery(chain_id="cosmoshub-4", valoper="cosmosvaloper157v7tczs40axfgejp2m43kwuzqe0wsy0rv8puv", moniker="posthuman"),
    SlashQuery(chain_id="cosmoshub-4", valoper="cosmosvaloper140l6y2gp3gxvay6qtn70re7z2s0gn57zfd832j", moniker="lavenderfive"),
    SlashQuery(chain_id="cosmoshub-4", valoper="cosmosvaloper1wqy2s6nwnxj57l0l5rdjxxr646p3al6y70435m", moniker="klubstaking"),
    SlashQuery(chain_id="cosmoshub-4", valoper="cosmosvaloper1wrx0x9m9ykdhw9sg04v7uljme53wuj03aa5d4f", moniker="meria"),
    SlashQuery(chain_id="juno-1", valoper="junovaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcqcnylw", moniker="sg1", done=True),
    SlashQuery(
        chain_id="phoenix-1", valoper="terravaloper1lelhxdzwn9ddecv6sv0kcxj5tguurxnzcfs5wf", moniker="orbitalcommand", done=True
    ),
)


@dataclass(frozen=True)
class ValidatorWeightChange:
    """A weight the multisig sets so an epoch unbonding has slack. The epoch's split must place the whole amount into
    the validators above their target; when every weighted validator is at or above target the capacity equals the
    amount exactly, and the 1-base-unit trim on each fully drained slashed validator leaves a remainder nothing can
    absorb (the Hub's epoch 1500 on 2026-10-12). Weight 1 on an empty validator puts it under target, which makes the
    others' excess exceed the amount; nothing delegates to it, since v35 no longer rebalances."""

    chain_id: str
    valoper: str
    moniker: str
    weight: int
    why: str


VALIDATOR_WEIGHT_CHANGES = (
    ValidatorWeightChange(
        chain_id="cosmoshub-4",
        valoper="cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn",
        moniker="cosmostation",
        weight=1,
        why="epoch 1500 failed by 10 uatom with zero slack; weight 1 here gives the epoch 1504 retry ~195 ATOM of slack",
    ),
)
# Rewrites the whole host zone: the Hub's 183 validators used 823,100 gas on 2026-10-12 (a 300,000 attempt ran out).
VALIDATOR_WEIGHTS_GAS = 1_200_000
VALIDATOR_WEIGHTS_FEES = "6000ustrd"


@dataclass(frozen=True)
class ChannelRestore:
    """A closed ICA channel to reopen with restore-interchain-account. The msg is permissionless (any signer may
    send it); it goes through the multisig so the operators' own keys are not tied to a handshake."""

    chain_id: str
    connection_id: str  # the zone's connection on Stride (host zone `connection_id`)
    account_type: str  # the ICA owner suffix: {chain_id}.{account_type}
    closed_channel: str  # the newest channel of that port, STATE_CLOSED when the set was written


CHANNEL_RESTORES = (
    ChannelRestore(chain_id="dydx-mainnet-1", connection_id="connection-123", account_type="DELEGATION", closed_channel="channel-824"),
)
# Registers the account (a channel OpenInit) and resets the zone's delegation, deposit, unbonding and LSM records
# across the chain: 1,000,000 is above the 600,000 of a plain ICA msg, and the fee stays at the 0.005 ustrd/gas
# of the other Stride sets.
RESTORE_CHANNEL_GAS = 1_000_000
RESTORE_CHANNEL_FEES = "5000ustrd"


@dataclass(frozen=True)
class DrainRetry:
    """A full-drain batch the host rejected, resubmitted as a validators file without the validators that made it
    fail. On 2026-10-12 Injective rejected batch 1483 (32 validators, 18,703.52 INJ): three validators the 19:00
    epoch had already drained in full held ~1,000 wei each with Stride's record a few wei above the host, and the
    1-wei rounding trim cannot cover that on a sub-token balance, so the host refused their MsgUndelegate and the
    whole batch with it. Batch 1484 succeeded. The retry drains the other 29 and leaves the dust where it is."""

    chain_id: str
    label: str
    why: str
    excluded: tuple[tuple[str, str], ...]  # (moniker, valoper) left out, with the dust each holds
    validators: tuple[str, ...]  # the valopers to drain, offsets 0


DRAIN_RETRIES = (
    DrainRetry(
        chain_id="injective-1",
        label="batch 1483 retry",
        why="batch 1483 (packet 1483 on channel-816) was rejected by Injective on 2026-10-12 because three dust validators "
        "have records a few wei above the host; batch 1484 succeeded",
        excluded=(
            ("ubik", "injvaloper1h4t5dqenq86znza06thc64pzyxtme7zltyeg8k"),
            ("stakewithus", "injvaloper1v2xylfeyry808ltt0hwxggzt0fns6uwh74mku3"),
            ("allnodes", "injvaloper14w4qu60azpz5zkw5yrcvmaxqh6xg5sny7qm6rm"),
        ),
        validators=(
        "injvaloper1acgud5qpn3frwzjrayqcdsdr9vkl3p6hrz34ts",
        "injvaloper125fkz3mq6qxxpkmphdl3ep92t0d3y9695mhclt",
        "injvaloper1ffsdugrhfzdyxltjve8v68n6aazyc6p97uhfn0",
        "injvaloper134dct56cq5v7uerxcy2cn4m06mqf4dxrlgpp24",
        "injvaloper1g4d6dmvnpg7w7yugy6kplndp7jpfmf3krtschp",
        "injvaloper1nq37nq79w2j76xj8qhjzcn6vh0wlx0qk2r7zm6",
        "injvaloper1esud09zs5754g5nlkmrgxsfdj276xm64cgmd3w",
        "injvaloper1xwsnq88kc8wcrp34qenxf3dvhl5n02yj93u755",
        "injvaloper1f2kdg34689x93cvw2y59z7y46dvz2fk8lhddfz",
        "injvaloper1fqrdtx7pyps6eytn3356j9cs4f8zl0eevlt3rt",
        "injvaloper10xhy8xurfwts9ckjkq0ga92mrjz9txyygymqzp",
        "injvaloper155yk4wfn0xqye80exlsr6hu4qdfsvsgwg3jckk",
        "injvaloper1zpy3qf7us3m0pxpqkp72gzjv55t70huy33t47x",
        "injvaloper14yeq3lkajldaggj28hmq8xng9xux7x5g46hezv",
        "injvaloper13v0sulppc8pgtk4907p3vaqgq54vyl9yqmf7av",
        "injvaloper19a77dzm2lrxt2gehqca3nyzq077kq7qsgvmrp4",
        "injvaloper1vgu49xjgle9k84smvrwlyuc38u0cakuz7q3a5f",
        "injvaloper1nngrhnm65wm8pu7wkah6hfs3vpd0xw463ydd65",
        "injvaloper1yljq5pdnx84kkg30jfmz6ddu4eyp7twyp4z40f",
        "injvaloper10pe4avat38u38yzj5hvnw235uecfff6c73s8fn",
        "injvaloper1nxq05qdle8w5gm6fagda56euwt4rgvv4ktddcc",
        "injvaloper1mlsg82x0mnw88u2teg0kceautjktgcl08tqfs4",
        "injvaloper1845wspsvm3z95a2zycz3t49gn7celq48dlyscc",
        "injvaloper1vm2gflv53mzy9ak4f8cr9ml3xus980wywe9fx2",
        "injvaloper177zqwtsnuhax28w5xpmf33wsl7l2xxf6y6xshc",
        "injvaloper1da0shwz2mcup5rxkykquc9a7mh4s2hkeah9wuc",
        "injvaloper1rqqpyuka5dxulzjslnzjld2ltcw5095rr0jz07",
        "injvaloper1nm48eujr28u3htqrjumfwhytn63rmca2prtklt",
        "injvaloper16jkt2gcm8qfxzjj8j6pvyc045ga82pmthz58mp",
        ),
    ),
)


# Sets whose every tx has landed (2026-10-20): each card stays for the record as done, with nothing to copy.
COMPLETED_SET_IDS = frozenset({"slash-queries", "validator-weights", "live-test-undelegate", "full-drain"})


@dataclass(frozen=True)
class ChainTools:
    """The binary, chain and multisig a set's commands run against."""

    binary: str
    chain_id: str
    node: str
    multisig_address: str

    @property
    def keyring_note(self) -> str:
        # `tx sign --multisig <address>` resolves the address through the signer's own keyring, so everyone needs the key.
        return f"needs the {MULTISIG_KEY} multisig key in your keyring: {self.binary} keys show {self.multisig_address}"


STRIDE_TOOLS = ChainTools(binary="strided", chain_id=CHAIN_ID, node=NODE, multisig_address=MULTISIG_ADDRESS)
OSMOSIS_TOOLS = ChainTools(
    binary="osmosisd", chain_id=config.OSMOSIS_CHAIN_ID, node=OSMOSIS_NODE, multisig_address=OSMOSIS_VAULT_ADDRESS
)
KEYRING_NOTE = STRIDE_TOOLS.keyring_note

LIVE_TEST_GAS = 12_000_000
# Gas from the mainnet-export measurement (ops plan, drain-rest): the Hub is heaviest, osmosis-1 and ssc-1 next.
FULL_DRAIN_GAS_BY_CHAIN_ID = {"cosmoshub-4": 25_000_000, "osmosis-1": 15_000_000, "ssc-1": 15_000_000}
FULL_DRAIN_DEFAULT_GAS = 13_000_000
GAS_PRICE_USTRD = Decimal("0.005")


@dataclass(frozen=True)
class Signer:
    tag: str  # who they are in the UI
    key: str  # their member key in their own keyring


SAM = Signer(tag="Sam", key="FS5")
AIDAN = Signer(tag="Aidan", key="FA5")
RILEY = Signer(tag="Riley", key="FR5")
SIGNERS = (SAM, AIDAN, RILEY)
DEFAULT_SIGNERS = (SAM, AIDAN)  # any two of the three suffice
BROADCASTER = SAM  # collects the other signature (it arrives in ~/Downloads), multisigns and broadcasts
DOWNLOADS = "~/Downloads"


@dataclass(frozen=True)
class Command:
    tag: str  # who runs it: a signer's tag or "anyone"
    label: str
    text: str


@dataclass(frozen=True)
class MultisigTx:
    chain_id: str
    title: str
    ready: bool  # False with `reason` when the inputs are not known yet
    reason: str | None
    commands: list[Command]  # generate, sign for each signer (default ones, then the backup), multisign+broadcast
    files: list[str]  # the /tmp paths the commands share, for the "share these" note
    members: list[str] = dataclasses.field(default_factory=list)  # what a bundle holds ("zone pool"), shown as chips
    done: bool = False  # the tx has landed (every pool of the bundle exists): shown, not re-run
    blocked: bool = False  # waits on something outside our hands (config.BLOCKED_HOLDER_CHAINS)
    seeded: bool | None = None  # bundles: every pool's stToken denom is in the test wallet (created pools count)

    def payload(self) -> dict[str, Any]:
        return dataclasses.asdict(self)


@dataclass(frozen=True)
class TxSet:
    id: str
    step_id: str  # the ops step it belongs to
    title: str
    description: str  # one sentence: what the set does
    notes: list[str]  # short bullets behind a fold: order, gates, fees, gotchas
    txs: list[MultisigTx]  # in config.ZONES order, the txs of one zone contiguous and in the order they go out

    def payload(self) -> dict[str, Any]:
        return dataclasses.asdict(self)


# ---- public API


def tx_sets(
    validators_data: dict[str, Any] | None, funds_data: dict[str, Any] | None, pools_data: dict[str, Any] | None
) -> list[TxSet]:
    """Every set, from the snapshots' `data` (None before a tab's first snapshot)."""
    zones_by_chain_id = {zone["chain_id"]: zone for zone in (validators_data or {}).get("zones", [])}
    snapshot_reason = None if validators_data is not None else NO_SNAPSHOT_REASON

    live_test = TxSet(
        id="live-test-undelegate",
        step_id="drain-live-test",
        title="Live test: undelegate one validator per zone",
        description="Drain one live-test validator per zone in full with MsgUndelegateFromValidators.",
        notes=[
            "Validators tab: its recorded delegation drops to zero (dust, if slashed); the Ops check turns green when the unbonding entry lands.",
            ONE_AT_A_TIME,
        ],
        txs=[
            _live_test_tx(zone=zone, snapshot_zone=zones_by_chain_id.get(zone.chain_id), snapshot_reason=snapshot_reason)
            for zone in config.ZONES
        ],
    )
    full_drain = TxSet(
        id="full-drain",
        step_id="drain-rest",
        title="Full drain: undelegate every remaining validator",
        description="Undelegate every validator left after the live test, per zone, with --all.",
        notes=[
            "Relay the acks before the timeouts and wait for every batch to ack.",
            "A batch the host rejects leaves its validators' records unchanged (flags cleared); the retry card under that zone resubmits the batch as a validators file without the validators that made it fail.",
            ONE_AT_A_TIME,
        ],
        txs=[
            tx
            for zone in config.ZONES
            for tx in (
                _full_drain_tx(zone=zone, snapshot_zone=zones_by_chain_id.get(zone.chain_id), snapshot_reason=snapshot_reason),
                *(_drain_retry_tx(zone=zone, retry=retry) for retry in DRAIN_RETRIES if retry.chain_id == zone.chain_id),
            )
        ],
    )
    sets = [
        _pool_creation_set(pools_data=pools_data),
        _slash_queries_set(),
        _validator_weights_set(),
        _restore_channels_set(),
        live_test,
        full_drain,
        _ica_transfers_set(funds_data=funds_data),
        _staketia_claim_balance_set(),
        _pool_funding_set(pools_data=pools_data),
    ]
    return [_completed(tx_set=tx_set) if tx_set.id in COMPLETED_SET_IDS else tx_set for tx_set in sets]


# ---- sets


def _completed(tx_set: TxSet) -> TxSet:
    """Every tx of the set as a done card: the same shape `_stride_tx` gives a single done tx."""
    done_txs = [
        dataclasses.replace(tx, ready=False, reason=None, commands=[], files=[], done=True) for tx in tx_set.txs
    ]
    return dataclasses.replace(tx_set, txs=done_txs)


@dataclass(frozen=True)
class CreationCandidate:
    """A planned pool and why it is not in a bundle yet (None: it is); `plan` is None when the zone's snapshot is unreadable."""

    zone: config.ZoneConfig
    plan: dict[str, Any] | None
    reason: str | None

    @property
    def label(self) -> str:
        return f"{self.zone.chain_id} {self.plan['alloyed_subdenom']}" if self.plan else f"{self.zone.chain_id} pool creation"


def _pool_creation_set(pools_data: dict[str, Any] | None) -> TxSet:
    pools_zones = _zones_by_chain_id(data=pools_data)
    pools_to_create = _pools_to_create(pools_zones=pools_zones)
    fee_reason = _creation_fee_reason(pools_zones=pools_zones, pools_to_create=pools_to_create)
    candidates = _creation_candidates(pools_zones=pools_zones, pools_data=pools_data)
    # Every pool with an instantiate message has a fixed place in a bundle from day one, so the three txs can be
    # scaffolded before the seeding is done; a pool with no message (an unreadable zone, an unresolved route) is
    # listed after the bundles instead.
    bundleable = [candidate for candidate in candidates if candidate.plan and not candidate.plan["error"]]
    unbundleable = [candidate for candidate in candidates if not (candidate.plan and not candidate.plan["error"])]
    # Blocked pools (a holder chain we cannot seed from yet) go last, in bundles of their own, so the open bundles can
    # be signed without them.
    open_pools = [candidate for candidate in bundleable if not candidate.plan.get("blocked")]
    blocked_pools = [candidate for candidate in bundleable if candidate.plan.get("blocked")]
    bundles = _open_bundles(open_pools=open_pools, pool_ids=_pool_ids(pools_zones=pools_zones)) + _chunks(blocked_pools)
    fee = _creation_fee_text(pools_zones=pools_zones)
    return TxSet(
        id="pool-creation",
        step_id="vote-pools-create",
        title="Pool creation: instantiate every planned transmuter pool",
        description=f"Create every planned pool on Osmosis from the vault, {CREATE_POOL_BUNDLE_SIZE} pools per multisig tx.",
        notes=[
            "A bundle is ready once the test wallet holds every stToken denom in it and the vault can pay its fees.",
            f"Creation fee: {fee} per pool, paid by the vault; {pools_to_create} pools still to create.",
            "Generate one create-pool per message, then the jq line merges them into the bundle's unsigned tx: that is what gets signed and broadcast.",
            f"Gas {CREATE_POOL_MSG_GAS:,} per message at {OSMOSIS_GAS_PRICE}/gas, about 3x the base fee: a cheaper tx is rejected at broadcast and must be re-signed.",
            "Factors: 1e18 scale for six-decimal zones; 1e6 for haqq, dYdX and Injective (1e18 overflows the contract on their supply).",
            "The rate is read when this page renders: generate and sign a bundle the same day.",
            "A failing message reverts the whole bundle; pool ids follow message order. Confirm on the Pools tab before the next bundle.",
            ONE_AT_A_TIME,
        ],
        txs=[
            *(
                _bundle_tx(index=index, total=len(bundles), bundle=bundle, fee_reason=fee_reason)
                for index, bundle in enumerate(bundles, start=1)
            ),
            *([_waiting_tx(waiting=unbundleable)] if unbundleable else []),
        ],
    )


def _creation_candidates(pools_zones: dict[str, dict[str, Any]], pools_data: dict[str, Any] | None) -> list[CreationCandidate]:
    """Every planned pool still to create, zone by zone in config order; a zone whose snapshot is unreadable is one
    candidate with that reason."""
    candidates: list[CreationCandidate] = []
    for zone in config.ZONES:
        pools_zone = pools_zones.get(zone.chain_id)
        snapshot_reason = _snapshot_reason(zone_entry=pools_zone, data=pools_data, source="Pools")
        planned = [] if snapshot_reason else pools_zone.get("planned") or []
        if not planned:
            reason = snapshot_reason or "no planned pool in the Pools snapshot"
            candidates.append(CreationCandidate(zone=zone, plan=None, reason=reason))
            continue
        # Created pools stay in their bundle (membership and numbering never shift); the bundle shows as done.
        candidates.extend(CreationCandidate(zone=zone, plan=plan, reason=_creation_reason(plan=plan)) for plan in planned)
    return candidates


def _open_bundles(open_pools: list[CreationCandidate], pool_ids: dict[str, int]) -> list[list[CreationCandidate]]:
    """Created pools first, in creation order (pool id) and in the bundles that made them
    (config.CREATED_BUNDLE_SIZES); then one bundle per zone group of config.CREATE_POOL_BUNDLE_ZONES; then whatever
    zone the groups do not name, in zone order."""
    created = sorted(
        (candidate for candidate in open_pools if candidate.plan["live_contract"]),
        key=lambda candidate: pool_ids.get(candidate.plan["live_contract"], 0),
    )
    remaining = [candidate for candidate in open_pools if not candidate.plan["live_contract"]]
    grouped = [
        [candidate for zone in group for candidate in remaining if candidate.zone.chain_id == zone]
        for group in config.CREATE_POOL_BUNDLE_ZONES
    ]
    named = {zone for group in config.CREATE_POOL_BUNDLE_ZONES for zone in group}
    unnamed = [candidate for candidate in remaining if candidate.zone.chain_id not in named]
    return _sized_chunks(created, sizes=config.CREATED_BUNDLE_SIZES) + [bundle for bundle in grouped if bundle] + _chunks(unnamed)


def _pool_ids(pools_zones: dict[str, dict[str, Any]]) -> dict[str, int]:
    """Contract -> pool id for every live pool the Pools snapshot holds (creation order)."""
    return {
        pool["contract"]: int(pool["pool_id"])
        for zone in pools_zones.values()
        if not zone.get("error")
        for pool in zone.get("pools") or []
        if pool.get("pool_id")
    }


def _sized_chunks(candidates: list[CreationCandidate], sizes: tuple[int, ...]) -> list[list[CreationCandidate]]:
    """The given sizes first, then the cap at a time."""
    chunks: list[list[CreationCandidate]] = []
    start = 0
    for size in sizes:
        if start >= len(candidates):
            return chunks
        chunks.append(candidates[start : start + size])
        start += size
    return chunks + _chunks(candidates[start:])


def _chunks(candidates: list[CreationCandidate]) -> list[list[CreationCandidate]]:
    return [candidates[start : start + CREATE_POOL_BUNDLE_SIZE] for start in range(0, len(candidates), CREATE_POOL_BUNDLE_SIZE)]


def _bundle_tx(index: int, total: int, bundle: list[CreationCandidate], fee_reason: str | None) -> MultisigTx:
    stem = f"{WORKDIR}/create-b{index}"
    to_create = [candidate for candidate in bundle if not candidate.plan["live_contract"]]
    members = [candidate.label for candidate in bundle]
    if not to_create:
        title = f"bundle {index} of {total} · {len(bundle)} pools · done"
        return MultisigTx(chain_id=CREATION_GROUP, title=title, ready=False, reason=None, commands=[], files=[], members=members, done=True, seeded=True)

    gas = CREATE_POOL_TX_OVERHEAD_GAS + CREATE_POOL_MSG_GAS * len(to_create)
    generate_lines = [
        f"osmosisd tx cosmwasmpool create-pool {CREATE_POOL_CODE_ID} "
        f"'{json.dumps(candidate.plan['instantiate_msg'], separators=(',', ':'))}' --from {OSMOSIS_VAULT_ADDRESS} "
        f"--generate-only --chain-id {OSMOSIS_TOOLS.chain_id} --node {OSMOSIS_NODE} --gas {gas} "
        f"--gas-prices {OSMOSIS_GAS_PRICE} > {_bundle_part_stem(stem=stem, position=position, candidate=candidate)}.unsigned.json"
        for position, candidate in enumerate(to_create, start=1)
    ]
    # The glob sorts the zero-padded parts into message order; the first file's auth_info (fee and gas) is kept.
    merge_line = (
        "jq -s '(map(.body.messages) | add) as $msgs | .[0] | .body.messages = $msgs | .signatures = []' "
        f"{stem}-*.unsigned.json > {stem}.unsigned.json"
    )
    blocked = bundle[0].plan.get("blocked")
    created = f" · {len(bundle) - len(to_create)} of {len(bundle)} created" if len(to_create) < len(bundle) else ""
    title = f"bundle {index} of {total} · {len(bundle)} pools{' (blocked)' if blocked else ''}{created}"
    reason = f"blocked: {blocked}" if blocked else fee_reason or _bundle_reason(bundle=to_create)
    return MultisigTx(
        chain_id=CREATION_GROUP,
        title=title,
        ready=reason is None,
        reason=reason,
        blocked=bool(blocked),
        seeded=_bundle_seeded(to_create=to_create),
        commands=_commands(
            generate="\n".join([f"mkdir -p {WORKDIR}", *generate_lines, merge_line]),
            file_stem=stem,
            tools=OSMOSIS_TOOLS,
            generate_label=f"Write each pool's unsigned tx and merge them into one ({len(to_create)} messages, --gas {gas})",
        ),
        files=_shared_files(file_stem=stem),
        members=members,
    )


def _bundle_seeded(to_create: list[CreationCandidate]) -> bool | None:
    """Tri-state over the pools still to create: False if any denom is missing from the test wallet, None if any is
    unknown (and none missing), else True."""
    flags = [candidate.plan["seeded"] for candidate in to_create]
    if any(flag is False for flag in flags):
        return False
    return None if any(flag is None for flag in flags) else True


def _bundle_reason(bundle: list[CreationCandidate]) -> str | None:
    """What the bundle waits for: its members that the test wallet does not hold yet (or whose balance is unknown)."""
    waiting = [candidate for candidate in bundle if candidate.reason]
    if not waiting:
        return None
    shown = ", ".join(candidate.label for candidate in waiting[:WAITING_SHOWN])
    more = f" and {len(waiting) - WAITING_SHOWN} more" if len(waiting) > WAITING_SHOWN else ""
    return f"{len(waiting)} of {len(bundle)} pools waiting on the test wallet: {shown}{more}"


def _bundle_part_stem(stem: str, position: int, candidate: CreationCandidate) -> str:
    return f"{stem}-{position:02d}-{candidate.plan['alloyed_subdenom'].lower().replace('.', '-')}"


def _waiting_tx(waiting: list[CreationCandidate]) -> MultisigTx:
    """The pools no bundle can hold (no instantiate message: an unreadable zone or an unresolved route), each with why."""
    shown = waiting[:WAITING_SHOWN]
    more = f"; … and {len(waiting) - WAITING_SHOWN} more" if len(waiting) > WAITING_SHOWN else ""
    reason = "; ".join(f"{candidate.label}: {candidate.reason}" for candidate in shown) + more
    title = f"not in a bundle: {len(waiting)} pools without a message yet"
    return MultisigTx(chain_id=CREATION_GROUP, title=title, ready=False, reason=reason, commands=[], files=[])


def _creation_fee_text(pools_zones: dict[str, dict[str, Any]]) -> str:
    """The live creation fee as any zone reports it, for the set description."""
    fees = [zone.get("creation_fee") for zone in pools_zones.values()]
    fee = next((fee for fee in fees if fee), None)
    return f"{_whole_units(amount=int(fee['amount']))} {_denom_label(denom=fee['denom'])}" if fee else "not known yet"


def _ica_transfers_set(funds_data: dict[str, Any] | None) -> TxSet:
    funds_zones = _zones_by_chain_id(data=funds_data)
    return TxSet(
        id="ica-transfers",
        step_id="transfers",
        title="ICA transfers: send every ICA's balance to the Osmosis vault",
        description="Send each zone's FEE, WITHDRAWAL, DELEGATION and REDEMPTION ICA balance to the Osmosis vault, in that order.",
        notes=[
            "Per ICA: a one-token test tx first, then the rest tx for the live balance once the test has settled.",
            "The rest amount comes from the Funds snapshot and stays not ready while a transfer from that ICA is in flight.",
            "A foreign denom an ICA holds (dYdX's USDC, spam tokens) is its own tx: skip any not worth a signing round.",
            "24h timeout: a refund means resubmit.",
            ONE_AT_A_TIME,
        ],
        txs=[
            tx
            for zone in config.ZONES
            for tx in _ica_transfer_txs(zone=zone, funds_zone=funds_zones.get(zone.chain_id), funds_data=funds_data)
        ],
    )


def _slash_queries_set() -> TxSet:
    return TxSet(
        id="slash-queries",
        step_id="day0-refresh",
        title="Slash queries: update-delegation for the validators whose rate moved",
        description="Re-query the shares-to-tokens rate of each listed validator with update-delegation and apply its slash.",
        notes=[
            "Validators tab: the slash-query pill goes on when the tx lands and off when the callback arrives, with the diff gone.",
            "Do not resubmit a validator while its slash-query pill is on: the follow-up delegation query is skipped until the callback clears it.",
            f"Another validator, or calibrate-delegation where the rate is unchanged but the delta is off: {_slash_query_template()}, then the same sign and broadcast steps.",
            ONE_AT_A_TIME,
        ],
        txs=[_slash_query_tx(query=query) for query in SLASH_QUERIES],
    )


def _validator_weights_set() -> TxSet:
    return TxSet(
        id="validator-weights",
        step_id="day0-epoch-submits",
        title="Validator weights: slack for an epoch unbonding that failed on exact capacity",
        description="Give a zone's epoch unbonding slack by setting a small weight on an empty validator with change-validator-weights.",
        notes=[
            "The generate step writes the weights file next to the unsigned tx; the Validators tab shows the new weight once it lands.",
            "Only the unbond arithmetic changes: v35 has no rebalancing, so nothing is delegated to the validator.",
            "The zone's queued records are retried at its next unbonding epoch (unbonding period / 7 + 1 days), not at the next day epoch.",
            f"Another zone or validator: {_validator_weights_template()}, then the same sign and broadcast steps.",
            ONE_AT_A_TIME,
        ],
        txs=[_validator_weight_tx(change=change) for change in VALIDATOR_WEIGHT_CHANGES],
    )


def _restore_channels_set() -> TxSet:
    return TxSet(
        id="restore-channels",
        step_id="dydx-restore-delegation",
        title="Restore channels: reopen a closed ICA channel",
        description="Reopen a zone's closed ICA channel with restore-interchain-account, which also resets the zone's in-flight records.",
        notes=[
            "restore-interchain-account is permissionless (anyone may send it); it is built as a multisig tx so the operators' keys carry it.",
            "After the broadcast the new channel sits in INIT until a relayer lands TRY on the host, ACK on Stride and CONFIRM on the host; check the Channels tab for it.",
            "Before broadcasting, check the zone's relayer has fees on the host chain: a relayer that fails ChanOpenTry for insufficient funds gives up after 5 retries.",
            "A Go relayer (rly) only acts on handshakes it sees live: after a restart it never retries an INIT channel, so finish a stale one with hermes (`hermes --config <temp> tx chan-open-try --dst-chain <host> --src-chain stride-1 --dst-connection <host connection> --dst-port icahost --src-port icacontroller-<owner> --src-channel <INIT channel>`); the relayer then does ACK and CONFIRM itself.",
            "Each extra restore leaves another orphan INIT channel, so do not resubmit while one is still handshaking.",
            ONE_AT_A_TIME,
        ],
        txs=[_channel_restore_tx(restore=restore) for restore in CHANNEL_RESTORES],
    )


def _staketia_claim_balance_set() -> TxSet:
    zone = config.ZONES_BY_CHAIN_ID[STAKETIA_CHAIN_ID]
    return TxSet(
        id="staketia-claim-balance",
        step_id="tia-claim-balance",
        title="Staketia claim balance: move the claim address's TIA to the celestia delegation ICA",
        description="Move the staketia claim address's TIA to the celestia delegation ICA: 1 TIA as the test, then amount 0 for the rest.",
        notes=["The TIA leaves Celestia with the DELEGATION transfer.", ONE_AT_A_TIME],
        txs=[
            _staketia_tx(zone=zone, label="test", amount=STAKETIA_TEST_AMOUNT, title_amount="1 TIA"),
            _staketia_tx(zone=zone, label="rest", amount=0, title_amount="the whole remainder (amount 0)"),
        ],
    )


def _pool_funding_set(pools_data: dict[str, Any] | None) -> TxSet:
    pools_zones = _zones_by_chain_id(data=pools_data)
    return TxSet(
        id="pool-funding",
        step_id="join-pools",
        title="Pool funding: join and close every transmuter pool",
        description="Fund each pool with a test join, then the rest of its allocation, then mark the native token corrupted.",
        notes=[
            "Per zone: every pool's test join first, then verify on the Pools tab (native one token, vault shares one token, the test wallet's shares present, rate exact).",
            "Then per pool: the rest join and the mark back to back; the mark closes the window in which an outsider could join.",
            "The mark is the moderator's call: the first tx of each zone makes the vault the moderator of its pools, the last hands the role back to osmo1ugrn8…hjdpgy once every pool is marked.",
            "The rest join is ready once the Pools snapshot shows the test join as vault shares.",
            "The vault needs about 0.15 OSMO per tx, three txs per pool.",
            ONE_AT_A_TIME,
        ],
        txs=[
            tx
            for zone in config.ZONES
            for tx in _pool_txs(zone=zone, pools_zone=pools_zones.get(zone.chain_id), pools_data=pools_data)
        ],
    )


# ---- txs


def _live_test_tx(zone: config.ZoneConfig, snapshot_zone: dict[str, Any] | None, snapshot_reason: str | None) -> MultisigTx:
    pick = snapshot_zone.get("live_test_pick") if snapshot_zone else None
    reason = _live_test_reason(snapshot_zone=snapshot_zone, snapshot_reason=snapshot_reason, has_pick=pick is not None)

    valoper = pick["address"] if pick else PLACEHOLDER_VALOPER
    file_stem = f"{WORKDIR}/live-test-{zone.chain_id}"
    validators_file = f"{file_stem}.json"
    generate = (
        f"mkdir -p {WORKDIR}\n"
        f"echo '[{{\"address\": \"{valoper}\", \"offset\": \"0\"}}]' > {validators_file}\n"
        f"{_generate_line(arguments=f'{zone.chain_id} {validators_file}', gas=LIVE_TEST_GAS, file_stem=file_stem)}"
    )
    return MultisigTx(
        chain_id=zone.chain_id,
        title=f"{zone.chain_id} · live test: {_pick_description(pick=pick, zone=zone)}",
        ready=reason is None,
        reason=reason,
        commands=_commands(generate=generate, file_stem=file_stem),
        files=_shared_files(file_stem=file_stem),
    )


def _live_test_reason(snapshot_zone: dict[str, Any] | None, snapshot_reason: str | None, has_pick: bool) -> str | None:
    if snapshot_reason:
        return snapshot_reason

    zone_error = _zone_error(snapshot_zone=snapshot_zone)
    if zone_error:
        return zone_error

    if has_pick:
        return None
    return snapshot_zone["live_test_reason"]


def _full_drain_tx(zone: config.ZoneConfig, snapshot_zone: dict[str, Any] | None, snapshot_reason: str | None) -> MultisigTx:
    reason = snapshot_reason or _zone_error(snapshot_zone=snapshot_zone)

    gas = FULL_DRAIN_GAS_BY_CHAIN_ID.get(zone.chain_id, FULL_DRAIN_DEFAULT_GAS)
    file_stem = f"{WORKDIR}/full-drain-{zone.chain_id}"
    generate = (
        f"mkdir -p {WORKDIR}\n"
        f"{_generate_line(arguments=f'{zone.chain_id} --all', gas=gas, file_stem=file_stem)}"
    )
    return MultisigTx(
        chain_id=zone.chain_id,
        title=f"{zone.chain_id} · full drain, every remaining validator",
        ready=reason is None,
        reason=reason,
        commands=_commands(generate=generate, file_stem=file_stem),
        files=_shared_files(file_stem=file_stem),
    )


def _drain_retry_tx(zone: config.ZoneConfig, retry: DrainRetry) -> MultisigTx:
    """The rejected batch again, as a validators file (offsets 0) that leaves out the validators that made it fail."""
    gas = FULL_DRAIN_GAS_BY_CHAIN_ID.get(zone.chain_id, FULL_DRAIN_DEFAULT_GAS)
    file_stem = f"{WORKDIR}/full-drain-retry-{zone.chain_id}"
    validators_file = f"{file_stem}.json"
    validators_json = json.dumps([{"address": valoper, "offset": "0"} for valoper in retry.validators])
    generate = (
        f"mkdir -p {WORKDIR}\n"
        f"cat > {validators_file} <<'EOF'\n{validators_json}\nEOF\n"
        f"{_generate_line(arguments=f'{zone.chain_id} {validators_file}', gas=gas, file_stem=file_stem)}"
    )
    excluded = ", ".join(f"{moniker} ({valoper[-6:]})" for moniker, valoper in retry.excluded)
    return MultisigTx(
        chain_id=zone.chain_id,
        title=f"{zone.chain_id} · full drain {retry.label}: {len(retry.validators)} validators, without {excluded} — {retry.why}",
        ready=True,
        reason=None,
        commands=_commands(generate=generate, file_stem=file_stem),
        files=_shared_files(file_stem=file_stem),
    )


def _ica_transfer_txs(
    zone: config.ZoneConfig, funds_zone: dict[str, Any] | None, funds_data: dict[str, Any] | None
) -> list[MultisigTx]:
    """Per ICA a test and a rest tx, then one tx per foreign denom it holds; placeholders until the Funds snapshot."""
    snapshot_reason = _snapshot_reason(zone_entry=funds_zone, data=funds_data, source="Funds")
    host_denom = PLACEHOLDER_DENOM if snapshot_reason else funds_zone["host_denom"]
    transfers = [] if snapshot_reason else funds_zone.get("transfers") or []
    test_amount = 10**zone.decimals

    txs: list[MultisigTx] = []
    for ica in ICA_ORDER:
        balances = {} if snapshot_reason else funds_zone["ica_balances"].get(ica, {})
        host_balance = None if snapshot_reason else int(balances.get(host_denom, 0))
        reason = snapshot_reason or _nothing_to_send_reason(ica=ica, balance=host_balance, test_amount=test_amount, denom=host_denom)
        rest_reason = reason or _in_flight_reason(ica=ica, transfers=transfers)
        balance_text = PLACEHOLDER_BALANCE if host_balance is None else str(host_balance)

        txs.append(
            _ica_transfer_tx(
                zone=zone, ica=ica, label="test", amount=str(test_amount), denom=host_denom, reason=reason,
                title=f"{zone.chain_id} · {ica} ICA · test: {test_amount}{host_denom}",
            )
        )
        txs.append(
            _ica_transfer_tx(
                zone=zone, ica=ica, label="rest", amount=balance_text, denom=host_denom, reason=rest_reason,
                title=f"{zone.chain_id} · {ica} ICA · rest: {balance_text}{host_denom} (live balance, copy after the test has settled)",
            )
        )
        txs.extend(
            _ica_transfer_tx(
                zone=zone, ica=ica, label=f"denom-{hashlib.sha256(denom.encode()).hexdigest()[:8]}", amount=amount, denom=denom, reason=snapshot_reason,
                title=f"{zone.chain_id} · {ica} ICA · foreign denom: {amount}{denom} (full balance)",
            )
            for denom, amount in _foreign_balances(balances=balances, host_denom=host_denom)
        )
    return txs


def _nothing_to_send_reason(ica: str, balance: int | None, test_amount: int, denom: str) -> str | None:
    if balance is not None and balance < test_amount:
        return f"the {ica} ICA holds {balance}{denom}, below the {test_amount}{denom} test amount: nothing to send"
    return None


def _in_flight_reason(ica: str, transfers: list[dict[str, Any]]) -> str | None:
    """The rest tx waits while a transfer from the ICA is in flight: the snapshot's balance predates its landing."""
    in_flight = any(
        transfer["ica"] == ica and transfer["status"] == funds.TransferStatus.IN_FLIGHT for transfer in transfers
    )
    if in_flight:
        return f"a transfer from the {ica} ICA is in flight: copy once it settles and the Funds snapshot refreshes"
    return None


def _foreign_balances(balances: dict[str, str], host_denom: str) -> list[tuple[str, str]]:
    return [(denom, amount) for denom, amount in sorted(balances.items()) if denom != host_denom and int(amount) > 0]


def _ica_transfer_tx(
    zone: config.ZoneConfig, ica: str, label: str, amount: str, denom: str, reason: str | None, title: str
) -> MultisigTx:
    file_stem = f"{WORKDIR}/transfer-{zone.chain_id}-{ica.lower()}-{label}"
    generate_line = _stride_generate_line(
        subcommand=f"transfer-from-ica {zone.chain_id} {ica} {amount}{denom}", file_stem=file_stem
    )
    return _stride_tx(zone=zone, title=title, reason=reason, generate_line=generate_line, file_stem=file_stem)


def _staketia_tx(zone: config.ZoneConfig, label: str, amount: int, title_amount: str) -> MultisigTx:
    file_stem = f"{WORKDIR}/staketia-claim-balance-{label}"
    generate_line = _stride_generate_line(subcommand=f"transfer-staketia-claim-balance {amount}", file_stem=file_stem)
    return _stride_tx(
        zone=zone,
        title=f"{zone.chain_id} · staketia claim balance · {label}: {title_amount}",
        reason=None,
        generate_line=generate_line,
        file_stem=file_stem,
    )


def _slash_query_template() -> str:
    """The generate line for a validator the set does not list, in the step's own placeholders."""
    return _stride_generate_line(
        subcommand="update-delegation|calibrate-delegation <CHAIN_ID> <VALOPER>",
        file_stem=f"{WORKDIR}/<CHAIN_ID>-<VALOPER>",
        gas=SLASH_QUERY_GAS,
        fees=SLASH_QUERY_FEES,
    )


def _slash_query_tx(query: SlashQuery) -> MultisigTx:
    zone = config.ZONES_BY_CHAIN_ID[query.chain_id]
    file_stem = f"{WORKDIR}/update-delegation-{zone.chain_id}-{query.valoper[-6:]}"
    generate_line = _stride_generate_line(
        subcommand=f"update-delegation {zone.chain_id} {query.valoper}",
        file_stem=file_stem,
        gas=SLASH_QUERY_GAS,
        fees=SLASH_QUERY_FEES,
    )
    return _stride_tx(
        zone=zone,
        title=f"{zone.chain_id} · update-delegation: {query.moniker} ({query.valoper})",
        reason=None,
        generate_line=generate_line,
        file_stem=file_stem,
        done=query.done,
    )


def _validator_weights_template() -> str:
    """The generate line for a change the set does not list, in the step's own placeholders."""
    return _stride_generate_line(
        subcommand="change-validator-weights <CHAIN_ID> <WEIGHTS_FILE>",
        file_stem=f"{WORKDIR}/validator-weights-<CHAIN_ID>",
        gas=VALIDATOR_WEIGHTS_GAS,
        fees=VALIDATOR_WEIGHTS_FEES,
    )


def _validator_weight_tx(change: ValidatorWeightChange) -> MultisigTx:
    zone = config.ZONES_BY_CHAIN_ID[change.chain_id]
    file_stem = f"{WORKDIR}/validator-weights-{zone.chain_id}-{change.valoper[-6:]}"
    weights_file = f"{file_stem}.weights.json"
    weights_json = json.dumps({"validator_weights": [{"address": change.valoper, "weight": change.weight}]})
    generate_line = (
        f"cat > {weights_file} <<'EOF'\n{weights_json}\nEOF\n"
        + _stride_generate_line(
            subcommand=f"change-validator-weights {zone.chain_id} {weights_file}",
            file_stem=file_stem,
            gas=VALIDATOR_WEIGHTS_GAS,
            fees=VALIDATOR_WEIGHTS_FEES,
        )
    )
    return _stride_tx(
        zone=zone,
        title=f"{zone.chain_id} · change-validator-weights: {change.moniker} -> {change.weight} ({change.why})",
        reason=None,
        generate_line=generate_line,
        file_stem=file_stem,
    )


def _channel_restore_tx(restore: ChannelRestore) -> MultisigTx:
    zone = config.ZONES_BY_CHAIN_ID[restore.chain_id]
    file_stem = f"{WORKDIR}/restore-{zone.chain_id}-{restore.account_type.lower()}"
    generate_line = _stride_generate_line(
        subcommand=f"restore-interchain-account {zone.chain_id} {restore.connection_id} {zone.chain_id}.{restore.account_type}",
        file_stem=file_stem,
        gas=RESTORE_CHANNEL_GAS,
        fees=RESTORE_CHANNEL_FEES,
    )
    return _stride_tx(
        zone=zone,
        title=f"{zone.chain_id} · restore-interchain-account: {restore.account_type} ({restore.closed_channel} is closed)",
        reason=None,
        generate_line=generate_line,
        file_stem=file_stem,
    )


def _stride_tx(
    zone: config.ZoneConfig, title: str, reason: str | None, generate_line: str, file_stem: str, done: bool = False
) -> MultisigTx:
    # A done tx keeps its card for the record but nothing to copy, the same shape as a done pool bundle.
    if done:
        return MultisigTx(chain_id=zone.chain_id, title=title, ready=False, reason=None, commands=[], files=[], done=True)
    return MultisigTx(
        chain_id=zone.chain_id,
        title=title,
        ready=reason is None,
        reason=reason,
        commands=_commands(
            generate=f"mkdir -p {WORKDIR}\n{generate_line}", file_stem=file_stem, generate_label="Write the unsigned tx"
        ),
        files=_shared_files(file_stem=file_stem),
    )


def _stride_generate_line(subcommand: str, file_stem: str, gas: int = TRANSFER_GAS, fees: str = TRANSFER_FEES) -> str:
    return (
        f"strided tx stakeibc {subcommand} --from {MULTISIG_ADDRESS} --generate-only --chain-id {CHAIN_ID} "
        f"--node {NODE} --gas {gas} --fees {fees} > {file_stem}.unsigned.json"
    )


def _creation_reason(plan: dict[str, Any]) -> str | None:
    """Why the pool cannot go into a bundle yet, in the order the operator should resolve them."""
    if plan["error"]:
        return plan["error"]
    if plan["live_contract"]:
        return f"already created: {plan['live_contract']}"
    if plan["seeded"] is None:
        return f"the test wallet's balance of {plan['denom_on_osmosis']} is not known yet (see the Pools tab)"
    if not plan["seeded"]:
        return f"the test wallet {config.POOL_SEED_ADDRESS} does not hold {plan['denom_on_osmosis']}: send it there first"
    return None


def _pools_to_create(pools_zones: dict[str, dict[str, Any]]) -> int:
    """Planned pools without a live contract across every readable zone: they all draw on the one vault balance."""
    return sum(
        1
        for zone in pools_zones.values()
        if not zone.get("error")
        for plan in zone.get("planned") or []
        if not plan["live_contract"] and not plan.get("blocked")  # a blocked pool is not being created now
    )


def _creation_fee_reason(pools_zones: dict[str, dict[str, Any]], pools_to_create: int) -> str | None:
    """The vault pays every zone's fees from one balance, so the gate is the fee x all pools still to create."""
    readable = [zone for zone in pools_zones.values() if not zone.get("error")]
    fee = next((zone["creation_fee"] for zone in readable if zone.get("creation_fee")), None)
    balance = next((zone["vault_fee_balance"] for zone in readable if zone.get("vault_fee_balance") is not None), None)
    if fee is None or balance is None:
        return "the pool creation fee or the vault's balance of it is not known yet (see the Pools tab)"
    if int(balance) >= int(fee["amount"]) * pools_to_create:
        return None
    label = _denom_label(denom=fee["denom"])
    return (
        f"vault holds {_whole_units(amount=int(balance))} {label}, needs {_whole_units(amount=int(fee['amount']))} × "
        f"{pools_to_create} = {_whole_units(amount=int(fee['amount']) * pools_to_create)} {label}: top it up first"
    )


def _denom_label(denom: str) -> str:
    """`allUSDC` for a tokenfactory denom, the denom itself otherwise."""
    return denom.rsplit("/", 1)[-1] if denom.startswith("factory/") else denom


def _whole_units(amount: int) -> str:
    """Six-decimal base units as whole tokens (the fee denoms, allUSDC and uosmo, are six-decimal)."""
    whole, rest = divmod(amount, 10**6)
    return f"{whole:,}" if rest == 0 else f"{whole:,}.{rest:06d}".rstrip("0")


def _pool_txs(zone: config.ZoneConfig, pools_zone: dict[str, Any] | None, pools_data: dict[str, Any] | None) -> list[MultisigTx]:
    """The zone's test joins, then each pool's rest join and mark back to back; one not-ready tx when there is nothing."""
    snapshot_reason = _snapshot_reason(zone_entry=pools_zone, data=pools_data, source="Pools")
    pools = [] if snapshot_reason else _fundable_pools(pools_zone=pools_zone)
    if not pools:
        reason = snapshot_reason or "no pool to fund in the Pools snapshot"
        return [_empty_pool_tx(zone=zone, label="pool funding", reason=reason)]

    osmosis_denom = pools_zone["osmosis_denom"]
    test_amount = 10**zone.decimals
    return [
        _moderator_tx(zone=zone, pools=pools, to_vault=True),
        *(_join_tx(zone=zone, pool=pool, osmosis_denom=osmosis_denom, test_amount=test_amount, is_test=True) for pool in pools),
        *(
            tx
            for pool in pools
            for tx in (
                _join_tx(zone=zone, pool=pool, osmosis_denom=osmosis_denom, test_amount=test_amount, is_test=False),
                _mark_tx(zone=zone, pool=pool, osmosis_denom=osmosis_denom),
            )
        ),
        _moderator_tx(zone=zone, pools=pools, to_vault=False),
    ]


MODERATOR_MSG_GAS = 300_000  # assign_moderator is a storage write; the 09-25 pools took well under this


def _moderator_tx(zone: config.ZoneConfig, pools: list[dict[str, Any]], to_vault: bool) -> MultisigTx:
    """One tx, assign_moderator on every pool of the zone (the admin's call): the vault takes the role before the joins
    so it can mark the native token, and hands it back to config.POOL_MODERATOR once every pool is marked."""
    target = config.OSMOSIS_VAULT if to_vault else config.POOL_MODERATOR
    label = "moderator-to-vault" if to_vault else "moderator-back"
    stem = f"{WORKDIR}/pool-{zone.chain_id}-{label}"
    gas = CREATE_POOL_TX_OVERHEAD_GAS + MODERATOR_MSG_GAS * len(pools)
    message_json = json.dumps({"assign_moderator": {"address": target}}, separators=(",", ":"))
    generate_lines = [
        f"osmosisd tx wasm execute {pool['contract']} '{message_json}' --from {OSMOSIS_VAULT_ADDRESS} --generate-only "
        f"--chain-id {OSMOSIS_TOOLS.chain_id} --node {OSMOSIS_NODE} --gas {gas} --gas-prices {OSMOSIS_GAS_PRICE} "
        f"> {stem}-{position:02d}-{_pool_label(pool=pool)}.unsigned.json"
        for position, pool in enumerate(pools, start=1)
    ]
    merge_line = (
        "jq -s '(map(.body.messages) | add) as $msgs | .[0] | .body.messages = $msgs | .signatures = []' "
        f"{stem}-*.unsigned.json > {stem}.unsigned.json"
    )
    title = f"{zone.chain_id} · moderator → {'the vault, before the joins' if to_vault else 'osmo1ugrn8…hjdpgy, after the marks'} ({len(pools)} pools)"
    reason = _moderator_reason(pools=pools, to_vault=to_vault)
    return MultisigTx(
        chain_id=zone.chain_id,
        title=title,
        ready=reason is None,
        reason=reason,
        commands=_commands(
            generate="\n".join([f"mkdir -p {WORKDIR}", *generate_lines, merge_line]),
            file_stem=stem,
            tools=OSMOSIS_TOOLS,
            generate_label=f"Write each pool's unsigned tx and merge them into one ({len(pools)} messages, --gas {gas})",
        ),
        files=_shared_files(file_stem=stem),
    )


def _moderator_reason(pools: list[dict[str, Any]], to_vault: bool) -> str | None:
    moderators = [pool.get("moderator") for pool in pools]
    if any(moderator is None for moderator in moderators):
        return "a pool's moderator is not known yet (see the Pools tab)"
    if to_vault:
        if all(pool["native_marked"] for pool in pools):
            return "every pool is already marked: nothing left for the vault to do as moderator"
        if all(moderator == config.OSMOSIS_VAULT for moderator in moderators):
            return "the vault already moderates every pool of the zone"
        return None
    unmarked = [pool for pool in pools if not pool["native_marked"]]
    if unmarked:
        return f"{len(unmarked)} of {len(pools)} pools not marked yet: the vault still needs the moderator role"
    if all(moderator == config.POOL_MODERATOR for moderator in moderators):
        return "the moderator is already back on every pool of the zone"
    return None


def _fundable_pools(pools_zone: dict[str, Any]) -> list[dict[str, Any]]:
    fundable = [pool for pool in pools_zone["pools"] if pool["kind"] != POOL_KIND_UNRECOGNISED]
    return sorted(fundable, key=lambda pool: pool["kind"] != POOL_KIND_CANONICAL)  # stable: canonical first


def _empty_pool_tx(zone: config.ZoneConfig, label: str, reason: str) -> MultisigTx:
    return MultisigTx(
        chain_id=zone.chain_id, title=f"{zone.chain_id} · {label}", ready=False, reason=reason, commands=[], files=[]
    )


def _join_tx(zone: config.ZoneConfig, pool: dict[str, Any], osmosis_denom: str | None, test_amount: int, is_test: bool) -> MultisigTx:
    amount_text, reason = (
        _test_join_amount(pool=pool, test_amount=test_amount)
        if is_test
        else _rest_join_amount(pool=pool, test_amount=test_amount)
    )
    reason = _unknown_denom_reason(osmosis_denom=osmosis_denom) or reason
    label = "test" if is_test else "rest"
    denom = osmosis_denom or PLACEHOLDER_OSMOSIS_DENOM
    file_stem = _pool_file_stem(zone=zone, pool=pool, label=label)
    join_title = f"{'test join' if is_test else 'join rest'}, {amount_text}{denom}"
    return _osmosis_tx(
        zone=zone,
        title=f"{zone.chain_id} · {_pool_name(pool=pool)} · {join_title}",
        reason=reason,
        generate_line=_osmosis_generate_line(
            contract=pool["contract"], message={"join_pool": {}}, amount=f"{amount_text}{denom}", file_stem=file_stem
        ),
        file_stem=file_stem,
    )


def _test_join_amount(pool: dict[str, Any], test_amount: int) -> tuple[str, str | None]:
    """One whole token, or the whole allocation when that is smaller; nothing goes until the allocation is known."""
    if pool["allocation"] is None:
        return str(test_amount), ALLOCATION_UNKNOWN_REASON
    allocation = int(pool["allocation"])
    if allocation <= 0:
        return "0", ALLOCATION_EMPTY_REASON
    return str(min(allocation, test_amount)), None


def _rest_join_amount(pool: dict[str, Any], test_amount: int) -> tuple[str, str | None]:
    """What the allocation leaves after the shares the vault already holds, and why it cannot go yet: it waits until
    the Pools snapshot shows exactly the test join, so the amount is computed from what actually landed."""
    if pool["allocation"] is None:
        return PLACEHOLDER_REST_AMOUNT, ALLOCATION_UNKNOWN_REASON

    allocation = int(pool["allocation"])
    vault_shares = int(pool["vault_shares"])
    remaining = str(allocation - vault_shares)
    if allocation <= 0:
        return remaining, ALLOCATION_EMPTY_REASON
    if pool["funded_exactly"]:
        return remaining, "the pool is already funded exactly"
    if allocation <= test_amount:
        return remaining, f"the allocation ({allocation}) does not exceed the test join: nothing left to join"
    if vault_shares != test_amount:
        return remaining, (
            f"the Pools snapshot does not show the test join yet (vault shares {vault_shares}, expected {test_amount})"
        )
    return remaining, None


def _mark_reason(pool: dict[str, Any]) -> str | None:
    """A mark before the funding join would refuse that join (the mark forbids raising the native balance), so the
    mark waits for the pool to be funded exactly; a marked pool is done."""
    if pool.get("native_marked"):
        return "already marked"
    if pool.get("funded_exactly") is None:
        return "the pool's allocation is not known yet (see the Pools tab)"
    if not pool["funded_exactly"]:
        return "not funded exactly yet: the mark would refuse the funding join"
    if pool.get("moderator") not in (None, config.OSMOSIS_VAULT):
        return "the vault is not the moderator yet: send the zone's moderator → vault tx first"
    return None


def _mark_tx(zone: config.ZoneConfig, pool: dict[str, Any], osmosis_denom: str | None) -> MultisigTx:
    file_stem = _pool_file_stem(zone=zone, pool=pool, label="mark")
    reason = _unknown_denom_reason(osmosis_denom=osmosis_denom) or _mark_reason(pool=pool)
    message = {"mark_corrupted_assets": {"denoms": [osmosis_denom or PLACEHOLDER_OSMOSIS_DENOM]}}
    return _osmosis_tx(
        zone=zone,
        title=f"{zone.chain_id} · {_pool_name(pool=pool)} · mark corrupted",
        reason=reason,
        generate_line=_osmosis_generate_line(contract=pool["contract"], message=message, amount=None, file_stem=file_stem),
        file_stem=file_stem,
    )


def _unknown_denom_reason(osmosis_denom: str | None) -> str | None:
    return None if osmosis_denom else "the zone's native denom on Osmosis is not known yet (see the Pools tab)"


def _pool_name(pool: dict[str, Any]) -> str:
    return f"{pool['kind']} pool {_pool_label(pool=pool)}"


def _pool_label(pool: dict[str, Any]) -> str:
    return pool["pool_id"] or pool["contract"][-6:]


def _pool_file_stem(zone: config.ZoneConfig, pool: dict[str, Any], label: str) -> str:
    return f"{WORKDIR}/pool-{zone.chain_id}-{_pool_label(pool=pool)}-{label}"


def _osmosis_generate_line(contract: str, message: dict[str, Any], amount: str | None, file_stem: str) -> str:
    # Compact JSON, single-quoted for the shell: the transmuter's execute variants take no spaces.
    message_json = json.dumps(message, separators=(",", ":"))
    amount_flag = f" --amount {amount}" if amount else ""
    return (
        f"osmosisd tx wasm execute {contract} '{message_json}'{amount_flag} --from {OSMOSIS_VAULT_ADDRESS} "
        f"--generate-only --chain-id {OSMOSIS_TOOLS.chain_id} --node {OSMOSIS_NODE} --gas {POOL_GAS} --gas-prices {OSMOSIS_GAS_PRICE} "
        f"> {file_stem}.unsigned.json"
    )


def _osmosis_tx(zone: config.ZoneConfig, title: str, reason: str | None, generate_line: str, file_stem: str) -> MultisigTx:
    return MultisigTx(
        chain_id=zone.chain_id,
        title=title,
        ready=reason is None,
        reason=reason,
        commands=_commands(
            generate=f"mkdir -p {WORKDIR}\n{generate_line}",
            file_stem=file_stem,
            tools=OSMOSIS_TOOLS,
            generate_label="Write the unsigned tx",
        ),
        files=_shared_files(file_stem=file_stem),
    )


def _commands(
    generate: str,
    file_stem: str,
    tools: ChainTools = STRIDE_TOOLS,
    generate_label: str = "Write the validators file and the unsigned tx",
) -> list[Command]:
    unsigned = f"{file_stem}.unsigned.json"
    sign_commands = [
        Command(
            tag=signer.tag,
            label=_sign_label(signer=signer, keyring_note=tools.keyring_note),
            text=(
                f"{tools.binary} tx sign {unsigned} --multisig {tools.multisig_address} --from {signer.key} "
                f"--chain-id {tools.chain_id} --node {tools.node} \\\n  --output-document {_signature_file(file_stem=file_stem, signer=signer)}"
            ),
        )
        for signer in SIGNERS
    ]
    signature_files = " ".join(_signature_path(file_stem=file_stem, signer=signer) for signer in DEFAULT_SIGNERS)
    combine = Command(
        tag=BROADCASTER.tag,
        label="Combine and broadcast",
        text=(
            # --output-document, not a shell redirect: cobra prints the signed tx to stderr in binaries that do not
            # route cmd output to stdout (osmosisd), so `>` would leave an empty file.
            f"{tools.binary} tx multisign {unsigned} {MULTISIG_KEY} {signature_files} --chain-id {tools.chain_id} "
            f"--node {tools.node} --output-document {file_stem}.signed.json\n"
            f"{tools.binary} tx broadcast {file_stem}.signed.json --node {tools.node} --broadcast-mode sync"
        ),
    )
    return [Command(tag=ANYONE, label=generate_label, text=generate), *sign_commands, combine]


def _sign_label(signer: Signer, keyring_note: str = KEYRING_NOTE) -> str:
    if signer == SAM:
        return f"Sign (online: the multisig's account number and sequence are looked up; {keyring_note})"
    if signer in DEFAULT_SIGNERS:
        return f"Sign ({keyring_note})"
    return f"Backup signer (any two signatures suffice; {keyring_note})"


def _generate_line(arguments: str, gas: int, file_stem: str) -> str:
    fees = int(gas * GAS_PRICE_USTRD)
    return (
        f"strided tx stakeibc undelegate-from-validators {arguments} --from {MULTISIG_ADDRESS} --generate-only \\\n"
        f"  --chain-id {CHAIN_ID} --node {NODE} --gas {gas} --fees {fees}ustrd > {file_stem}.unsigned.json"
    )


def _signature_file(file_stem: str, signer: Signer) -> str:
    return f"{file_stem}.{signer.key}.json"


def _signature_path(file_stem: str, signer: Signer) -> str:
    """Where the broadcaster finds a signature: their own where they wrote it, the other signer's in their Downloads
    (it arrives over Slack)."""
    if signer == BROADCASTER:
        return _signature_file(file_stem=file_stem, signer=signer)
    return f"{DOWNLOADS}/{file_stem.rsplit('/', 1)[-1]}.{signer.key}.json"


def _shared_files(file_stem: str) -> list[str]:
    return [f"{file_stem}.unsigned.json", *(_signature_file(file_stem=file_stem, signer=signer) for signer in SIGNERS)]


# ---- helpers


def _zones_by_chain_id(data: dict[str, Any] | None) -> dict[str, dict[str, Any]]:
    return {zone["chain_id"]: zone for zone in (data or {}).get("zones", [])}


def _snapshot_reason(zone_entry: dict[str, Any] | None, data: dict[str, Any] | None, source: str) -> str | None:
    """Why a zone's txs cannot be built from a snapshot: not taken yet, zone missing, or the zone's own error."""
    if data is None:
        return f"waiting for the {source} snapshot"
    if zone_entry is None:
        return f"zone missing from the {source} snapshot"
    return zone_entry.get("error")


def _zone_error(snapshot_zone: dict[str, Any] | None) -> str | None:
    if snapshot_zone is None:
        return "zone missing from the Validators snapshot"
    return snapshot_zone.get("error")


def _pick_description(pick: dict[str, str] | None, zone: config.ZoneConfig) -> str:
    if pick is None:
        return PLACEHOLDER_VALOPER

    amount = int(pick["recorded"])
    shown = (
        f"{amount:,} u{zone.symbol.lower()}"
        if zone.decimals == 6
        else f"{Decimal(amount).scaleb(-zone.decimals).normalize():,f} {zone.symbol}"
    )
    return f"{pick['moniker']} ({_short_address(address=pick['address'])}), {shown}"


def _short_address(address: str) -> str:
    # bech32 data never contains "1", so the last one separates the prefix from the data.
    prefix_end = address.rfind("1") + 1
    return f"{address[:prefix_end + 3]}…"
