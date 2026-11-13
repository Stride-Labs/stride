# Wind-down Sweep Runner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-fast:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the export-based sweep batch builder with a package that plans the v35 holder sweep from live chain state, signs and submits each batch with the `stride-sweeper` key, keeps an append-only ledger, and shows progress on a new Sweep tab of the wind-down dashboard.

**Architecture:** A stdlib-only Python package `scripts/wind-down/sweep/` (config, chain I/O seam, address helpers, ledger, holders, planner, CLI). The CLI writes `state/plan.json`, batch files and `state/ledger.jsonl`. The dashboard gets a `sweep` collector (bulk live balances, on-demand refresh) and a `/api/sweep` route that composes the plan and ledger from disk with the live snapshot; `static/sweep.js` renders the signed-off mock.

**Tech Stack:** Python 3.12 standard library (`urllib`, `subprocess`, `unittest`), vanilla JS, the `strided` CLI. Spec: `docs/superpowers/specs/2026-10-09-wind-down-sweep-runner-design.md`.

## Global Constraints

- Python conventions from `~/.agents/AGENTS.md`: module imports (`from sweep import config` is a module import and is fine; never import a function by name), every parameter and return typed, `X | None`, dataclasses for multi-field results, named parameters on calls with more than one argument, guard clauses, `Decimal` for USD, no magic strings.
- No third-party packages. No network in any test: `sweep.chainio` is the only module that opens a socket or runs a subprocess, and tests patch its functions.
- The operator key is `stride-sweeper` in keyring backend `test`. Never run a `strided keys` command without `--keyring-backend test`.
- The sweep package never imports the dashboard. The dashboard imports the package as `sweep.<module>` with `scripts/wind-down` on `sys.path`.
- Every integer in a dashboard payload is a string (the page uses BigInt). Ledger and plan files store integers as strings too.
- A `write` to a state file is atomic: write `<name>.tmp` then `os.replace`, except the ledger, which appends one line with `flush` + `fsync`.
- Editing a `.py` file under `scripts/wind-down` with the Edit tool can trigger a formatter hook that rewrites the whole file: check `git diff --stat` before committing and re-apply from the shell if the diff ballooned.
- Run the package tests with `cd scripts/wind-down && python3 -m unittest discover -s sweep -t .` and the dashboard tests with `python3 -m unittest discover -s scripts/wind-down/dashboard`. Both must pass at the end of every task.
- Commit after every task with a message in the repo's style (`wind-down sweep: <what>`), ending with the attribution lines from the session reminder.

## File structure

| Path | Responsibility |
| --- | --- |
| `scripts/wind-down/sweep/__init__.py` | Marks the package (empty). |
| `scripts/wind-down/sweep/config.py` | Every constant: denoms and rough prices, operator key, endpoints, gas, limits, paths. |
| `scripts/wind-down/sweep/exclusions.json` | The reviewed exclusion file (sections, reason, labels). |
| `scripts/wind-down/sweep/addresses.py` | bech32 validation, module and escrow address derivation, `ibc/` hashing. |
| `scripts/wind-down/sweep/chainio.py` | REST GET with pagination, `strided` subprocess, clock and sleep. The test seam. |
| `scripts/wind-down/sweep/ledger.py` | Ledger events, append/read, derived batch states, gas calibration. |
| `scripts/wind-down/sweep/holders.py` | Live reads (denoms, balances, accounts, skip inputs) and holder classification. |
| `scripts/wind-down/sweep/planner.py` | Tiers, batch packing, plan (de)serialisation, batch files, ladder. |
| `scripts/wind-down/sweep/cli.py` | `plan`, `run`, `status`, `resolve`; preflight; the `strided` command line; tx polling and event parsing. |
| `scripts/wind-down/sweep/README.md` | How to use it. |
| `scripts/wind-down/sweep/state/` | `plan.json`, `batch-*.txt`, `ledger.jsonl`, `accounts.json` (created at runtime; a `.gitkeep` is committed). |
| `scripts/wind-down/sweep/test_*.py` | One test module per module. |
| `scripts/wind-down/dashboard/sweep_tab.py` | The Sweep tab: `collect()` (bulk live balances) and `compose()` (pure). |
| `scripts/wind-down/dashboard/test_sweep_tab.py` | Tests for `compose`. |
| `scripts/wind-down/dashboard/static/sweep.js` | The page. |
| `scripts/wind-down/dashboard/server.py`, `static/index.html`, `static/app.js`, `config.py`, `README.md` | Wiring and docs. |
| `scripts/wind-down/dashboard/ops/plan.json`, `test_ops.py` | The renamed sweep steps and the new test step. |
| Deleted: `scripts/wind-down/build_sweep_batches.py`, `scripts/wind-down/test_build_sweep_batches.py`. | |

---

### Task 1: Package scaffolding: config, exclusions file, addresses, chainio

**Files:**
- Create: `scripts/wind-down/sweep/__init__.py`, `scripts/wind-down/sweep/config.py`, `scripts/wind-down/sweep/exclusions.json`, `scripts/wind-down/sweep/addresses.py`, `scripts/wind-down/sweep/chainio.py`, `scripts/wind-down/sweep/state/.gitkeep`
- Test: `scripts/wind-down/sweep/test_addresses.py`, `scripts/wind-down/sweep/test_chainio.py`

**Interfaces:**
- Consumes: `scripts/wind-down/bech32_ref.py` (`bech32_decode(bech) -> (hrp | None, data | None)`, `convertbits(data, frombits, tobits, pad)`, `encode(hrp, address_bytes) -> str`).
- Produces: everything in `config.py` below; `addresses.address_bytes(address: str) -> bytes`, `addresses.is_stride_address(address: str) -> bool`, `addresses.module_address(name: str) -> str`, `addresses.escrow_address(channel_id: str) -> str`, `addresses.ibc_denom(path: str) -> str`; `chainio.ChainError`, `chainio.NotFound(ChainError)`, `chainio.rest_get(path: str, params: dict[str, str] | None = None) -> Any`, `chainio.rest_get_all_pages(path: str, key: str, params: dict[str, str] | None = None) -> list[Any]`, `chainio.latest_height() -> int`, `chainio.CommandResult(returncode: int, stdout: str, stderr: str)`, `chainio.strided(args: list[str]) -> CommandResult`, `chainio.now() -> datetime.datetime` (UTC, aware), `chainio.sleep(seconds: float) -> None`.
- Review: no

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/sweep/test_addresses.py`:

```python
"""Address helpers: the derivations the chain uses for module accounts and transfer escrows."""

import unittest

from sweep import addresses, config


class AddressTests(unittest.TestCase):
    def test_module_address_matches_the_sdk(self) -> None:
        # authtypes.NewModuleAddress("distribution"), as printed by `strided q auth module-account distribution`
        self.assertEqual(addresses.module_address(name="distribution"), "stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan")

    def test_escrow_address_is_twenty_bytes_under_the_stride_prefix(self) -> None:
        escrow = addresses.escrow_address(channel_id="channel-5")
        self.assertTrue(escrow.startswith("stride1"))
        self.assertEqual(len(addresses.address_bytes(address=escrow)), config.ADDRESS_LENGTH_BYTES)

    def test_ibc_denom_hashes_the_trace_path(self) -> None:
        self.assertEqual(
            addresses.ibc_denom(path="transfer/channel-0/uatom"),
            "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
        )

    def test_is_stride_address_rejects_other_prefixes_and_bad_checksums(self) -> None:
        self.assertTrue(addresses.is_stride_address(address=config.SWEEP_OPERATOR))
        self.assertFalse(addresses.is_stride_address(address="osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"))
        self.assertFalse(addresses.is_stride_address(address="stride1yz3mp7c21q3qrq6krk5f8ed5qhjyhgyw2k6vgm"))
        self.assertEqual(addresses.address_bytes(address="not-bech32"), b"")


if __name__ == "__main__":
    unittest.main()
```

`scripts/wind-down/sweep/test_chainio.py`:

```python
"""chainio: the REST wrapper's error mapping and pagination, with urllib patched out."""

import io
import json
import unittest
import urllib.error
from unittest import mock

from sweep import chainio


def _http_error(code: int, body: dict) -> urllib.error.HTTPError:
    return urllib.error.HTTPError(url="u", code=code, msg="m", hdrs=None, fp=io.BytesIO(json.dumps(body).encode()))


class RestGetTests(unittest.TestCase):
    def test_404_and_sdk_not_found_bodies_raise_not_found(self) -> None:
        for code, body in ((404, {"message": "no"}), (500, {"code": 5, "message": "account stride1x not found"})):
            with self.subTest(code=code), mock.patch.object(chainio.urllib.request, "urlopen", side_effect=_http_error(code, body)):
                with self.assertRaises(chainio.NotFound):
                    chainio.rest_get(path="/cosmos/auth/v1beta1/accounts/stride1x")

    def test_other_http_errors_raise_chain_error(self) -> None:
        with mock.patch.object(chainio.urllib.request, "urlopen", side_effect=_http_error(502, {"message": "bad gateway"})):
            with self.assertRaises(chainio.ChainError):
                chainio.rest_get(path="/x")

    def test_all_pages_follows_next_key(self) -> None:
        pages = [
            {"denom_owners": [{"address": "a"}], "pagination": {"next_key": "k2"}},
            {"denom_owners": [{"address": "b"}], "pagination": {"next_key": None}},
        ]
        with mock.patch.object(chainio, "rest_get", side_effect=pages) as rest_get:
            items = chainio.rest_get_all_pages(path="/p", key="denom_owners", params={"denom": "ustrd"})
        self.assertEqual([item["address"] for item in items], ["a", "b"])
        self.assertEqual(rest_get.call_args_list[1].kwargs["params"]["pagination.key"], "k2")


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd scripts/wind-down && python3 -m unittest discover -s sweep -t . -v`
Expected: import errors (`No module named 'sweep'`).

- [ ] **Step 3: Write the package**

`scripts/wind-down/sweep/__init__.py`: empty file.

`scripts/wind-down/sweep/state/.gitkeep`: empty file.

`scripts/wind-down/sweep/config.py`:

```python
"""Every constant of the sweep runner: the denoms and their rough prices, the operator key, endpoints, gas, limits
and state paths. Edit prices here; they are a proxy for the floor, not accounting."""

import pathlib
from dataclasses import dataclass
from decimal import Decimal

PACKAGE_DIR = pathlib.Path(__file__).resolve().parent
STATE_DIR = PACKAGE_DIR / "state"
EXCLUSIONS_PATH = PACKAGE_DIR / "exclusions.json"
PLAN_PATH = STATE_DIR / "plan.json"
LEDGER_PATH = STATE_DIR / "ledger.jsonl"
ACCOUNTS_CACHE_PATH = STATE_DIR / "accounts.json"

STRIDED_BINARY = "strided"
SWEEP_OPERATOR_KEY = "stride-sweeper"
KEYRING_BACKEND = "test"  # never `os`: it prompts for the keychain password on every call
SWEEP_OPERATOR = "stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9"
TEST_ADDRESS = "stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg"  # ours; the mainnet rehearsal sweeps it alone
CHAIN_ID = "stride-1"
BINARY_VERSION_PREFIX = "v35"
REST = "https://stride-strd-api.polkachu.com"
RPC = "https://stride-strd-rpc.polkachu.com:443"
USER_AGENT = "curl/8.0"  # Polkachu rejects the urllib default
HTTP_TIMEOUT_SECONDS = 30
PAGE_SIZE = 1000

BECH32_PREFIX = "stride"
ADDRESS_LENGTH_BYTES = 20
TRANSFER_PORT = "transfer"
IBC_PREFIX = "ibc/"
STRIDE_TO_OSMOSIS_CHANNEL = "channel-5"
OSMOSIS_CHAIN_ID = "osmosis-1"
FEE_DENOM = "ustrd"

# Rough USD per whole native token (set 2026-10-09). The floor is a packet-count knob, so a factor of two is fine.
NATIVE_PRICES_USD: dict[str, Decimal] = {
    "ATOM": Decimal("4"), "OSMO": Decimal("0.3"), "TIA": Decimal("1.5"), "DYDX": Decimal("0.5"),
    "ISLM": Decimal("0.004"), "INJ": Decimal("8"), "JUNO": Decimal("0.1"), "BAND": Decimal("0.3"),
    "LUNA": Decimal("0.1"), "SOMM": Decimal("0.01"), "SAGA": Decimal("0.1"), "STRD": Decimal("0.05"),
}


@dataclass(frozen=True)
class NativeDenom:
    """A Stride-native sweep denom: an stToken (priced at its host's price times the zone's redemption rate) or ustrd."""

    denom: str
    symbol: str
    decimals: int
    price_symbol: str
    host_chain_id: str | None  # the stakeibc host zone whose redemption rate prices it; None for ustrd


@dataclass(frozen=True)
class VoucherDenom:
    """An IBC voucher that unwinds one hop: `ibc/sha256(transfer/<channel>/<base>)`, sent back over that channel."""

    base: str
    channel: str
    symbol: str
    decimals: int
    price_symbol: str
    chain_id: str


# To Osmosis over channel-5: the eleven in-scope stTokens and ustrd (spec §7).
NATIVE_SWEEP_DENOMS: tuple[NativeDenom, ...] = (
    NativeDenom("stuatom", "stATOM", 6, "ATOM", "cosmoshub-4"),
    NativeDenom("stuosmo", "stOSMO", 6, "OSMO", "osmosis-1"),
    NativeDenom("stutia", "stTIA", 6, "TIA", "celestia"),
    NativeDenom("stinj", "stINJ", 18, "INJ", "injective-1"),
    NativeDenom("stadydx", "stDYDX", 18, "DYDX", "dydx-mainnet-1"),
    NativeDenom("staISLM", "stISLM", 18, "ISLM", "haqq_11235-1"),
    NativeDenom("stujuno", "stJUNO", 6, "JUNO", "juno-1"),
    NativeDenom("stuband", "stBAND", 6, "BAND", "laozi-mainnet"),
    NativeDenom("stuluna", "stLUNA", 6, "LUNA", "phoenix-1"),
    NativeDenom("stusomm", "stSOMM", 6, "SOMM", "sommelier-3"),
    NativeDenom("stusaga", "stSAGA", 6, "SAGA", "ssc-1"),
    NativeDenom("ustrd", "STRD", 6, "STRD", None),
)
# Back one hop over the channel each arrived on. Must stay a subset of types.SweepUnwindChannels
# (x/stakeibc/types/wind_down.go); the planner re-checks every one against the live denom trace.
VOUCHER_SWEEP_DENOMS: tuple[VoucherDenom, ...] = (
    VoucherDenom("uatom", "channel-0", "ATOM", 6, "ATOM", "cosmoshub-4"),
    VoucherDenom("utia", "channel-162", "TIA", 6, "TIA", "celestia"),
    VoucherDenom("uosmo", "channel-5", "OSMO", 6, "OSMO", "osmosis-1"),
    VoucherDenom("ujuno", "channel-24", "JUNO", 6, "JUNO", "juno-1"),
    VoucherDenom("usomm", "channel-150", "SOMM", 6, "SOMM", "sommelier-3"),
    VoucherDenom("usaga", "channel-213", "SAGA", 6, "SAGA", "ssc-1"),
    VoucherDenom("adydx", "channel-160", "DYDX", 18, "DYDX", "dydx-mainnet-1"),
)
# Mirror of types.SweepUnwindChannels: Stride channel -> counterparty bech32 prefix.
UNWIND_CHANNELS: dict[str, str] = {
    "channel-0": "cosmos", "channel-162": "celestia", "channel-5": "osmo", "channel-24": "juno",
    "channel-150": "somm", "channel-213": "saga", "channel-160": "dydx",
}
# Mirror of isSweepableNativeDenom's constants; the rest are st<host_denom> of non-deprecated host zones.
ALWAYS_SWEEPABLE_NATIVE_DENOMS = frozenset({"ustrd", "stutia"})

# Mirror of types.SweepProtocolAddresses: staketia S0-S3, stakedym S4-S7, the staketia operator.
PROTOCOL_ADDRESSES = frozenset({
    "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04",  # staketia deposit
    "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g",  # staketia redemption
    "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd",  # staketia claim
    "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj",  # stakedym deposit
    "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n",  # stakedym redemption
    "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras",  # stakedym claim
    "stride18p7xg4hj2u3zpk0v9gq68pjyuuua5wa387sjjc",  # staketia safe
    "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7",  # stakedym safe
    "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw",  # staketia operator
})
# Module accounts the bank keeper blocks (app.BlacklistedModuleAccountAddrs); keep in sync with app/app.go.
BLOCKED_MODULE_NAMES = (
    "fee_collector", "distribution", "cons_redistribute", "mint", "bonded_tokens_pool", "not_bonded_tokens_pool",
    "gov", "transfer", "claim", "interchainquery", "interchainaccounts", "wasm", "icqoracle", "auction",
    "strdburner", "poa",
)

BASE_ACCOUNT = "/cosmos.auth.v1beta1.BaseAccount"
MODULE_ACCOUNT = "/cosmos.auth.v1beta1.ModuleAccount"
INTERCHAIN_ACCOUNT = "/ibc.applications.interchain_accounts.v1.InterchainAccount"
VESTING_ACCOUNT_TYPES = frozenset({
    "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
    "/cosmos.vesting.v1beta1.DelayedVestingAccount",
    "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
    "/stride.vesting.StridePeriodicVestingAccount",
})
SWEEPABLE_ACCOUNT_TYPES = frozenset({BASE_ACCOUNT}) | VESTING_ACCOUNT_TYPES

# Batches: packed under a gas budget at plan time, simulated at run time, refused over the block limit.
MAX_ADDRESSES_PER_BATCH = 100
GAS_BUDGET_PER_BATCH = 40_000_000
BLOCK_GAS_LIMIT = 100_000_000  # stride-1's consensus max_gas
GAS_PER_TRANSFER_DEFAULT = 150_000  # until the ledger has a confirmed batch to calibrate from
CALIBRATION_MARGIN = Decimal("1.2")
GAS_ADJUSTMENT = Decimal("1.3")
GAS_PRICE_USTRD = Decimal("0.001")  # stride-1's minimum is 0.0005ustrd

MAX_PLAN_AGE_SECONDS = 6 * 3600
TX_POLL_SECONDS = 3
TX_WAIT_SECONDS = 180
RESOLVE_WAIT_SECONDS = 600
# The floor ladder printed by `plan` and shown on the dashboard: the plan's floor, then these below it.
LADDER_FLOORS: tuple[Decimal, ...] = (Decimal(10), Decimal(5), Decimal(1), Decimal(0))
```

`scripts/wind-down/sweep/exclusions.json`:

```json
{
  "sections": [
    {
      "name": "team",
      "reason": "moved by hand; v35 also sends the community pool here",
      "addresses": [
        {"address": "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh", "label": "F5 team multisig"}
      ]
    },
    {
      "name": "relayers",
      "reason": "their STRD pays for relaying until the halt",
      "addresses": [
        {"address": "stride1e6llcr7fkxvqdgyrcgzdlwll9tkvfh2rnfcpyd", "label": "relayer (by hand)"},
        {"address": "stride1fegapd4jc3ejqeg0eu3jk4hvr74hg660a3gcsp", "label": "relayer (by hand)"},
        {"address": "stride1x3mtu4540z8q45mmyn9tw34fmy74xqm47lh7k6", "label": "relayer-band"},
        {"address": "stride15fd7g4g26enekm50ve9dhwaseanav09my88559", "label": "relayer-celestia"},
        {"address": "stride18qysgypvp7aky0dntqzkl6uxv7w3zszt55uv0m", "label": "relayer-dydx"},
        {"address": "stride1c89e5smtnem6lvmqpajhakk7hvkm2luq5cqqm3", "label": "relayer-dymension"},
        {"address": "stride1ds5klkz9s0v3ky4j2dmn377pd7wxtyxq5c7dk3", "label": "relayer-gaia"},
        {"address": "stride10cyp5nu07hrymlhazemfh2u8khf3m9xc2xw535", "label": "relayer-ics"},
        {"address": "stride1ge22qnyg4avg0glvz00ruwpulhe9p7qnxe4426", "label": "relayer-injective"},
        {"address": "stride1d5e6c09mt37qcusfrrrfvz4dwnaxlv3ltwahfn", "label": "relayer-juno"},
        {"address": "stride1wme78wwkmtkjax588xt7ntl743fq8vdteq2kt3", "label": "relayer-osmosis"},
        {"address": "stride1h93r6y9sm2dtq68dqwvqp36faqucge7wzfxhws", "label": "relayer-saga"},
        {"address": "stride15s0ey2k5sqhkdu0d2770y8cykgm6emx2fyar7v", "label": "relayer-sei"},
        {"address": "stride1qa8l9le4fvvxe4ny2m2mcd00cakgx9lz96kash", "label": "relayer-sommelier"},
        {"address": "stride1m5st93c4elt4ttt2kdzgm9cpj953thkk5l0k9d", "label": "relayer-terra"}
      ]
    }
  ]
}
```

`scripts/wind-down/sweep/addresses.py`:

```python
"""Address helpers: bech32 validation and the derivations the chain uses (module accounts, ICS-20 escrows, ibc/ denoms).
The vendored bech32 reference implementation lives one directory up."""

import hashlib
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # for bech32_ref, shared with coverage_check.py

import bech32_ref  # noqa: E402

from sweep import config  # noqa: E402

ESCROW_ADDRESS_VERSION = "ics20-1"


def address_bytes(address: str) -> bytes:
    """The raw bytes behind a bech32 address, or b"" when the string is not valid bech32."""
    _, data = bech32_ref.bech32_decode(address)
    if data is None:
        return b""
    converted = bech32_ref.convertbits(data, 5, 8, False)
    return bytes(converted) if converted is not None else b""


def is_stride_address(address: str) -> bool:
    """A valid `stride1…` address of exactly 20 bytes (the chain's first skip rule)."""
    hrp, _ = bech32_ref.bech32_decode(address)
    return hrp == config.BECH32_PREFIX and len(address_bytes(address=address)) == config.ADDRESS_LENGTH_BYTES


def module_address(name: str) -> str:
    """SDK authtypes.NewModuleAddress: sha256(name)[:20], bech32 stride."""
    return bech32_ref.encode(config.BECH32_PREFIX, hashlib.sha256(name.encode()).digest()[: config.ADDRESS_LENGTH_BYTES])


def escrow_address(channel_id: str) -> str:
    """ibc-go transfertypes.GetEscrowAddress: sha256("ics20-1\\0" + port/channel)[:20], bech32 stride."""
    pre_image = ESCROW_ADDRESS_VERSION.encode() + b"\x00" + f"{config.TRANSFER_PORT}/{channel_id}".encode()
    return bech32_ref.encode(config.BECH32_PREFIX, hashlib.sha256(pre_image).digest()[: config.ADDRESS_LENGTH_BYTES])


def ibc_denom(path: str) -> str:
    """The on-chain denom of an IBC voucher from its full trace path, e.g. `transfer/channel-0/uatom`."""
    return config.IBC_PREFIX + hashlib.sha256(path.encode()).hexdigest().upper()
```

`scripts/wind-down/sweep/chainio.py`:

```python
"""The package's only I/O: Stride REST reads, the strided subprocess, the clock. Tests patch these functions."""

import datetime
import json
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Any

from sweep import config

SDK_NOT_FOUND_CODE = 5


class ChainError(Exception):
    """A REST call failed (network, HTTP error, bad JSON)."""


class NotFound(ChainError):
    """The chain answered that the thing does not exist: an unknown account, a tx not yet indexed."""


@dataclass(frozen=True)
class CommandResult:
    returncode: int
    stdout: str
    stderr: str


def rest_get(path: str, params: dict[str, str] | None = None) -> Any:
    """GET a Stride REST path and decode the JSON body, retrying once on a network error."""
    query = f"?{urllib.parse.urlencode(params)}" if params else ""
    url = f"{config.REST}{path}{query}"
    try:
        return _fetch_json(url=url)
    except (urllib.error.URLError, TimeoutError):
        pass
    try:
        return _fetch_json(url=url)
    except urllib.error.HTTPError as error:
        raise _http_error(path=path, error=error) from error
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as error:
        raise ChainError(f"{path}: {error}") from error


def rest_get_all_pages(path: str, key: str, params: dict[str, str] | None = None) -> list[Any]:
    """Collect `response[key]` across Cosmos `pagination.key` / `next_key` pages."""
    items: list[Any] = []
    page_params = {"pagination.limit": str(config.PAGE_SIZE), **(params or {})}
    while True:
        response = rest_get(path=path, params=page_params)
        items.extend(response.get(key) or [])
        next_key = (response.get("pagination") or {}).get("next_key")
        if not next_key:
            return items
        page_params = {**page_params, "pagination.key": next_key}


def latest_height() -> int:
    body = rest_get(path="/cosmos/base/tendermint/v1beta1/blocks/latest")
    return int(body["block"]["header"]["height"])


def strided(args: list[str]) -> CommandResult:
    """Run the strided binary with the given arguments; the caller decides what a non-zero exit means."""
    completed = subprocess.run([config.STRIDED_BINARY, *args], capture_output=True, text=True, check=False)
    return CommandResult(returncode=completed.returncode, stdout=completed.stdout, stderr=completed.stderr)


def now() -> datetime.datetime:
    return datetime.datetime.now(datetime.UTC)


def sleep(seconds: float) -> None:
    time.sleep(seconds)


def _fetch_json(url: str) -> Any:
    request = urllib.request.Request(url, headers={"User-Agent": config.USER_AGENT})
    with urllib.request.urlopen(request, timeout=config.HTTP_TIMEOUT_SECONDS) as response:
        return json.load(response)


def _http_error(path: str, error: urllib.error.HTTPError) -> ChainError:
    """gRPC-gateway answers 404 for a missing tx, but 500 with SDK code 5 for a missing account: both are NotFound."""
    try:
        body = json.loads(error.read() or b"{}")
    except (json.JSONDecodeError, OSError):
        body = {}
    message = str(body.get("message", ""))
    if error.code == 404 or body.get("code") == SDK_NOT_FOUND_CODE or "not found" in message.lower():
        return NotFound(f"{path}: {message or error.code}")
    return ChainError(f"{path}: HTTP {error.code} {message}")
```

Note on the retry in `rest_get`: the first attempt swallows only network errors; an `HTTPError` is a `URLError` subclass, so catch `HTTPError` first if you restructure. Simplest correct form is the one above, because `HTTPError` raised on the first attempt is re-raised by the second attempt's handler only if it happens twice; to keep one code path, implement `_fetch_json_with_retry` as: try once; on `URLError` that is not an `HTTPError`, try again; let `HTTPError` propagate to the mapping. Write it that way:

```python
def rest_get(path: str, params: dict[str, str] | None = None) -> Any:
    query = f"?{urllib.parse.urlencode(params)}" if params else ""
    url = f"{config.REST}{path}{query}"
    try:
        return _fetch_json_with_retry(url=url)
    except urllib.error.HTTPError as error:
        raise _http_error(path=path, error=error) from error
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as error:
        raise ChainError(f"{path}: {error}") from error


def _fetch_json_with_retry(url: str) -> Any:
    try:
        return _fetch_json(url=url)
    except urllib.error.HTTPError:
        raise
    except (urllib.error.URLError, TimeoutError):
        return _fetch_json(url=url)
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && python3 -m unittest discover -s sweep -t . -v`
Expected: 7 tests, OK.

- [ ] **Step 5: Commit**

```bash
git add scripts/wind-down/sweep
git commit -m "wind-down sweep: package scaffolding (config, exclusions file, address helpers, chain I/O seam)"
```

---

### Task 2: The ledger

**Files:**
- Create: `scripts/wind-down/sweep/ledger.py`
- Test: `scripts/wind-down/sweep/test_ledger.py`

**Interfaces:**
- Consumes: `config.LEDGER_PATH`, `config.CALIBRATION_MARGIN`.
- Produces:
  - `class BatchState(StrEnum): PENDING, SUBMITTED, CONFIRMED, FAILED, LOST`
  - `class Kind(StrEnum): SUBMITTED, CONFIRMED, FAILED, LOST`
  - `@dataclass Transfer(address: str, denom: str, amount: int, channel: str, receiver: str)`
  - `@dataclass Skip(address: str, reason: str)`
  - `@dataclass Event(kind: Kind, batch_id: str, at: str, run_id: int | None = None, tx_hash: str | None = None, height: int | None = None, gas_wanted: int | None = None, gas_used: int | None = None, addresses: int | None = None, transfers_estimate: int | None = None, transfers: list[Transfer] = [], skipped: list[Skip] = [], code: int | None = None, codespace: str | None = None, raw_log: str | None = None)`
  - `submitted(run_id, batch_id, tx_hash, at, addresses, transfers_estimate, gas_wanted) -> Event`, `confirmed(batch_id, tx_hash, at, height, gas_used, transfers, skipped) -> Event`, `failed(batch_id, tx_hash, at, code, codespace, raw_log) -> Event`, `lost(batch_id, tx_hash, at) -> Event`
  - `append(event: Event, path: pathlib.Path = config.LEDGER_PATH) -> None`
  - `read(path: pathlib.Path = config.LEDGER_PATH) -> list[Event]` (missing file → `[]`; a bad line raises `LedgerError` naming the line number)
  - `batch_states(events: list[Event]) -> dict[str, BatchState]`
  - `unresolved(events: list[Event]) -> list[Event]` (submitted events with no terminal event for the batch)
  - `calibration(events: list[Event]) -> int | None` (max `gas_used / len(transfers)` over confirmed events with transfers, times `CALIBRATION_MARGIN`, as int)
  - `confirmed_transfers(events: list[Event]) -> dict[str, list[Transfer]]` (address → every confirmed transfer)
  - `latest_run_id(events: list[Event]) -> int` (0 when empty)
  - `event_to_dict(event) -> dict`, `event_from_dict(data) -> Event` (ints as strings on disk)
- Review: yes (the ledger is the resume state for a tx that moves user funds)

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/sweep/test_ledger.py`:

```python
"""The ledger: append-only JSON lines, the batch states derived from them, and gas calibration."""

import pathlib
import tempfile
import unittest

from sweep import ledger

T1 = ledger.Transfer(address="stride1a", denom="stuatom", amount=1_000_000, channel="channel-5", receiver="osmo1a")
T2 = ledger.Transfer(address="stride1a", denom="ustrd", amount=5, channel="channel-5", receiver="osmo1a")


class LedgerFileTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.path = pathlib.Path(self.tmp.name) / "ledger.jsonl"

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_missing_file_reads_empty(self) -> None:
        self.assertEqual(ledger.read(path=self.path), [])

    def test_round_trip_keeps_every_field(self) -> None:
        submitted = ledger.submitted(run_id=1, batch_id="001-001", tx_hash="AB", at="t0", addresses=1, transfers_estimate=2, gas_wanted=300)
        confirmed = ledger.confirmed(batch_id="001-001", tx_hash="AB", at="t1", height=10, gas_used=240, transfers=[T1, T2], skipped=[ledger.Skip(address="stride1b", reason="interchain account")])
        ledger.append(event=submitted, path=self.path)
        ledger.append(event=confirmed, path=self.path)
        self.assertEqual(ledger.read(path=self.path), [submitted, confirmed])
        self.assertEqual(len(self.path.read_text().splitlines()), 2)

    def test_truncated_line_is_reported_with_its_number(self) -> None:
        ledger.append(event=ledger.lost(batch_id="001-001", tx_hash="AB", at="t"), path=self.path)
        with self.path.open("a") as handle:
            handle.write('{"kind": "confirmed", "batch_id": "001-0')
        with self.assertRaisesRegex(ledger.LedgerError, "line 2"):
            ledger.read(path=self.path)


class DerivedStateTests(unittest.TestCase):
    def test_batch_states_and_unresolved(self) -> None:
        events = [
            ledger.submitted(run_id=1, batch_id="001-001", tx_hash="A", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=1, transfers=[T1], skipped=[]),
            ledger.submitted(run_id=1, batch_id="001-002", tx_hash="B", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.submitted(run_id=1, batch_id="001-003", tx_hash="C", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.failed(batch_id="001-003", tx_hash="C", at="t", code=11, codespace="sdk", raw_log="out of gas"),
            ledger.submitted(run_id=1, batch_id="001-004", tx_hash="D", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.lost(batch_id="001-004", tx_hash="D", at="t"),
        ]
        states = ledger.batch_states(events=events)
        self.assertEqual(states, {
            "001-001": ledger.BatchState.CONFIRMED, "001-002": ledger.BatchState.SUBMITTED,
            "001-003": ledger.BatchState.FAILED, "001-004": ledger.BatchState.LOST,
        })
        self.assertEqual([event.batch_id for event in ledger.unresolved(events=events)], ["001-002"])
        self.assertEqual(ledger.latest_run_id(events=events), 1)
        self.assertEqual(ledger.confirmed_transfers(events=events), {"stride1a": [T1]})

    def test_calibration_uses_the_worst_confirmed_ratio_with_margin(self) -> None:
        self.assertIsNone(ledger.calibration(events=[]))
        events = [
            ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=200_000, transfers=[T1, T2], skipped=[]),
            ledger.confirmed(batch_id="001-002", tx_hash="B", at="t", height=2, gas_used=330_000, transfers=[T1, T2, T1], skipped=[]),
            ledger.confirmed(batch_id="001-003", tx_hash="C", at="t", height=3, gas_used=50_000, transfers=[], skipped=[]),
        ]
        self.assertEqual(ledger.calibration(events=events), 132_000)  # 110,000 per transfer x 1.2


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_ledger -v`
Expected: `ModuleNotFoundError: No module named 'sweep.ledger'`.

- [ ] **Step 3: Write the ledger**

`scripts/wind-down/sweep/ledger.py`:

```python
"""The append-only record of every batch submission and its outcome: state/ledger.jsonl, one JSON object per line.
Nothing here is ever rewritten; resume, status and the dashboard all derive from the lines."""

import dataclasses
import json
import os
import pathlib
from dataclasses import dataclass, field
from decimal import Decimal
from enum import StrEnum

from sweep import config


class LedgerError(Exception):
    """A line of the ledger cannot be read."""


class Kind(StrEnum):
    SUBMITTED = "submitted"
    CONFIRMED = "confirmed"
    FAILED = "failed"
    LOST = "lost"


class BatchState(StrEnum):
    PENDING = "pending"
    SUBMITTED = "submitted"
    CONFIRMED = "confirmed"
    FAILED = "failed"
    LOST = "lost"


TERMINAL_KINDS = {Kind.CONFIRMED, Kind.FAILED, Kind.LOST}
INT_FIELDS = ("run_id", "height", "gas_wanted", "gas_used", "addresses", "transfers_estimate", "code")


@dataclass(frozen=True)
class Transfer:
    address: str
    denom: str
    amount: int
    channel: str
    receiver: str


@dataclass(frozen=True)
class Skip:
    address: str
    reason: str


@dataclass(frozen=True)
class Event:
    kind: Kind
    batch_id: str
    at: str
    run_id: int | None = None
    tx_hash: str | None = None
    height: int | None = None
    gas_wanted: int | None = None
    gas_used: int | None = None
    addresses: int | None = None
    transfers_estimate: int | None = None
    transfers: list[Transfer] = field(default_factory=list)
    skipped: list[Skip] = field(default_factory=list)
    code: int | None = None
    codespace: str | None = None
    raw_log: str | None = None


# ---- constructors


def submitted(run_id: int, batch_id: str, tx_hash: str, at: str, addresses: int, transfers_estimate: int, gas_wanted: int) -> Event:
    return Event(kind=Kind.SUBMITTED, batch_id=batch_id, at=at, run_id=run_id, tx_hash=tx_hash, addresses=addresses,
                 transfers_estimate=transfers_estimate, gas_wanted=gas_wanted)


def confirmed(batch_id: str, tx_hash: str, at: str, height: int, gas_used: int, transfers: list[Transfer], skipped: list[Skip]) -> Event:
    return Event(kind=Kind.CONFIRMED, batch_id=batch_id, at=at, tx_hash=tx_hash, height=height, gas_used=gas_used,
                 transfers=list(transfers), skipped=list(skipped))


def failed(batch_id: str, tx_hash: str, at: str, code: int, codespace: str, raw_log: str) -> Event:
    return Event(kind=Kind.FAILED, batch_id=batch_id, at=at, tx_hash=tx_hash, code=code, codespace=codespace, raw_log=raw_log)


def lost(batch_id: str, tx_hash: str, at: str) -> Event:
    return Event(kind=Kind.LOST, batch_id=batch_id, at=at, tx_hash=tx_hash)


# ---- file


def append(event: Event, path: pathlib.Path = config.LEDGER_PATH) -> None:
    """One line, flushed and fsynced, so a crash leaves at most a torn last line (which `read` reports)."""
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(event_to_dict(event=event), separators=(",", ":")) + "\n")
        handle.flush()
        os.fsync(handle.fileno())


def read(path: pathlib.Path = config.LEDGER_PATH) -> list[Event]:
    if not path.exists():
        return []
    events: list[Event] = []
    for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        if not line.strip():
            continue
        try:
            events.append(event_from_dict(data=json.loads(line)))
        except (json.JSONDecodeError, KeyError, ValueError) as error:
            raise LedgerError(f"{path} line {number} is unreadable: {error}") from None
    return events


def event_to_dict(event: Event) -> dict:
    data = dataclasses.asdict(event)
    data["kind"] = event.kind.value
    for name in INT_FIELDS:
        if data[name] is not None:
            data[name] = str(data[name])
    for transfer in data["transfers"]:
        transfer["amount"] = str(transfer["amount"])
    return data


def event_from_dict(data: dict) -> Event:
    return Event(
        kind=Kind(data["kind"]),
        batch_id=data["batch_id"],
        at=data["at"],
        tx_hash=data.get("tx_hash"),
        codespace=data.get("codespace"),
        raw_log=data.get("raw_log"),
        transfers=[Transfer(address=t["address"], denom=t["denom"], amount=int(t["amount"]), channel=t["channel"], receiver=t["receiver"]) for t in data.get("transfers", [])],
        skipped=[Skip(address=s["address"], reason=s["reason"]) for s in data.get("skipped", [])],
        **{name: (int(data[name]) if data.get(name) is not None else None) for name in INT_FIELDS},
    )


# ---- derived state


def batch_states(events: list[Event]) -> dict[str, BatchState]:
    states: dict[str, BatchState] = {}
    for event in events:
        states[event.batch_id] = BatchState(event.kind.value)
    return states


def unresolved(events: list[Event]) -> list[Event]:
    """Submitted batches that have no confirmed, failed or lost line yet."""
    terminal = {event.batch_id for event in events if event.kind in TERMINAL_KINDS}
    return [event for event in events if event.kind == Kind.SUBMITTED and event.batch_id not in terminal]


def calibration(events: list[Event]) -> int | None:
    """Gas per transfer for the planner: the worst confirmed ratio with a margin, or None before any confirmation."""
    ratios = [Decimal(event.gas_used) / len(event.transfers) for event in events
              if event.kind == Kind.CONFIRMED and event.gas_used is not None and event.transfers]
    if not ratios:
        return None
    return int(max(ratios) * config.CALIBRATION_MARGIN)


def confirmed_transfers(events: list[Event]) -> dict[str, list[Transfer]]:
    by_address: dict[str, list[Transfer]] = {}
    for event in events:
        if event.kind != Kind.CONFIRMED:
            continue
        for transfer in event.transfers:
            by_address.setdefault(transfer.address, []).append(transfer)
    return by_address


def latest_run_id(events: list[Event]) -> int:
    return max((event.run_id for event in events if event.run_id is not None), default=0)
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_ledger -v`
Expected: 5 tests, OK.

- [ ] **Step 5: Commit**

```bash
git add scripts/wind-down/sweep/ledger.py scripts/wind-down/sweep/test_ledger.py
git commit -m "wind-down sweep: the append-only ledger and the batch states derived from it"
```

---
### Task 3: Holders: live reads and classification

**Files:**
- Create: `scripts/wind-down/sweep/holders.py`
- Test: `scripts/wind-down/sweep/test_holders.py`

**Interfaces:**
- Consumes: `chainio.rest_get`, `chainio.rest_get_all_pages`, `chainio.latest_height`, `chainio.NotFound`; `addresses.*`; `config.*`.
- Produces:
  - `@dataclass SweepDenom(denom: str, symbol: str, decimals: int, price_usd: Decimal, destination: str, channel: str)`
  - `@dataclass Holder(address: str, balances: dict[str, int], usd: Decimal, keyless: bool)` with property `transfers -> int` (denoms with a positive amount)
  - `@dataclass Excluded(address: str, reason: str, usd: Decimal)`
  - `@dataclass HolderSet(height: int, denoms: list[SweepDenom], holders: list[Holder], excluded: list[Excluded], skipped: list[Excluded], below_floor: list[Holder])` (`excluded` = exclusion file, `skipped` = chain rules and contracts)
  - `@dataclass Exclusion(address: str, section: str, label: str, reason: str)`
  - `class ExclusionsError(Exception)`, `class DenomError(Exception)`
  - `load_exclusions(path: pathlib.Path = config.EXCLUSIONS_PATH) -> dict[str, Exclusion]`
  - `resolve_denoms() -> list[SweepDenom]` (live; raises `DenomError`)
  - `read_balances(denoms: list[SweepDenom]) -> dict[str, dict[str, int]]` (address → denom → amount, zero amounts dropped)
  - `usd_value(balances: dict[str, int], denoms: dict[str, SweepDenom]) -> Decimal`
  - `@dataclass AccountInfo(type: str, has_pubkey: bool, sequence: int)`
  - `class AccountCache` with `load(path) -> AccountCache` (classmethod), `save(path) -> None`, `lookup(address: str) -> AccountInfo | None` (None = account not found; reads through `chainio` and caches; a cached entry with `has_pubkey=False` is re-read)
  - `@dataclass SkipInputs(module_addresses: set[str], escrows: set[str], contracts: set[str])`, `read_skip_inputs() -> SkipInputs`
  - `spendable_balances(address: str) -> dict[str, int]`
  - `skip_reason(address: str, account: AccountInfo | None, inputs: SkipInputs, exclusions: dict[str, Exclusion]) -> str | None`
  - `classify(denoms, balances, floor_usd: Decimal, exclusions, inputs, accounts: AccountCache, test_address: str | None, height: int) -> HolderSet`
  - `read_holder_set(floor_usd: Decimal, exclusions, accounts: AccountCache, test_address: str | None) -> HolderSet` (height, denoms, balances, skip inputs, classify)
- Review: yes (decides which accounts get swept)

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/sweep/test_holders.py`:

```python
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
EXCLUDED = addresses.module_address(name="not-a-real-module-just-twenty-bytes")
ATOM = addresses.ibc_denom(path="transfer/channel-0/uatom")

HOST_ZONES = [
    {"chain_id": "cosmoshub-4", "host_denom": "uatom", "redemption_rate": "1.5", "deprecated": False},
    {"chain_id": "celestia", "host_denom": "utia", "redemption_rate": "1.2", "deprecated": False},
    {"chain_id": "evmos_9001-2", "host_denom": "aevmos", "redemption_rate": "1.1", "deprecated": True},
]
OWNERS = {
    "stuatom": [(BASE, "10000000"), (VESTING, "2000000"), (KEYLESS, "1000000"), (ICA, "99000000"), (MISSING, "99000000"),
                (DUST, "1000"), (MODULE, "99000000"), (ESCROW, "99000000"), (CONTRACT, "99000000"), (EXCLUDED, "99000000")],
    "ustrd": [(BASE, "5000000"), (VESTING, "100000000")],
    ATOM: [(BASE, "250000")],
}
ACCOUNTS = {
    BASE: {"@type": config.BASE_ACCOUNT, "address": BASE, "pub_key": {"key": "x"}, "sequence": "4"},
    VESTING: {"@type": "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
              "base_vesting_account": {"base_account": {"address": VESTING, "pub_key": {"key": "x"}, "sequence": "1"}}},
    KEYLESS: {"@type": config.BASE_ACCOUNT, "address": KEYLESS, "pub_key": None, "sequence": "0"},
    ICA: {"@type": config.INTERCHAIN_ACCOUNT, "base_account": {"address": ICA, "pub_key": None, "sequence": "0"}},
    DUST: {"@type": config.BASE_ACCOUNT, "address": DUST, "pub_key": {"key": "x"}, "sequence": "9"},
    MODULE: {"@type": config.MODULE_ACCOUNT, "base_account": {"address": MODULE, "pub_key": None, "sequence": "0"}, "name": "distribution"},
    ESCROW: {"@type": config.BASE_ACCOUNT, "address": ESCROW, "pub_key": None, "sequence": "0"},
    CONTRACT: {"@type": config.BASE_ACCOUNT, "address": CONTRACT, "pub_key": None, "sequence": "0"},
    EXCLUDED: {"@type": config.BASE_ACCOUNT, "address": EXCLUDED, "pub_key": {"key": "x"}, "sequence": "2"},
}
SPENDABLE = {VESTING: [{"denom": "stuatom", "amount": "2000000"}, {"denom": "ustrd", "amount": "40000000"}]}


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
        self.assertEqual([h.address for h in holder_set.below_floor], [DUST])
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
        holder_set = holders.read_holder_set(floor_usd=Decimal("1000"), exclusions=exclusions, accounts=accounts, test_address=DUST)
        self.assertEqual([h.address for h in holder_set.holders], [DUST])

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


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_holders -v`
Expected: `ModuleNotFoundError: No module named 'sweep.holders'`.

- [ ] **Step 3: Write holders.py**

```python
"""Live holder state and the sweep's inclusion rules.

Reads, all through chainio: the sweep denoms (checked the way the chain checks them), every holder of each denom in bulk,
the auth account of each candidate above the floor (cached), the spendable balance of vesting candidates, and the
module, escrow and contract address sets. Classification mirrors sweepSkipReason in x/stakeibc/keeper/wind_down_sweep.go
in the chain's order, then adds the two rules the chain cannot make: wasm contracts and the reviewed exclusions file."""

import json
import pathlib
from dataclasses import dataclass, field
from decimal import Decimal

from sweep import addresses, chainio, config

STTOKEN_PREFIX = "st"
STATE_OPEN = "STATE_OPEN"


class ExclusionsError(Exception):
    """exclusions.json is malformed."""


class DenomError(Exception):
    """A configured sweep denom has no on-chain destination; the chain would reject every batch naming it."""


@dataclass(frozen=True)
class SweepDenom:
    denom: str
    symbol: str
    decimals: int
    price_usd: Decimal  # per whole token
    destination: str  # chain id
    channel: str  # the Stride transfer channel the sweep sends over


@dataclass
class Holder:
    address: str
    balances: dict[str, int]
    usd: Decimal
    keyless: bool

    @property
    def transfers(self) -> int:
        return sum(1 for amount in self.balances.values() if amount > 0)


@dataclass(frozen=True)
class Excluded:
    address: str
    reason: str
    usd: Decimal


@dataclass
class HolderSet:
    height: int
    denoms: list[SweepDenom]
    holders: list[Holder]
    excluded: list[Excluded]  # the exclusions file
    skipped: list[Excluded]  # the chain's rules and the contract rule
    below_floor: list[Holder]


@dataclass(frozen=True)
class Exclusion:
    address: str
    section: str
    label: str
    reason: str


@dataclass(frozen=True)
class AccountInfo:
    type: str
    has_pubkey: bool
    sequence: int


@dataclass(frozen=True)
class SkipInputs:
    module_addresses: set[str]
    escrows: set[str]
    contracts: set[str]


# ---- exclusions file


def load_exclusions(path: pathlib.Path = config.EXCLUSIONS_PATH) -> dict[str, Exclusion]:
    try:
        sections = json.loads(path.read_text())["sections"]
    except (OSError, json.JSONDecodeError, KeyError) as error:
        raise ExclusionsError(f"{path}: {error}") from None

    exclusions: dict[str, Exclusion] = {}
    for section in sections:
        for entry in section["addresses"]:
            address = entry["address"]
            if not addresses.is_stride_address(address=address):
                raise ExclusionsError(f"{path}: {address!r} in section {section['name']} is not a 20-byte stride address")
            if address in exclusions:
                raise ExclusionsError(f"{path}: {address} is listed twice")
            exclusions[address] = Exclusion(address=address, section=section["name"], label=entry["label"], reason=section["reason"])
    return exclusions


# ---- denoms


def resolve_denoms() -> list[SweepDenom]:
    """The configured denoms, each checked the way resolveSweepDestination checks it, with its price."""
    host_zones = chainio.rest_get_all_pages(path="/Stride-Labs/stride/stakeibc/host_zone", key="host_zone")
    zones_by_chain = {zone["chain_id"]: zone for zone in host_zones}
    allowed_native = set(config.ALWAYS_SWEEPABLE_NATIVE_DENOMS) | {
        STTOKEN_PREFIX + zone["host_denom"] for zone in host_zones if not zone.get("deprecated")
    }

    denoms = [_native_denom(spec=spec, allowed=allowed_native, zones_by_chain=zones_by_chain) for spec in config.NATIVE_SWEEP_DENOMS]
    denoms.extend(_voucher_denom(spec=spec) for spec in config.VOUCHER_SWEEP_DENOMS)
    return denoms


def _native_denom(spec: config.NativeDenom, allowed: set[str], zones_by_chain: dict[str, dict]) -> SweepDenom:
    if spec.denom not in allowed:
        raise DenomError(f"native denom {spec.denom} is neither ustrd, stutia nor an active host zone's stToken")
    rate = Decimal(1)
    if spec.host_chain_id is not None:
        zone = zones_by_chain.get(spec.host_chain_id)
        if zone is None:
            raise DenomError(f"{spec.denom}: host zone {spec.host_chain_id} not found")
        rate = Decimal(zone["redemption_rate"])
    return SweepDenom(denom=spec.denom, symbol=spec.symbol, decimals=spec.decimals,
                      price_usd=config.NATIVE_PRICES_USD[spec.price_symbol] * rate,
                      destination=config.OSMOSIS_CHAIN_ID, channel=config.STRIDE_TO_OSMOSIS_CHANNEL)


def _voucher_denom(spec: config.VoucherDenom) -> SweepDenom:
    if spec.channel not in config.UNWIND_CHANNELS:
        raise DenomError(f"{spec.base} over {spec.channel}: not in UNWIND_CHANNELS")
    denom = addresses.ibc_denom(path=f"{config.TRANSFER_PORT}/{spec.channel}/{spec.base}")
    try:
        trace = chainio.rest_get(path=f"/ibc/apps/transfer/v1/denoms/{denom.removeprefix(config.IBC_PREFIX)}")["denom"]
    except chainio.NotFound:
        raise DenomError(f"{spec.base} over {spec.channel}: {denom} has no denom trace on chain") from None
    outer = trace["trace"][0]
    if outer["port_id"] != config.TRANSFER_PORT or outer["channel_id"] != spec.channel:
        raise DenomError(f"{denom}: outermost hop is {outer['port_id']}/{outer['channel_id']}, expected transfer/{spec.channel}")
    return SweepDenom(denom=denom, symbol=spec.symbol, decimals=spec.decimals,
                      price_usd=config.NATIVE_PRICES_USD[spec.price_symbol], destination=spec.chain_id, channel=spec.channel)


# ---- bulk reads


def read_balances(denoms: list[SweepDenom]) -> dict[str, dict[str, int]]:
    """Every holder of every sweep denom: address -> denom -> amount, zero amounts dropped."""
    balances: dict[str, dict[str, int]] = {}
    for sweep_denom in denoms:
        owners = chainio.rest_get_all_pages(path="/cosmos/bank/v1beta1/denom_owners_by_query", key="denom_owners",
                                            params={"denom": sweep_denom.denom})
        for owner in owners:
            amount = int(owner["balance"]["amount"])
            if amount > 0:
                balances.setdefault(owner["address"], {})[sweep_denom.denom] = amount
    return balances


def read_skip_inputs() -> SkipInputs:
    module_accounts = chainio.rest_get(path="/cosmos/auth/v1beta1/module_accounts")["accounts"]
    module_addresses = {account["base_account"]["address"] for account in module_accounts}
    module_addresses |= {addresses.module_address(name=name) for name in config.BLOCKED_MODULE_NAMES}

    channels = chainio.rest_get_all_pages(path="/ibc/core/channel/v1/channels", key="channels")
    escrows = {addresses.escrow_address(channel_id=channel["channel_id"]) for channel in channels if channel["port_id"] == config.TRANSFER_PORT}

    contracts: set[str] = set()
    for code in chainio.rest_get_all_pages(path="/cosmwasm/wasm/v1/code", key="code_infos"):
        listed = chainio.rest_get_all_pages(path=f"/cosmwasm/wasm/v1/code/{code['code_id']}/contracts", key="contracts")
        # Only a 20-byte contract can collide with a holder; 32-byte ones fail the first rule anyway
        contracts |= {address for address in listed if addresses.is_stride_address(address=address)}
    return SkipInputs(module_addresses=module_addresses, escrows=escrows, contracts=contracts)


def spendable_balances(address: str) -> dict[str, int]:
    coins = chainio.rest_get_all_pages(path=f"/cosmos/bank/v1beta1/spendable_balances/{address}", key="balances")
    return {coin["denom"]: int(coin["amount"]) for coin in coins}


def usd_value(balances: dict[str, int], denoms: dict[str, SweepDenom]) -> Decimal:
    return sum((Decimal(amount) / Decimal(10) ** denoms[denom].decimals * denoms[denom].price_usd
                for denom, amount in balances.items() if denom in denoms), Decimal(0))


# ---- account cache


class AccountCache:
    """Auth lookups keyed by address. A type never changes, so a cached entry is final, except a keyless-looking one
    (no pubkey), which may have signed since and is re-read on every lookup."""

    def __init__(self, entries: dict[str, AccountInfo]) -> None:
        self._entries = entries

    @classmethod
    def load(cls, path: pathlib.Path = config.ACCOUNTS_CACHE_PATH) -> "AccountCache":
        if not path.exists():
            return cls(entries={})
        raw = json.loads(path.read_text())
        return cls(entries={address: AccountInfo(type=e["type"], has_pubkey=e["has_pubkey"], sequence=int(e["sequence"])) for address, e in raw.items()})

    def save(self, path: pathlib.Path = config.ACCOUNTS_CACHE_PATH) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        raw = {address: {"type": e.type, "has_pubkey": e.has_pubkey, "sequence": str(e.sequence)} for address, e in sorted(self._entries.items())}
        tmp = path.with_suffix(".tmp")
        tmp.write_text(json.dumps(raw, indent=1))
        tmp.replace(path)

    def lookup(self, address: str) -> AccountInfo | None:
        cached = self._entries.get(address)
        if cached is not None and cached.has_pubkey:
            return cached
        try:
            body = chainio.rest_get(path=f"/cosmos/auth/v1beta1/accounts/{address}")
        except chainio.NotFound:
            return None
        info = account_info(account=body["account"])
        self._entries[address] = info
        return info


def account_info(account: dict) -> AccountInfo:
    """Vesting, module and interchain accounts nest the base account."""
    base = account.get("base_account") or account.get("base_vesting_account", {}).get("base_account") or account
    return AccountInfo(type=account["@type"], has_pubkey=bool(base.get("pub_key")), sequence=int(base.get("sequence", 0)))


# ---- classification


def skip_reason(address: str, account: AccountInfo | None, inputs: SkipInputs, exclusions: dict[str, Exclusion]) -> str | None:
    """The chain's rules in the chain's order, then the builder's two. None means sweepable."""
    if not addresses.is_stride_address(address=address):
        return "address is not 20 bytes"
    if address in config.PROTOCOL_ADDRESSES or address == config.SWEEP_OPERATOR:
        return "protocol address"
    if address in inputs.module_addresses:
        return "blocked module address"
    if address in inputs.escrows:
        return "transfer escrow address"
    if account is None:
        return "account not found"
    if account.type == config.INTERCHAIN_ACCOUNT:
        return "interchain account"
    if account.type not in config.SWEEPABLE_ACCOUNT_TYPES:
        return f"account type {account.type} is not sweepable"
    if address in inputs.contracts:
        return "wasm contract address"
    exclusion = exclusions.get(address)
    if exclusion is not None:
        return f"excluded: {exclusion.section}: {exclusion.label}"
    return None


def classify(
    denoms: list[SweepDenom],
    balances: dict[str, dict[str, int]],
    floor_usd: Decimal,
    exclusions: dict[str, Exclusion],
    inputs: SkipInputs,
    accounts: AccountCache,
    test_address: str | None,
    height: int,
) -> HolderSet:
    by_denom = {sweep_denom.denom: sweep_denom for sweep_denom in denoms}
    holder_set = HolderSet(height=height, denoms=denoms, holders=[], excluded=[], skipped=[], below_floor=[])

    for address, coins in balances.items():
        if test_address is not None and address != test_address:
            continue
        usd = usd_value(balances=coins, denoms=by_denom)
        if test_address is None and usd < floor_usd:
            holder_set.below_floor.append(Holder(address=address, balances=coins, usd=usd, keyless=False))
            continue

        # Only a candidate above the floor costs a per-address read; the skip rules need its account
        account = accounts.lookup(address=address) if _skip_needs_account(address=address, inputs=inputs) else None
        reason = skip_reason(address=address, account=account, inputs=inputs, exclusions=exclusions)
        if reason is not None:
            target = holder_set.excluded if reason.startswith("excluded:") else holder_set.skipped
            target.append(Excluded(address=address, reason=reason, usd=usd))
            continue
        assert account is not None  # skip_reason returned None, so the account was found

        # The chain moves SpendableCoin: a vesting account's locked part stays behind
        spendable = coins if account.type not in config.VESTING_ACCOUNT_TYPES else _spendable_part(address=address, coins=coins)
        if not spendable:
            continue
        holder_set.holders.append(Holder(address=address, balances=spendable, usd=usd_value(balances=spendable, denoms=by_denom),
                                         keyless=not account.has_pubkey and account.sequence == 0))

    holder_set.holders.sort(key=lambda holder: holder.usd, reverse=True)
    holder_set.below_floor.sort(key=lambda holder: holder.usd, reverse=True)
    return holder_set


def _skip_needs_account(address: str, inputs: SkipInputs) -> bool:
    """The rules before the account lookup decide without it; skip the read when one of them already fires."""
    return (addresses.is_stride_address(address=address) and address not in config.PROTOCOL_ADDRESSES
            and address != config.SWEEP_OPERATOR and address not in inputs.module_addresses and address not in inputs.escrows)


def _spendable_part(address: str, coins: dict[str, int]) -> dict[str, int]:
    spendable = spendable_balances(address=address)
    return {denom: spendable.get(denom, 0) for denom in coins if spendable.get(denom, 0) > 0}


def read_holder_set(floor_usd: Decimal, exclusions: dict[str, Exclusion], accounts: AccountCache, test_address: str | None) -> HolderSet:
    height = chainio.latest_height()
    denoms = resolve_denoms()
    balances = read_balances(denoms=denoms)
    inputs = read_skip_inputs()
    return classify(denoms=denoms, balances=balances, floor_usd=floor_usd, exclusions=exclusions, inputs=inputs,
                    accounts=accounts, test_address=test_address, height=height)
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_holders -v`
Expected: 8 tests, OK. If `test_classify_applies_every_rule` disagrees on `MISSING`: it is a protocol address (the staketia redemption account) so "protocol address" is right, and the account lookup is skipped for it.

- [ ] **Step 5: Commit**

```bash
git add scripts/wind-down/sweep/holders.py scripts/wind-down/sweep/test_holders.py
git commit -m "wind-down sweep: live holder reads and the inclusion rules"
```

---

### Task 4: Planner: tiers, batches, plan file

**Files:**
- Create: `scripts/wind-down/sweep/planner.py`
- Test: `scripts/wind-down/sweep/test_planner.py`

**Interfaces:**
- Consumes: `holders.HolderSet`, `holders.Holder`, `holders.SweepDenom`, `holders.Excluded`; `ledger.Event`, `ledger.calibration`, `ledger.latest_run_id`, `ledger.unresolved`; `config.*`.
- Produces:
  - `class TierName(StrEnum): TEST = "test", CANARY = "canary", MAIN = "main", KEYLESS = "keyless"`
  - `@dataclass PlannedAddress(address: str, balances: dict[str, int], usd: Decimal, transfers: int)`
  - `@dataclass Batch(id: str, file: str, sha256: str, addresses: list[PlannedAddress], usd: Decimal, transfers: int, estimated_gas: int)`
  - `@dataclass Tier(name: TierName, batches: list[Batch])`
  - `@dataclass LadderRung(floor: Decimal, holders: int, usd: Decimal)`
  - `@dataclass Plan(run_id: int, created_at: str, height: int, floor_usd: Decimal, test: bool, canary: int, gas_per_transfer: int, prices: dict[str, Decimal], denoms: list[SweepDenom], tiers: list[Tier], excluded: list[Excluded], skipped: list[Excluded], below_floor_count: int, below_floor_usd: Decimal, ladder: list[LadderRung])` with `pending_batches() -> list[tuple[Tier, Batch]]` (tier order) and `batch(batch_id) -> Batch | None`
  - `class PlanError(Exception)`
  - `build_plan(holder_set, floor_usd, run_id, canary, test, gas_per_transfer, max_addresses, gas_budget, created_at) -> Plan`
  - `pack(holders: list[Holder], tier: TierName, run_id: int, first_index: int, gas_per_transfer, max_addresses, gas_budget) -> list[Batch]`
  - `ladder(holders: list[Holder], floor_usd: Decimal) -> list[LadderRung]` (the plan's floor then `config.LADDER_FLOORS` below it)
  - `write_plan(plan: Plan, state_dir: pathlib.Path = config.STATE_DIR) -> None` (plan.json atomically, batch files, removes stale `batch-*.txt`)
  - `load_plan(path: pathlib.Path = config.PLAN_PATH) -> Plan | None`
  - `plan_to_dict(plan) -> dict`, `plan_from_dict(data) -> Plan`
  - `batch_file_content(batch) -> str` (one address per line, trailing newline), `batch_sha256(content: str) -> str`
  - `next_run_id(events: list[ledger.Event]) -> int`, `gas_per_transfer(events) -> int`
- Review: yes (decides the order and size of txs that move user funds)

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/sweep/test_planner.py`:

```python
"""Tier membership and order, batch packing by addresses and by gas, the plan file round trip, the ladder."""

import json
import pathlib
import tempfile
import unittest
from decimal import Decimal

from sweep import config, holders, ledger, planner

DENOMS = [
    holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5"),
    holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5"),
]


def holder(index: int, usd: int, keyless: bool = False, denoms: int = 1) -> holders.Holder:
    balances = {"stuatom": usd * 1_000_000 // 6} if denoms == 1 else {"stuatom": usd * 1_000_000 // 12, "ustrd": usd * 10_000_000}
    return holders.Holder(address=f"stride1holder{index:04d}", balances=balances, usd=Decimal(usd), keyless=keyless)


def holder_set(holder_list: list[holders.Holder], below: list[holders.Holder] | None = None) -> holders.HolderSet:
    return holders.HolderSet(height=100, denoms=DENOMS, holders=sorted(holder_list, key=lambda h: h.usd, reverse=True),
                             excluded=[holders.Excluded(address="stride1excluded", reason="excluded: team: F5", usd=Decimal(9))],
                             skipped=[holders.Excluded(address="stride1escrow", reason="transfer escrow address", usd=Decimal(99))],
                             below_floor=below or [])


class TierTests(unittest.TestCase):
    def test_canary_is_the_smallest_then_main_by_value_then_keyless_last(self) -> None:
        hs = holder_set([holder(1, 500), holder(2, 50), holder(3, 5000), holder(4, 20), holder(5, 900, keyless=True), holder(6, 30)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=2, canary=2, test=False, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        tiers = {tier.name: [a.address for b in tier.batches for a in b.addresses] for tier in plan.tiers}
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.CANARY, planner.TierName.MAIN, planner.TierName.KEYLESS])
        self.assertEqual(tiers[planner.TierName.CANARY], ["stride1holder0004", "stride1holder0006"])
        self.assertEqual(tiers[planner.TierName.MAIN], ["stride1holder0003", "stride1holder0001", "stride1holder0002"])
        self.assertEqual(tiers[planner.TierName.KEYLESS], ["stride1holder0005"])
        self.assertEqual([b.id for tier in plan.tiers for b in tier.batches], ["002-001", "002-002", "002-003"])

    def test_test_plan_has_one_tier_with_one_address(self) -> None:
        hs = holder_set([holder(1, 0)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(100), run_id=1, canary=3, test=True, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.TEST])
        self.assertEqual(plan.tiers[0].batches[0].id, "001-001")
        self.assertTrue(plan.test)

    def test_empty_tiers_are_omitted(self) -> None:
        plan = planner.build_plan(holder_set=holder_set([holder(1, 50)]), floor_usd=Decimal(10), run_id=1, canary=0, test=False,
                                  gas_per_transfer=100_000, max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.MAIN])


class PackingTests(unittest.TestCase):
    def test_splits_at_max_addresses(self) -> None:
        batches = planner.pack(holders=[holder(i, 10) for i in range(7)], tier=planner.TierName.MAIN, run_id=1, first_index=1,
                               gas_per_transfer=100_000, max_addresses=3, gas_budget=40_000_000)
        self.assertEqual([len(b.addresses) for b in batches], [3, 3, 1])
        self.assertEqual([b.id for b in batches], ["001-001", "001-002", "001-003"])
        self.assertEqual(batches[0].estimated_gas, 300_000)
        self.assertEqual(batches[0].usd, Decimal(30))

    def test_splits_at_the_gas_budget_counting_transfers_per_holder(self) -> None:
        # each holder has two denoms = two transfers = 200,000 gas; a 500,000 budget fits two holders
        batches = planner.pack(holders=[holder(i, 12, denoms=2) for i in range(5)], tier=planner.TierName.MAIN, run_id=1, first_index=4,
                               gas_per_transfer=100_000, max_addresses=100, gas_budget=500_000)
        self.assertEqual([len(b.addresses) for b in batches], [2, 2, 1])
        self.assertEqual(batches[0].transfers, 4)
        self.assertEqual(batches[0].id, "001-004")

    def test_a_single_holder_over_budget_still_gets_a_batch(self) -> None:
        batches = planner.pack(holders=[holder(1, 12, denoms=2)], tier=planner.TierName.MAIN, run_id=1, first_index=1,
                               gas_per_transfer=100_000, max_addresses=100, gas_budget=50_000)
        self.assertEqual(len(batches), 1)


class LadderAndCalibrationTests(unittest.TestCase):
    def test_ladder_counts_holders_at_each_floor(self) -> None:
        rungs = planner.ladder(holders=[holder(1, 50), holder(2, 7), holder(3, 3), holder(4, 0)], floor_usd=Decimal(25))
        self.assertEqual([(r.floor, r.holders, r.usd) for r in rungs], [
            (Decimal(25), 1, Decimal(50)), (Decimal(10), 1, Decimal(50)), (Decimal(5), 2, Decimal(57)),
            (Decimal(1), 3, Decimal(60)), (Decimal(0), 4, Decimal(60)),
        ])

    def test_gas_per_transfer_falls_back_to_the_default(self) -> None:
        self.assertEqual(planner.gas_per_transfer(events=[]), config.GAS_PER_TRANSFER_DEFAULT)
        t = ledger.Transfer(address="a", denom="d", amount=1, channel="c", receiver="r")
        events = [ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=100_000, transfers=[t], skipped=[])]
        self.assertEqual(planner.gas_per_transfer(events=events), 120_000)
        self.assertEqual(planner.next_run_id(events=[ledger.submitted(run_id=4, batch_id="004-001", tx_hash="A", at="t", addresses=1, transfers_estimate=1, gas_wanted=1)]), 5)


class PlanFileTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_write_then_load_round_trips_and_replaces_old_batch_files(self) -> None:
        (self.state / "batch-000-009.txt").write_text("stale\n")
        hs = holder_set([holder(1, 500), holder(2, 50)], below=[holder(9, 3)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                                  max_addresses=1, gas_budget=40_000_000, created_at="2026-11-10T09:00:00+00:00")
        planner.write_plan(plan=plan, state_dir=self.state)

        self.assertFalse((self.state / "batch-000-009.txt").exists())
        self.assertEqual((self.state / "batch-001-001.txt").read_text(), "stride1holder0001\n")
        loaded = planner.load_plan(path=self.state / "plan.json")
        self.assertEqual(loaded, plan)
        self.assertEqual(loaded.batch(batch_id="001-002").sha256, planner.batch_sha256(content="stride1holder0002\n"))
        self.assertEqual(loaded.below_floor_count, 1)
        self.assertEqual(loaded.below_floor_usd, Decimal(3))
        raw = json.loads((self.state / "plan.json").read_text())
        self.assertEqual(raw["tiers"][0]["batches"][0]["addresses"][0]["balances"]["stuatom"], str(500 * 1_000_000 // 6))
        self.assertIsNone(planner.load_plan(path=self.state / "missing.json"))

    def test_pending_batches_follow_tier_order(self) -> None:
        hs = holder_set([holder(1, 500), holder(2, 50, keyless=True), holder(3, 20)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=1, canary=1, test=False, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([(tier.name, batch.id) for tier, batch in plan.pending_batches()],
                         [(planner.TierName.CANARY, "001-001"), (planner.TierName.MAIN, "001-002"), (planner.TierName.KEYLESS, "001-003")])


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_planner -v`
Expected: `ModuleNotFoundError: No module named 'sweep.planner'`.

- [ ] **Step 3: Write planner.py**

```python
"""From a HolderSet to a plan: tiers in sweep order, batches packed under the address and gas limits, and the plan
file plus one address file per batch (the CLI's input). The plan is overwritten by every `plan`; the ledger never is."""

import dataclasses
import hashlib
import json
import pathlib
from dataclasses import dataclass
from decimal import Decimal
from enum import StrEnum

from sweep import config, holders, ledger

BATCH_FILE_PREFIX = "batch-"
BATCH_FILE_SUFFIX = ".txt"


class TierName(StrEnum):
    TEST = "test"
    CANARY = "canary"
    MAIN = "main"
    KEYLESS = "keyless"


class PlanError(Exception):
    """The plan file cannot be read."""


@dataclass(frozen=True)
class PlannedAddress:
    address: str
    balances: dict[str, int]
    usd: Decimal
    transfers: int


@dataclass(frozen=True)
class Batch:
    id: str
    file: str
    sha256: str
    addresses: list[PlannedAddress]
    usd: Decimal
    transfers: int
    estimated_gas: int


@dataclass(frozen=True)
class Tier:
    name: TierName
    batches: list[Batch]


@dataclass(frozen=True)
class LadderRung:
    floor: Decimal
    holders: int
    usd: Decimal


@dataclass(frozen=True)
class Plan:
    run_id: int
    created_at: str
    height: int
    floor_usd: Decimal
    test: bool
    canary: int
    gas_per_transfer: int
    prices: dict[str, Decimal]
    denoms: list[holders.SweepDenom]
    tiers: list[Tier]
    excluded: list[holders.Excluded]
    skipped: list[holders.Excluded]
    below_floor_count: int
    below_floor_usd: Decimal
    ladder: list[LadderRung]

    def pending_batches(self) -> list[tuple[Tier, Batch]]:
        return [(tier, batch) for tier in self.tiers for batch in tier.batches]

    def batch(self, batch_id: str) -> Batch | None:
        return next((batch for _, batch in self.pending_batches() if batch.id == batch_id), None)


# ---- building


def build_plan(
    holder_set: holders.HolderSet,
    floor_usd: Decimal,
    run_id: int,
    canary: int,
    test: bool,
    gas_per_transfer: int,
    max_addresses: int,
    gas_budget: int,
    created_at: str,
) -> Plan:
    members = _tier_members(holder_list=holder_set.holders, canary=canary, test=test)

    # Batch numbers run across tiers so an id names one batch of this run whatever its tier
    tiers: list[Tier] = []
    index = 1
    for name, tier_holders in members:
        batches = pack(holders=tier_holders, tier=name, run_id=run_id, first_index=index, gas_per_transfer=gas_per_transfer,
                       max_addresses=max_addresses, gas_budget=gas_budget)
        if not batches:
            continue
        tiers.append(Tier(name=name, batches=batches))
        index += len(batches)

    return Plan(
        run_id=run_id, created_at=created_at, height=holder_set.height, floor_usd=floor_usd, test=test, canary=canary,
        gas_per_transfer=gas_per_transfer, prices={d.denom: d.price_usd for d in holder_set.denoms}, denoms=list(holder_set.denoms),
        tiers=tiers, excluded=list(holder_set.excluded), skipped=list(holder_set.skipped),
        below_floor_count=len(holder_set.below_floor), below_floor_usd=sum((h.usd for h in holder_set.below_floor), Decimal(0)),
        ladder=ladder(holders=holder_set.holders + holder_set.below_floor, floor_usd=floor_usd),
    )


def _tier_members(holder_list: list[holders.Holder], canary: int, test: bool) -> list[tuple[TierName, list[holders.Holder]]]:
    """Who goes in which tier, in sweep order. `holder_list` arrives sorted by USD descending."""
    if test:
        return [(TierName.TEST, list(holder_list))]
    keyed = [h for h in holder_list if not h.keyless]
    keyless = [h for h in holder_list if h.keyless]
    canaries = sorted(keyed, key=lambda h: h.usd)[:canary]
    canary_addresses = {h.address for h in canaries}
    main = [h for h in keyed if h.address not in canary_addresses]
    return [(TierName.CANARY, canaries), (TierName.MAIN, main), (TierName.KEYLESS, keyless)]


def pack(
    holders: list[holders.Holder],
    tier: TierName,
    run_id: int,
    first_index: int,
    gas_per_transfer: int,
    max_addresses: int,
    gas_budget: int,
) -> list[Batch]:
    """Walk the holders in order; open a new batch when the next one would break the address or gas limit. A single
    holder over the budget still gets its own batch (the runner's simulation is the real guard)."""
    batches: list[Batch] = []
    current: list[holders.Holder] = []
    current_gas = 0

    def close() -> None:
        if current:
            batches.append(_batch(members=current, run_id=run_id, index=first_index + len(batches), gas_per_transfer=gas_per_transfer))

    for holder in holders:
        gas = holder.transfers * gas_per_transfer
        if current and (len(current) >= max_addresses or current_gas + gas > gas_budget):
            close()
            current, current_gas = [], 0
        current.append(holder)
        current_gas += gas
    close()
    return batches


def _batch(members: list[holders.Holder], run_id: int, index: int, gas_per_transfer: int) -> Batch:
    batch_id = f"{run_id:03d}-{index:03d}"
    planned = [PlannedAddress(address=h.address, balances=dict(h.balances), usd=h.usd, transfers=h.transfers) for h in members]
    content = "".join(f"{p.address}\n" for p in planned)
    transfers = sum(p.transfers for p in planned)
    return Batch(id=batch_id, file=f"{BATCH_FILE_PREFIX}{batch_id}{BATCH_FILE_SUFFIX}", sha256=batch_sha256(content=content),
                 addresses=planned, usd=sum((p.usd for p in planned), Decimal(0)), transfers=transfers,
                 estimated_gas=transfers * gas_per_transfer)


def ladder(holders: list[holders.Holder], floor_usd: Decimal) -> list[LadderRung]:
    floors = [floor_usd] + [f for f in config.LADDER_FLOORS if f < floor_usd]
    return [LadderRung(floor=f, holders=sum(1 for h in holders if h.usd >= f), usd=sum((h.usd for h in holders if h.usd >= f), Decimal(0)))
            for f in floors]


def gas_per_transfer(events: list[ledger.Event]) -> int:
    return ledger.calibration(events=events) or config.GAS_PER_TRANSFER_DEFAULT


def next_run_id(events: list[ledger.Event]) -> int:
    return ledger.latest_run_id(events=events) + 1


# ---- files


def batch_file_content(batch: Batch) -> str:
    return "".join(f"{p.address}\n" for p in batch.addresses)


def batch_sha256(content: str) -> str:
    return hashlib.sha256(content.encode()).hexdigest()


def write_plan(plan: Plan, state_dir: pathlib.Path = config.STATE_DIR) -> None:
    state_dir.mkdir(parents=True, exist_ok=True)
    for stale in state_dir.glob(f"{BATCH_FILE_PREFIX}*{BATCH_FILE_SUFFIX}"):
        stale.unlink()
    for _, batch in plan.pending_batches():
        (state_dir / batch.file).write_text(batch_file_content(batch=batch))
    tmp = state_dir / "plan.json.tmp"
    tmp.write_text(json.dumps(plan_to_dict(plan=plan), indent=1))
    tmp.replace(state_dir / "plan.json")


def load_plan(path: pathlib.Path = config.PLAN_PATH) -> Plan | None:
    if not path.exists():
        return None
    try:
        return plan_from_dict(data=json.loads(path.read_text()))
    except (json.JSONDecodeError, KeyError, ValueError) as error:
        raise PlanError(f"{path}: {error}") from None


def plan_to_dict(plan: Plan) -> dict:
    return {
        "run_id": str(plan.run_id), "created_at": plan.created_at, "height": str(plan.height), "floor_usd": str(plan.floor_usd),
        "test": plan.test, "canary": str(plan.canary), "gas_per_transfer": str(plan.gas_per_transfer),
        "prices": {denom: str(price) for denom, price in plan.prices.items()},
        "denoms": [{"denom": d.denom, "symbol": d.symbol, "decimals": str(d.decimals), "price_usd": str(d.price_usd),
                    "destination": d.destination, "channel": d.channel} for d in plan.denoms],
        "tiers": [{"name": tier.name.value, "batches": [_batch_to_dict(batch=b) for b in tier.batches]} for tier in plan.tiers],
        "excluded": [_excluded_to_dict(entry=e) for e in plan.excluded],
        "skipped": [_excluded_to_dict(entry=e) for e in plan.skipped],
        "below_floor": {"count": str(plan.below_floor_count), "usd": str(plan.below_floor_usd)},
        "ladder": [{"floor": str(r.floor), "holders": str(r.holders), "usd": str(r.usd)} for r in plan.ladder],
    }


def _batch_to_dict(batch: Batch) -> dict:
    return {
        "id": batch.id, "file": batch.file, "sha256": batch.sha256, "usd": str(batch.usd), "transfers": str(batch.transfers),
        "estimated_gas": str(batch.estimated_gas),
        "addresses": [{"address": p.address, "balances": {d: str(a) for d, a in p.balances.items()}, "usd": str(p.usd),
                       "transfers": str(p.transfers)} for p in batch.addresses],
    }


def _excluded_to_dict(entry: holders.Excluded) -> dict:
    return {"address": entry.address, "reason": entry.reason, "usd": str(entry.usd)}


def plan_from_dict(data: dict) -> Plan:
    return Plan(
        run_id=int(data["run_id"]), created_at=data["created_at"], height=int(data["height"]), floor_usd=Decimal(data["floor_usd"]),
        test=bool(data["test"]), canary=int(data["canary"]), gas_per_transfer=int(data["gas_per_transfer"]),
        prices={denom: Decimal(price) for denom, price in data["prices"].items()},
        denoms=[holders.SweepDenom(denom=d["denom"], symbol=d["symbol"], decimals=int(d["decimals"]), price_usd=Decimal(d["price_usd"]),
                                   destination=d["destination"], channel=d["channel"]) for d in data["denoms"]],
        tiers=[Tier(name=TierName(t["name"]), batches=[_batch_from_dict(data=b) for b in t["batches"]]) for t in data["tiers"]],
        excluded=[_excluded_from_dict(data=e) for e in data["excluded"]],
        skipped=[_excluded_from_dict(data=e) for e in data["skipped"]],
        below_floor_count=int(data["below_floor"]["count"]), below_floor_usd=Decimal(data["below_floor"]["usd"]),
        ladder=[LadderRung(floor=Decimal(r["floor"]), holders=int(r["holders"]), usd=Decimal(r["usd"])) for r in data["ladder"]],
    )


def _batch_from_dict(data: dict) -> Batch:
    return Batch(
        id=data["id"], file=data["file"], sha256=data["sha256"], usd=Decimal(data["usd"]), transfers=int(data["transfers"]),
        estimated_gas=int(data["estimated_gas"]),
        addresses=[PlannedAddress(address=p["address"], balances={d: int(a) for d, a in p["balances"].items()}, usd=Decimal(p["usd"]),
                                  transfers=int(p["transfers"])) for p in data["addresses"]],
    )


def _excluded_from_dict(data: dict) -> holders.Excluded:
    return holders.Excluded(address=data["address"], reason=data["reason"], usd=Decimal(data["usd"]))
```

Note: `dataclasses` is imported for nothing in the code above; drop the import. The parameter named `holders` in `pack` and `ladder` shadows the module inside those functions; rename the parameter to `holder_list` in both (and in the tests' keyword calls: `planner.pack(holder_list=…)`, `planner.ladder(holder_list=…)`), and use `holders.Holder` for the type annotation. Apply that rename consistently in the test file too.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_planner -v`
Expected: 10 tests, OK.

- [ ] **Step 5: Commit**

```bash
git add scripts/wind-down/sweep/planner.py scripts/wind-down/sweep/test_planner.py
git commit -m "wind-down sweep: the planner (tiers, gas-budgeted batches, plan file)"
```

---

## Parallel-safe tasks

Tasks 5, 6 and 7 depend only on Tasks 1-4 and never on each other. Task 7 deletes the old builder and edits the ops plan; Task 6 edits dashboard files; Task 5 adds the CLI. The only shared file is `scripts/wind-down/dashboard/README.md` (Task 6 adds a section, Task 7 does not touch it), so there is no textual overlap.

### Task 5: The CLI: plan, run, status, resolve

**Files:**
- Create: `scripts/wind-down/sweep/cli.py`, `scripts/wind-down/sweep/README.md`
- Test: `scripts/wind-down/sweep/test_cli.py`

**Interfaces:**
- Consumes: `planner.Plan`, `planner.Batch`, `planner.Tier`, `planner.TierName`, `planner.build_plan`, `planner.write_plan`, `planner.load_plan`, `planner.batch_file_content`, `planner.batch_sha256`, `planner.gas_per_transfer`, `planner.next_run_id`; `holders.load_exclusions`, `holders.AccountCache`, `holders.read_holder_set`, `holders.Exclusion`, `holders.DenomError`, `holders.ExclusionsError`; `ledger.*`; `chainio.strided`, `chainio.rest_get`, `chainio.NotFound`, `chainio.ChainError`, `chainio.now`, `chainio.sleep`; `config.*`.
- Depends on: Tasks 1-4
- Produces (module functions, all tested):
  - `@dataclass Check(name: str, ok: bool, detail: str)`
  - `preflight(plan: Plan, events: list[ledger.Event], exclusions: dict[str, Exclusion], now: datetime) -> list[Check]`
  - `command_line(denoms: list[str], file: pathlib.Path, gas: int | None, dry_run: bool) -> list[str]`
  - `parse_gas_estimate(text: str) -> int | None`
  - `@dataclass BroadcastResult(code: int, tx_hash: str, codespace: str, raw_log: str)`, `parse_broadcast(stdout: str) -> BroadcastResult`
  - `@dataclass TxOutcome(code: int, height: int, gas_used: int, codespace: str, raw_log: str, transfers: list[ledger.Transfer], skipped: list[ledger.Skip])`, `parse_tx_response(body: dict) -> TxOutcome`
  - `poll_tx(tx_hash: str, wait_seconds: int) -> TxOutcome | None`
  - `run_batch(plan, tier, batch, state_dir, ledger_path, yes, dry_run) -> ledger.Event | None` (the event appended, None on dry run or a declined prompt)
  - `ask(prompt: str) -> str` (wraps `input`; tests patch it)
  - `main(argv: list[str] | None = None) -> int`
- Review: yes (signs and broadcasts the tx that moves user funds)

- [ ] **Step 1: Write the failing tests**

`scripts/wind-down/sweep/test_cli.py`:

```python
"""The runner: preflight rows, the exact strided command line, gas-estimate parsing, the broadcast and poll loop over
fixture responses, and the stop conditions. chainio is patched throughout."""

import datetime
import json
import pathlib
import tempfile
import unittest
from decimal import Decimal
from unittest import mock

from sweep import chainio, cli, config, holders, ledger, planner

DENOMS = [holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5"),
          holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5")]
A1 = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
A2 = "stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c"
NOW = datetime.datetime(2026, 11, 10, 9, 0, tzinfo=datetime.UTC)


def make_plan(created_at: str = NOW.isoformat(), addresses: tuple[str, ...] = (A1, A2)) -> planner.Plan:
    holder_list = [holders.Holder(address=a, balances={"stuatom": 1_000_000, "ustrd": 10}, usd=Decimal(6), keyless=False) for a in addresses]
    hs = holders.HolderSet(height=100, denoms=DENOMS, holders=holder_list, excluded=[], skipped=[], below_floor=[])
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(1), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=1, gas_budget=40_000_000, created_at=created_at)


TX_RESPONSE = {"tx_response": {
    "height": "4242", "txhash": "AB12", "code": 0, "codespace": "", "raw_log": "", "gas_wanted": "260000", "gas_used": "201000",
    "events": [
        {"type": "message", "attributes": [{"key": "action", "value": "/stride.stakeibc.MsgSweepTokensOffStride"}]},
        {"type": "sweep_transfer", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A1}, {"key": "denom", "value": "stuatom"},
            {"key": "amount", "value": "1000000"}, {"key": "channel", "value": "channel-5"}, {"key": "receiver", "value": "osmo1x"}]},
        {"type": "sweep_transfer", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A1}, {"key": "denom", "value": "ustrd"},
            {"key": "amount", "value": "10"}, {"key": "channel", "value": "channel-5"}, {"key": "receiver", "value": "osmo1x"}]},
        {"type": "sweep_skipped", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A2}, {"key": "reason", "value": "interchain account"}]},
    ],
}}


class ParsingTests(unittest.TestCase):
    def test_command_line_is_exactly_the_documented_shape(self) -> None:
        line = cli.command_line(denoms=["stuatom", "ustrd"], file=pathlib.Path("/s/batch-001-001.txt"), gas=260_000, dry_run=False)
        self.assertEqual(line, [
            "tx", "stakeibc", "sweep-tokens-off-stride", "stuatom,ustrd", "/s/batch-001-001.txt",
            "--from", "stride-sweeper", "--keyring-backend", "test", "--chain-id", "stride-1", "--node", config.RPC,
            "--gas", "260000", "--gas-prices", "0.001ustrd", "--broadcast-mode", "sync", "-y", "--output", "json",
        ])
        dry = cli.command_line(denoms=["stuatom"], file=pathlib.Path("/s/b.txt"), gas=None, dry_run=True)
        self.assertIn("--dry-run", dry)
        self.assertNotIn("--gas", dry)
        self.assertNotIn("-y", dry)

    def test_parse_gas_estimate(self) -> None:
        self.assertEqual(cli.parse_gas_estimate(text="some noise\ngas estimate: 201000\n"), 201_000)
        self.assertIsNone(cli.parse_gas_estimate(text="Error: rpc error"))

    def test_parse_broadcast(self) -> None:
        result = cli.parse_broadcast(stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}))
        self.assertEqual(result, cli.BroadcastResult(code=0, tx_hash="AB12", codespace="", raw_log=""))

    def test_parse_tx_response_reads_sweep_events(self) -> None:
        outcome = cli.parse_tx_response(body=TX_RESPONSE)
        self.assertEqual((outcome.code, outcome.height, outcome.gas_used), (0, 4242, 201_000))
        self.assertEqual(outcome.transfers, [
            ledger.Transfer(address=A1, denom="stuatom", amount=1_000_000, channel="channel-5", receiver="osmo1x"),
            ledger.Transfer(address=A1, denom="ustrd", amount=10, channel="channel-5", receiver="osmo1x"),
        ])
        self.assertEqual(outcome.skipped, [ledger.Skip(address=A2, reason="interchain account")])


class PreflightTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        self.plan = make_plan()
        planner.write_plan(plan=self.plan, state_dir=self.state)
        self.rest = {
            "/cosmos/base/tendermint/v1beta1/node_info": {"default_node_info": {"network": "stride-1"}},
            "/ibc/core/channel/v1/channels/channel-5/ports/transfer": {"channel": {"state": "STATE_OPEN"}},
            f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom": {"balance": {"denom": "ustrd", "amount": "5000000"}},
        }
        self.commands = {
            ("version",): chainio.CommandResult(returncode=0, stdout="v35.0.0\n", stderr=""),
            ("keys", "show", "stride-sweeper", "--keyring-backend", "test", "-a"): chainio.CommandResult(returncode=0, stdout=config.SWEEP_OPERATOR + "\n", stderr=""),
        }
        mock.patch.object(chainio, "rest_get", side_effect=lambda path, params=None: self.rest[path]).start()
        mock.patch.object(chainio, "strided", side_effect=lambda args: self.commands[tuple(args)]).start()
        mock.patch.object(config, "STATE_DIR", self.state).start()
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def checks(self, events: list[ledger.Event] | None = None, exclusions: dict | None = None, now: datetime.datetime = NOW) -> dict[str, cli.Check]:
        return {c.name: c for c in cli.preflight(plan=self.plan, events=events or [], exclusions=exclusions or {}, now=now)}

    def test_everything_passes_on_a_clean_setup(self) -> None:
        checks = self.checks()
        self.assertTrue(all(c.ok for c in checks.values()), [c for c in checks.values() if not c.ok])
        self.assertEqual(set(checks), {"chain id", "binary version", "operator key", "plan age", "batch files", "unresolved submissions",
                                       "exclusions and protocol addresses", "channels open", "operator fee balance"})

    def test_each_failure_mode(self) -> None:
        self.assertFalse(self.checks(now=NOW + datetime.timedelta(hours=7))["plan age"].ok)
        (self.state / "batch-001-001.txt").write_text(f"{A2}\n")
        self.assertFalse(self.checks()["batch files"].ok)
        planner.write_plan(plan=self.plan, state_dir=self.state)
        submitted = ledger.submitted(run_id=1, batch_id="001-001", tx_hash="X", at="t", addresses=1, transfers_estimate=2, gas_wanted=1)
        self.assertFalse(self.checks(events=[submitted])["unresolved submissions"].ok)
        exclusion = holders.Exclusion(address=A1, section="team", label="F5", reason="r")
        self.assertFalse(self.checks(exclusions={A1: exclusion})["exclusions and protocol addresses"].ok)
        self.rest["/ibc/core/channel/v1/channels/channel-5/ports/transfer"] = {"channel": {"state": "STATE_CLOSED"}}
        self.assertFalse(self.checks()["channels open"].ok)
        self.rest[f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom"] = {"balance": {"denom": "ustrd", "amount": "1"}}
        self.assertFalse(self.checks()["operator fee balance"].ok)
        self.commands[("version",)] = chainio.CommandResult(returncode=0, stdout="v34.0.0\n", stderr="")
        self.assertFalse(self.checks()["binary version"].ok)
        self.commands[("keys", "show", "stride-sweeper", "--keyring-backend", "test", "-a")] = chainio.CommandResult(returncode=0, stdout="stride1other\n", stderr="")
        self.assertFalse(self.checks()["operator key"].ok)

    def test_already_confirmed_batches_are_not_rechecked_for_exclusions(self) -> None:
        confirmed = [ledger.submitted(run_id=1, batch_id="001-001", tx_hash="X", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
                     ledger.confirmed(batch_id="001-001", tx_hash="X", at="t", height=1, gas_used=1, transfers=[], skipped=[])]
        exclusion = holders.Exclusion(address=A1, section="team", label="F5", reason="r")
        self.assertTrue(self.checks(events=confirmed, exclusions={A1: exclusion})["exclusions and protocol addresses"].ok)


class RunBatchTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        self.ledger_path = self.state / "ledger.jsonl"
        self.plan = make_plan()
        planner.write_plan(plan=self.plan, state_dir=self.state)
        self.tier, self.batch = self.plan.pending_batches()[0]
        self.strided_calls: list[list[str]] = []
        self.tx_lookups = 0
        mock.patch.object(chainio, "now", return_value=NOW).start()
        mock.patch.object(chainio, "sleep").start()
        mock.patch.object(cli, "ask", return_value="y").start()
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def fake_strided(self, args: list[str]) -> chainio.CommandResult:
        self.strided_calls.append(args)
        if "--dry-run" in args:
            return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
        return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")

    def fake_rest(self, path: str, params: dict | None = None) -> dict:
        self.tx_lookups += 1
        if self.tx_lookups < 3:
            raise chainio.NotFound(path)
        return TX_RESPONSE

    def test_simulate_broadcast_poll_and_record(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided), mock.patch.object(chainio, "rest_get", side_effect=self.fake_rest):
            event = cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=False, dry_run=False)

        self.assertEqual(event.kind, ledger.Kind.CONFIRMED)
        self.assertEqual(self.strided_calls[0][-1], "--dry-run")
        self.assertIn("--gas", self.strided_calls[1])
        self.assertEqual(self.strided_calls[1][self.strided_calls[1].index("--gas") + 1], "260000")  # 200,000 x 1.3
        events = ledger.read(path=self.ledger_path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.SUBMITTED, ledger.Kind.CONFIRMED])
        self.assertEqual(events[0].gas_wanted, 260_000)
        self.assertEqual(events[1].height, 4242)
        self.assertEqual(len(events[1].transfers), 2)
        self.assertEqual(events[1].skipped[0].address, A2)

    def test_dry_run_broadcasts_nothing(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided):
            self.assertIsNone(cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=True))
        self.assertEqual(len(self.strided_calls), 1)
        self.assertFalse(self.ledger_path.exists())

    def test_declined_prompt_signs_nothing(self) -> None:
        with mock.patch.object(cli, "ask", return_value="n"), mock.patch.object(chainio, "strided", side_effect=self.fake_strided):
            self.assertIsNone(cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=False, dry_run=False))
        self.assertEqual(self.strided_calls, [])

    def test_over_the_block_limit_refuses_the_batch(self) -> None:
        def huge(args: list[str]) -> chainio.CommandResult:
            return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 90000000\n")
        with mock.patch.object(chainio, "strided", side_effect=huge):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        self.assertFalse(self.ledger_path.exists())

    def test_rejected_broadcast_is_a_failed_event(self) -> None:
        def rejected(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "CD34", "codespace": "sdk", "code": 13, "raw_log": "insufficient fee"}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=rejected):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        events = ledger.read(path=self.ledger_path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.FAILED])
        self.assertEqual(events[0].code, 13)

    def test_unfound_tx_leaves_the_submission_unresolved(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided), \
             mock.patch.object(chainio, "rest_get", side_effect=chainio.NotFound("x")), \
             mock.patch.object(config, "TX_WAIT_SECONDS", 6):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        self.assertEqual([e.kind for e in ledger.read(path=self.ledger_path)], [ledger.Kind.SUBMITTED])


class RunLoopTests(unittest.TestCase):
    """`main(["run", ...])` over a plan with two single-address batches."""

    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        planner.write_plan(plan=make_plan(), state_dir=self.state)
        mock.patch.object(config, "STATE_DIR", self.state).start()
        mock.patch.object(config, "PLAN_PATH", self.state / "plan.json").start()
        mock.patch.object(config, "LEDGER_PATH", self.state / "ledger.jsonl").start()
        mock.patch.object(config, "EXCLUSIONS_PATH", self.state / "exclusions.json").start()
        (self.state / "exclusions.json").write_text('{"sections": []}')
        mock.patch.object(cli, "preflight", return_value=[cli.Check(name="all", ok=True, detail="stubbed")]).start()
        mock.patch.object(chainio, "now", return_value=NOW).start()
        mock.patch.object(chainio, "sleep").start()
        mock.patch.object(cli, "ask", return_value="a").start()
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def test_stops_after_a_batch_with_a_skip_event(self) -> None:
        def fake_strided(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=fake_strided), mock.patch.object(chainio, "rest_get", return_value=TX_RESPONSE):
            self.assertEqual(cli.main(argv=["run"]), 1)
        events = ledger.read(path=self.state / "ledger.jsonl")
        self.assertEqual([e.batch_id for e in events], ["001-001", "001-001"])  # the second batch was never started

    def test_continue_on_skip_runs_every_batch_and_batches_flag_limits(self) -> None:
        def fake_strided(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=fake_strided), mock.patch.object(chainio, "rest_get", return_value=TX_RESPONSE):
            self.assertEqual(cli.main(argv=["run", "--continue-on-skip", "--batches", "1"]), 0)
            self.assertEqual([e.batch_id for e in ledger.read(path=self.state / "ledger.jsonl")], ["001-001", "001-001"])
            self.assertEqual(cli.main(argv=["run", "--continue-on-skip"]), 0)
            self.assertEqual([e.batch_id for e in ledger.read(path=self.state / "ledger.jsonl")], ["001-001", "001-001", "001-002", "001-002"])

    def test_status_needs_no_network(self) -> None:
        with mock.patch.object(chainio, "rest_get", side_effect=AssertionError("network")), mock.patch.object(chainio, "strided", side_effect=AssertionError("subprocess")):
            self.assertEqual(cli.main(argv=["status"]), 0)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_cli -v`
Expected: `ModuleNotFoundError: No module named 'sweep.cli'`.

- [ ] **Step 3: Write cli.py**

```python
#!/usr/bin/env python3
"""The sweep runner.

    python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 [--canary 3] [--test] [--max-addresses 100] [--gas-budget 40000000]
    python3 scripts/wind-down/sweep/cli.py run [--tier test|canary|main|keyless] [--batches N] [--yes] [--dry-run] [--continue-on-skip]
    python3 scripts/wind-down/sweep/cli.py status
    python3 scripts/wind-down/sweep/cli.py resolve [--wait-seconds 600]

`plan` reads live holder state and writes state/plan.json plus one address file per batch. `run` walks the plan's
batches in tier order: preflight, then per batch a gas simulation, the signed broadcast from the stride-sweeper key,
and a poll for the tx result, each step appended to state/ledger.jsonl. It stops on any skip event, failed tx, or a
batch over the block gas limit. `status` reads the two files. `resolve` finishes a submission the poll gave up on.
"""

import argparse
import datetime
import json
import pathlib
import re
import sys
from dataclasses import dataclass
from decimal import Decimal

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # so `python3 sweep/cli.py` finds the package

from sweep import chainio, config, holders, ledger, planner  # noqa: E402

GAS_ESTIMATE_PATTERN = re.compile(r"gas estimate:\s*(\d+)")
EVENT_SWEEP_TRANSFER = "sweep_transfer"
EVENT_SWEEP_SKIPPED = "sweep_skipped"
STATE_OPEN = "STATE_OPEN"
ANSWER_YES = "y"
ANSWER_ALL = "a"


class RunStopped(Exception):
    """The run must not continue; the message says why."""


@dataclass(frozen=True)
class Check:
    name: str
    ok: bool
    detail: str


@dataclass(frozen=True)
class BroadcastResult:
    code: int
    tx_hash: str
    codespace: str
    raw_log: str


@dataclass(frozen=True)
class TxOutcome:
    code: int
    height: int
    gas_used: int
    codespace: str
    raw_log: str
    transfers: list[ledger.Transfer]
    skipped: list[ledger.Skip]


# ---- entry point


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv=argv)
    try:
        return args.command(args)
    except (RunStopped, holders.DenomError, holders.ExclusionsError, planner.PlanError, ledger.LedgerError, chainio.ChainError) as error:
        print(f"RESULT: FAIL — {error}")
        return 1


def parse_args(argv: list[str] | None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    commands = parser.add_subparsers(dest="command_name", required=True)

    plan = commands.add_parser("plan", help="read live state and write the batches")
    plan.add_argument("--floor-usd", type=Decimal, default=None, help="sweep every holder at or above this value (required unless --test)")
    plan.add_argument("--canary", type=int, default=0, help="this many of the smallest holders above the floor go first")
    plan.add_argument("--test", action="store_true", help=f"plan only the test address {config.TEST_ADDRESS}")
    plan.add_argument("--max-addresses", type=int, default=config.MAX_ADDRESSES_PER_BATCH)
    plan.add_argument("--gas-budget", type=int, default=config.GAS_BUDGET_PER_BATCH)
    plan.set_defaults(command=cmd_plan)

    run = commands.add_parser("run", help="sign and submit the pending batches in tier order")
    run.add_argument("--tier", type=planner.TierName, choices=list(planner.TierName), default=None)
    run.add_argument("--batches", type=int, default=None, help="stop after this many batches")
    run.add_argument("--yes", action="store_true", help="do not prompt per batch")
    run.add_argument("--dry-run", action="store_true", help="simulate and print the commands; broadcast nothing")
    run.add_argument("--continue-on-skip", action="store_true", help="keep going after a batch with sweep_skipped events")
    run.set_defaults(command=cmd_run)

    status = commands.add_parser("status", help="what the plan and ledger say; no network")
    status.set_defaults(command=cmd_status)

    resolve = commands.add_parser("resolve", help="finish submissions the poll gave up on")
    resolve.add_argument("--wait-seconds", type=int, default=config.RESOLVE_WAIT_SECONDS)
    resolve.set_defaults(command=cmd_resolve)
    return parser.parse_args(argv)


# ---- plan


def cmd_plan(args: argparse.Namespace) -> int:
    if not args.test and args.floor_usd is None:
        raise RunStopped("--floor-usd is required unless --test")
    events = ledger.read()
    pending = ledger.unresolved(events=events)
    if pending:
        raise RunStopped(f"{len(pending)} submitted batch(es) unresolved ({', '.join(e.batch_id for e in pending)}); run `resolve` first")

    exclusions = holders.load_exclusions()
    accounts = holders.AccountCache.load()
    floor = Decimal(0) if args.test else args.floor_usd
    print(f"reading live holders at floor ${floor} ...")
    holder_set = holders.read_holder_set(floor_usd=floor, exclusions=exclusions, accounts=accounts,
                                         test_address=config.TEST_ADDRESS if args.test else None)
    accounts.save()

    plan = planner.build_plan(holder_set=holder_set, floor_usd=floor, run_id=planner.next_run_id(events=events), canary=args.canary,
                              test=args.test, gas_per_transfer=planner.gas_per_transfer(events=events), max_addresses=args.max_addresses,
                              gas_budget=args.gas_budget, created_at=chainio.now().isoformat())
    planner.write_plan(plan=plan)
    print(render_plan(plan=plan))
    batches = len(plan.pending_batches())
    print(f"RESULT: PLANNED — run {plan.run_id}, {batches} batches under {config.STATE_DIR}; next: `cli.py run`")
    return 0


def render_plan(plan: planner.Plan) -> str:
    lines = [f"run {plan.run_id} · floor ${plan.floor_usd} · height {plan.height} · gas/transfer {plan.gas_per_transfer}"]
    for tier in plan.tiers:
        addresses = sum(len(b.addresses) for b in tier.batches)
        usd = sum((b.usd for b in tier.batches), Decimal(0))
        lines.append(f"  {tier.name:8} {len(tier.batches):4} batches {addresses:6} addresses ${usd:,.2f}")
    lines.append(f"  excluded {len(plan.excluded)} (${sum((e.usd for e in plan.excluded), Decimal(0)):,.2f}):")
    lines.extend(f"    {e.address}  {e.reason}  ${e.usd:,.2f}" for e in plan.excluded)
    lines.append(f"  skipped by chain rules {len(plan.skipped)} (${sum((e.usd for e in plan.skipped), Decimal(0)):,.2f}):")
    lines.extend(f"    {e.address}  {e.reason}  ${e.usd:,.2f}" for e in plan.skipped[:20])
    if len(plan.skipped) > 20:
        lines.append(f"    ... {len(plan.skipped) - 20} more in plan.json")
    lines.append(f"  below floor {plan.below_floor_count} addresses ${plan.below_floor_usd:,.2f}")
    lines.append("  ladder (holders not yet swept at each floor):")
    lines.extend(f"    ${r.floor:>6}  {r.holders:6} holders  ${r.usd:,.2f}" for r in plan.ladder)
    return "\n".join(lines)


# ---- run


def cmd_run(args: argparse.Namespace) -> int:
    plan = planner.load_plan()
    if plan is None:
        raise RunStopped("no plan; run `plan` first")
    events = ledger.read()
    exclusions = holders.load_exclusions()

    checks = preflight(plan=plan, events=events, exclusions=exclusions, now=chainio.now())
    for check in checks:
        print(f"  [{'ok' if check.ok else 'FAIL'}] {check.name}: {check.detail}")
    if not all(check.ok for check in checks):
        raise RunStopped("preflight failed")

    states = ledger.batch_states(events=events)
    todo = [(tier, batch) for tier, batch in plan.pending_batches()
            if batch.id not in states and (args.tier is None or tier.name == args.tier)]
    if args.batches is not None:
        todo = todo[: args.batches]
    print(f"{len(todo)} batch(es) to run")

    yes = args.yes
    done = 0
    for tier, batch in todo:
        event = run_batch(plan=plan, tier=tier, batch=batch, state_dir=config.STATE_DIR, ledger_path=config.LEDGER_PATH, yes=yes, dry_run=args.dry_run)
        if event is None and not args.dry_run:
            print("declined; stopping")
            break
        if event is not None:
            done += 1
            yes = yes or _answered_all
            if event.skipped and not args.continue_on_skip:
                raise RunStopped(f"batch {batch.id} had {len(event.skipped)} skip event(s): the planner and the chain disagree; "
                                 "inspect the ledger, fix the rule, re-plan (or pass --continue-on-skip)")
    print(f"RESULT: DONE — {done} batch(es) {'simulated' if args.dry_run else 'confirmed'}")
    return 0


_answered_all = False  # `a` at the prompt: yes to every later batch of this run


def ask(prompt: str) -> str:
    return input(prompt).strip().lower()


def run_batch(
    plan: planner.Plan,
    tier: planner.Tier,
    batch: planner.Batch,
    state_dir: pathlib.Path,
    ledger_path: pathlib.Path,
    yes: bool,
    dry_run: bool,
) -> ledger.Event | None:
    """One batch: describe, confirm, simulate, broadcast, poll, record. Returns the terminal event (or the submitted
    one when the poll gave up), None on a dry run or a declined prompt. Raises RunStopped when the run must not go on."""
    global _answered_all
    denoms = [d.denom for d in plan.denoms]
    file = state_dir / batch.file
    print(describe_batch(tier=tier, batch=batch))

    if not yes and not _answered_all and not dry_run:
        answer = ask(f"sweep batch {batch.id}? [y/N/a] ")
        if answer == ANSWER_ALL:
            _answered_all = True
        elif answer != ANSWER_YES:
            return None

    # Simulate first: a batch over the block limit must never be broadcast
    simulation = chainio.strided(args=command_line(denoms=denoms, file=file, gas=None, dry_run=True))
    estimate = parse_gas_estimate(text=simulation.stderr + simulation.stdout)
    if estimate is None:
        raise RunStopped(f"batch {batch.id}: simulation gave no gas estimate:\n{simulation.stderr}{simulation.stdout}")
    gas = int(Decimal(estimate) * config.GAS_ADJUSTMENT)
    print(f"  gas estimate {estimate:,} → wanted {gas:,} (planned {batch.estimated_gas:,})")
    if gas > config.BLOCK_GAS_LIMIT:
        raise RunStopped(f"batch {batch.id}: {gas:,} gas exceeds the block limit {config.BLOCK_GAS_LIMIT:,}; re-plan with a smaller --gas-budget")
    if dry_run:
        print("  dry run: " + " ".join([config.STRIDED_BINARY, *command_line(denoms=denoms, file=file, gas=gas, dry_run=False)]))
        return None

    broadcast = chainio.strided(args=command_line(denoms=denoms, file=file, gas=gas, dry_run=False))
    if broadcast.returncode != 0:
        raise RunStopped(f"batch {batch.id}: strided exited {broadcast.returncode}:\n{broadcast.stderr}")
    result = parse_broadcast(stdout=broadcast.stdout)
    if result.code != 0:
        event = ledger.failed(batch_id=batch.id, tx_hash=result.tx_hash, at=chainio.now().isoformat(), code=result.code,
                              codespace=result.codespace, raw_log=result.raw_log)
        ledger.append(event=event, path=ledger_path)
        raise RunStopped(f"batch {batch.id}: broadcast rejected (code {result.code} {result.codespace}): {result.raw_log}")

    ledger.append(event=ledger.submitted(run_id=plan.run_id, batch_id=batch.id, tx_hash=result.tx_hash, at=chainio.now().isoformat(),
                                         addresses=len(batch.addresses), transfers_estimate=batch.transfers, gas_wanted=gas), path=ledger_path)
    print(f"  submitted {result.tx_hash}; waiting for inclusion")
    outcome = poll_tx(tx_hash=result.tx_hash, wait_seconds=config.TX_WAIT_SECONDS)
    if outcome is None:
        raise RunStopped(f"batch {batch.id}: {result.tx_hash} not found after {config.TX_WAIT_SECONDS}s; run `resolve`")
    event = record_outcome(batch_id=batch.id, tx_hash=result.tx_hash, outcome=outcome, ledger_path=ledger_path)
    if event.kind == ledger.Kind.FAILED:
        raise RunStopped(f"batch {batch.id}: tx failed (code {outcome.code} {outcome.codespace}): {outcome.raw_log}")
    return event


def record_outcome(batch_id: str, tx_hash: str, outcome: TxOutcome, ledger_path: pathlib.Path) -> ledger.Event:
    at = chainio.now().isoformat()
    if outcome.code != 0:
        event = ledger.failed(batch_id=batch_id, tx_hash=tx_hash, at=at, code=outcome.code, codespace=outcome.codespace, raw_log=outcome.raw_log)
    else:
        event = ledger.confirmed(batch_id=batch_id, tx_hash=tx_hash, at=at, height=outcome.height, gas_used=outcome.gas_used,
                                 transfers=outcome.transfers, skipped=outcome.skipped)
        print(f"  confirmed at {outcome.height}: {len(outcome.transfers)} transfers, {len(outcome.skipped)} skipped, gas used {outcome.gas_used:,}")
        for skip in outcome.skipped:
            print(f"    SKIPPED {skip.address}: {skip.reason}")
    ledger.append(event=event, path=ledger_path)
    return event


def describe_batch(tier: planner.Tier, batch: planner.Batch) -> str:
    top = ", ".join(f"{p.address[:14]}… ${p.usd:,.2f}" for p in batch.addresses[:3])
    return (f"batch {batch.id} · tier {tier.name} · {len(batch.addresses)} addresses · {batch.transfers} transfers · "
            f"${batch.usd:,.2f} · largest: {top}")


def command_line(denoms: list[str], file: pathlib.Path, gas: int | None, dry_run: bool) -> list[str]:
    """The strided arguments for one batch. The test suite pins this shape."""
    line = ["tx", "stakeibc", "sweep-tokens-off-stride", ",".join(denoms), str(file),
            "--from", config.SWEEP_OPERATOR_KEY, "--keyring-backend", config.KEYRING_BACKEND,
            "--chain-id", config.CHAIN_ID, "--node", config.RPC]
    if dry_run:
        return line + ["--gas-prices", f"{config.GAS_PRICE_USTRD}{config.FEE_DENOM}", "--dry-run"]
    return line + ["--gas", str(gas), "--gas-prices", f"{config.GAS_PRICE_USTRD}{config.FEE_DENOM}",
                   "--broadcast-mode", "sync", "-y", "--output", "json"]


def parse_gas_estimate(text: str) -> int | None:
    match = GAS_ESTIMATE_PATTERN.search(text)
    return int(match.group(1)) if match else None


def parse_broadcast(stdout: str) -> BroadcastResult:
    body = json.loads(stdout)
    return BroadcastResult(code=int(body.get("code", 0)), tx_hash=body["txhash"], codespace=body.get("codespace", ""), raw_log=body.get("raw_log", ""))


def poll_tx(tx_hash: str, wait_seconds: int) -> TxOutcome | None:
    waited = 0
    while True:
        try:
            return parse_tx_response(body=chainio.rest_get(path=f"/cosmos/tx/v1beta1/txs/{tx_hash}"))
        except chainio.NotFound:
            pass
        if waited >= wait_seconds:
            return None
        chainio.sleep(config.TX_POLL_SECONDS)
        waited += config.TX_POLL_SECONDS


def parse_tx_response(body: dict) -> TxOutcome:
    response = body["tx_response"]
    transfers: list[ledger.Transfer] = []
    skipped: list[ledger.Skip] = []
    for event in response.get("events", []):
        attributes = {a["key"]: a["value"] for a in event.get("attributes", [])}
        if event["type"] == EVENT_SWEEP_TRANSFER:
            transfers.append(ledger.Transfer(address=attributes["address"], denom=attributes["denom"], amount=int(attributes["amount"]),
                                             channel=attributes["channel"], receiver=attributes["receiver"]))
        elif event["type"] == EVENT_SWEEP_SKIPPED:
            skipped.append(ledger.Skip(address=attributes["address"], reason=attributes["reason"]))
    return TxOutcome(code=int(response.get("code", 0)), height=int(response["height"]), gas_used=int(response.get("gas_used", 0)),
                     codespace=response.get("codespace", ""), raw_log=response.get("raw_log", ""), transfers=transfers, skipped=skipped)


# ---- preflight


def preflight(plan: planner.Plan, events: list[ledger.Event], exclusions: dict[str, holders.Exclusion], now: datetime.datetime) -> list[Check]:
    states = ledger.batch_states(events=events)
    pending = [batch for _, batch in plan.pending_batches() if batch.id not in states]
    return [
        _check_chain_id(),
        _check_binary_version(),
        _check_operator_key(),
        _check_plan_age(plan=plan, now=now),
        _check_batch_files(plan=plan),
        _check_unresolved(events=events),
        _check_exclusions(pending=pending, exclusions=exclusions),
        _check_channels(plan=plan),
        _check_fee_balance(pending=pending),
    ]


def _check_chain_id() -> Check:
    network = chainio.rest_get(path="/cosmos/base/tendermint/v1beta1/node_info")["default_node_info"]["network"]
    return Check(name="chain id", ok=network == config.CHAIN_ID, detail=f"node reports {network}")


def _check_binary_version() -> Check:
    version = chainio.strided(args=["version"]).stdout.strip().splitlines()[-1] if chainio.strided(args=["version"]).stdout.strip() else ""
    return Check(name="binary version", ok=version.startswith(config.BINARY_VERSION_PREFIX), detail=f"strided version {version or '?'}")


def _check_operator_key() -> Check:
    shown = chainio.strided(args=["keys", "show", config.SWEEP_OPERATOR_KEY, "--keyring-backend", config.KEYRING_BACKEND, "-a"]).stdout.strip()
    return Check(name="operator key", ok=shown == config.SWEEP_OPERATOR, detail=f"{config.SWEEP_OPERATOR_KEY} is {shown or 'missing'}")


def _check_plan_age(plan: planner.Plan, now: datetime.datetime) -> Check:
    age = (now - datetime.datetime.fromisoformat(plan.created_at)).total_seconds()
    return Check(name="plan age", ok=age <= config.MAX_PLAN_AGE_SECONDS, detail=f"run {plan.run_id} planned {age / 60:.0f} min ago at height {plan.height}")


def _check_batch_files(plan: planner.Plan) -> Check:
    bad = [batch.id for _, batch in plan.pending_batches()
           if not (config.STATE_DIR / batch.file).exists() or planner.batch_sha256(content=(config.STATE_DIR / batch.file).read_text()) != batch.sha256]
    return Check(name="batch files", ok=not bad, detail="every file matches the plan" if not bad else f"changed or missing: {', '.join(bad)}")


def _check_unresolved(events: list[ledger.Event]) -> Check:
    pending = ledger.unresolved(events=events)
    return Check(name="unresolved submissions", ok=not pending, detail="none" if not pending else f"run `resolve`: {', '.join(e.batch_id for e in pending)}")


def _check_exclusions(pending: list[planner.Batch], exclusions: dict[str, holders.Exclusion]) -> Check:
    forbidden = set(exclusions) | set(config.PROTOCOL_ADDRESSES) | {config.SWEEP_OPERATOR}
    hits = [f"{p.address} ({batch.id})" for batch in pending for p in batch.addresses if p.address in forbidden]
    return Check(name="exclusions and protocol addresses", ok=not hits, detail="no pending batch names one" if not hits else f"in a batch: {', '.join(hits)}")


def _check_channels(plan: planner.Plan) -> Check:
    closed = []
    for channel in sorted({d.channel for d in plan.denoms}):
        state = chainio.rest_get(path=f"/ibc/core/channel/v1/channels/{channel}/ports/{config.TRANSFER_PORT}")["channel"]["state"]
        if state != STATE_OPEN:
            closed.append(f"{channel} {state}")
    return Check(name="channels open", ok=not closed, detail="every destination channel is OPEN" if not closed else ", ".join(closed))


def _check_fee_balance(pending: list[planner.Batch]) -> Check:
    body = chainio.rest_get(path=f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom", params={"denom": config.FEE_DENOM})
    balance = int(body["balance"]["amount"])
    needed = int(sum(Decimal(b.estimated_gas) * config.GAS_ADJUSTMENT * config.GAS_PRICE_USTRD for b in pending))
    return Check(name="operator fee balance", ok=balance >= needed, detail=f"{balance / 1e6:.2f} STRD held, {needed / 1e6:.2f} STRD needed for {len(pending)} batches")


# ---- status / resolve


def cmd_status(args: argparse.Namespace) -> int:
    plan = planner.load_plan()
    events = ledger.read()
    if plan is None:
        print("no plan")
        return 0
    states = ledger.batch_states(events=events)
    print(f"run {plan.run_id} · floor ${plan.floor_usd} · planned {plan.created_at} at height {plan.height}")
    for tier in plan.tiers:
        counts: dict[str, int] = {}
        for batch in tier.batches:
            state = states.get(batch.id, ledger.BatchState.PENDING)
            counts[state] = counts.get(state, 0) + 1
        print(f"  {tier.name:8} " + ", ".join(f"{n} {s}" for s, n in counts.items()))
    confirmed = [e for e in events if e.kind == ledger.Kind.CONFIRMED]
    transfers = sum(len(e.transfers) for e in confirmed)
    skipped = [(e.batch_id, s) for e in confirmed for s in e.skipped]
    print(f"  confirmed {len(confirmed)} batches, {transfers} transfers, {len(skipped)} skips")
    for batch_id, skip in skipped:
        print(f"    SKIP {batch_id} {skip.address}: {skip.reason}")
    for event in events:
        if event.kind == ledger.Kind.FAILED:
            print(f"    FAILED {event.batch_id} {event.tx_hash} code {event.code}: {event.raw_log}")
    for event in ledger.unresolved(events=events):
        print(f"    UNRESOLVED {event.batch_id} {event.tx_hash} submitted {event.at}")
    return 0


def cmd_resolve(args: argparse.Namespace) -> int:
    events = ledger.read()
    pending = ledger.unresolved(events=events)
    if not pending:
        print("nothing to resolve")
        return 0
    for event in pending:
        assert event.tx_hash is not None
        submitted_at = datetime.datetime.fromisoformat(event.at)
        remaining = max(0, args.wait_seconds - int((chainio.now() - submitted_at).total_seconds()))
        outcome = poll_tx(tx_hash=event.tx_hash, wait_seconds=remaining)
        if outcome is None:
            ledger.append(event=ledger.lost(batch_id=event.batch_id, tx_hash=event.tx_hash, at=chainio.now().isoformat()))
            print(f"  {event.batch_id} {event.tx_hash}: not found after {args.wait_seconds}s, recorded as lost; the next `plan` reads the balances")
            continue
        recorded = record_outcome(batch_id=event.batch_id, tx_hash=event.tx_hash, outcome=outcome, ledger_path=config.LEDGER_PATH)
        print(f"  {event.batch_id} {event.tx_hash}: {recorded.kind}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

Fix two things while writing it: `_check_binary_version` calls `strided` twice; call it once into a local. And `cmd_run` reads the module-level `_answered_all` after each batch (`yes = yes or _answered_all`), which is fine but make sure `run_batch` declares `global _answered_all` before assigning. In `RunLoopTests`, `ask` returns `a`, so the second batch runs without a prompt.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd scripts/wind-down && python3 -m unittest sweep.test_cli -v`
Expected: 15 tests, OK.

- [ ] **Step 5: Write the README**

`scripts/wind-down/sweep/README.md`:

```markdown
# Sweep runner

The holder sweep of the v35 wind-down (`MsgSweepTokensOffStride`, design in
`docs/superpowers/specs/2026-10-09-wind-down-sweep-runner-design.md`). Stdlib only.

    python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 --canary 3   # live state → state/plan.json + batch files
    python3 scripts/wind-down/sweep/cli.py run                               # preflight, then sign and submit each batch
    python3 scripts/wind-down/sweep/cli.py status                            # no network
    python3 scripts/wind-down/sweep/cli.py resolve                           # finish a submission the poll gave up on

The operator key is `stride-sweeper` in the **test** keyring. `config.py` holds the denoms, the rough prices, gas
settings and limits; `exclusions.json` the accounts the team moves by hand (a section, a reason, a label per address).
Both are enforced at `plan` and again at `run`.

## The test (after the 10-12 upgrade, before the sweep)

    python3 scripts/wind-down/sweep/cli.py plan --test     # one batch: stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg
    python3 scripts/wind-down/sweep/cli.py run

Then check the same bytes under `osmo1…` hold the stToken and STRD and under `cosmos1…` the ATOM, and that the Sweep
tab shows the batch confirmed. The test batch also calibrates gas per transfer for the real run.

## Sweep day

1. `plan --floor-usd <announced floor> --canary 3`, read the tables (excluded, skipped, ladder), `run`.
   The canary tier (the three smallest holders above the floor) goes first, then everyone by value descending.
2. Lower the floor: `plan --floor-usd 25`, `run`; repeat. Swept holders have no balance, so a re-plan never lists them;
   a transfer that timed out (24 h) refunds the holder, who reappears in the next plan.
3. The `keyless` tier (never-signed accounts) is last in every plan; run it at the end with `run --tier keyless`.

`run` stops on any `sweep_skipped` event (the planner and the chain disagree: inspect, fix, re-plan), on a failed tx,
on a batch over the block gas limit, and when a tx is not found within three minutes (`resolve`). `--dry-run` prints
every command and gas estimate without broadcasting; `--yes` skips the per-batch prompt (`a` at the prompt does the
same for the rest of the run); `--batches N` stops after N.

State files under `state/` are committed like `dashboard/ops/status.json`: `plan.json` and the batch files are
overwritten by every `plan`; `ledger.jsonl` is append-only.
```

- [ ] **Step 6: Run both suites, then commit**

Run: `cd scripts/wind-down && python3 -m unittest discover -s sweep -t .`
Expected: all OK.

```bash
git add scripts/wind-down/sweep/cli.py scripts/wind-down/sweep/test_cli.py scripts/wind-down/sweep/README.md
git commit -m "wind-down sweep: the runner CLI (plan, run with preflight and ledger, status, resolve)"
```

---
### Task 6: The dashboard Sweep tab

**Files:**
- Create: `scripts/wind-down/dashboard/sweep_tab.py`, `scripts/wind-down/dashboard/static/sweep.js`
- Modify: `scripts/wind-down/dashboard/server.py` (imports, `COLLECTORS`, a `/api/sweep` route), `scripts/wind-down/dashboard/config.py` (`REFRESH_INTERVAL_SECONDS["sweep"] = 1800`), `scripts/wind-down/dashboard/static/index.html` (tab button, view section, script tag), `scripts/wind-down/dashboard/static/app.js` (`TAB_NAMES` adds `'sweep'`; `SELF_POLLING_LABELS.sweep`), `scripts/wind-down/dashboard/static/style.css` (a `/* sweep */` block), `scripts/wind-down/dashboard/README.md` (a "Sweep tab" section)
- Test: `scripts/wind-down/dashboard/test_sweep_tab.py`

**Interfaces:**
- Consumes: `sweep.holders.resolve_denoms() -> list[SweepDenom]`, `sweep.holders.read_balances(denoms) -> dict[str, dict[str, int]]`, `sweep.holders.load_exclusions() -> dict[str, Exclusion]`, `sweep.chainio.latest_height()`, `sweep.chainio.rest_get`, `sweep.ledger.read() -> list[Event]`, `sweep.ledger.batch_states`, `sweep.ledger.Kind`, `sweep.planner.load_plan() -> Plan | None`, `sweep.planner.Plan` (fields `run_id, created_at, height, floor_usd, test, prices, denoms, tiers[].name/.batches[].id/.addresses[].address/.balances/.usd, excluded[], skipped[], below_floor_count, below_floor_usd`), `sweep.config.SWEEP_OPERATOR`, `sweep.config.FEE_DENOM`, `sweep.config.GAS_ADJUSTMENT`, `sweep.config.GAS_PRICE_USTRD`, `sweep.config.LADDER_FLOORS`.
- Depends on: Tasks 1-4
- Produces: `sweep_tab.collect() -> dict` (`{"height": str, "denoms": [...], "balances": {address: {denom: str}}, "operator_strd": str}`), `sweep_tab.compose(plan: Plan | None, events: list[Event], live: dict | None, exclusions: dict[str, Exclusion], live_fetched_at: str | None) -> dict` (the `data` of `GET /api/sweep`), `sweep_tab.body(view: dict) -> dict` (the route body: `{fetched_at, refreshing, data}`; reads plan, ledger and exclusions from disk).
- Review: no

The sweep package is imported with `scripts/wind-down` on `sys.path`. Add to the top of `sweep_tab.py` (and nothing else needs it, because `server.py` imports `sweep_tab`):

```python
import pathlib
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package lives beside the dashboard

from sweep import chainio as sweep_chainio  # noqa: E402
from sweep import config as sweep_config  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402
```

- [ ] **Step 1: Write the failing test**

`scripts/wind-down/dashboard/test_sweep_tab.py`:

```python
"""compose(): address states from the plan, the ledger and live balances; totals; by-denom; ladder; batches."""

import pathlib
import sys
import unittest
from decimal import Decimal

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package, before any import of it

import sweep_tab  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

STATOM = holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5")
STRD = holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5")
A_SWEPT, A_REFUNDED, A_PENDING, A_EXCLUDED, A_DUST, A_NEW, A_ESCROW = (f"stride1{name}" for name in ("swept", "refund", "pend", "excl", "dust", "new", "escrow"))


def make_plan() -> planner.Plan:
    def h(address: str, statom: int, keyless: bool = False) -> holders.Holder:
        return holders.Holder(address=address, balances={"stuatom": statom, "ustrd": 1_000_000}, usd=Decimal(statom) / 1_000_000 * 6 + Decimal("0.05"), keyless=keyless)
    hs = holders.HolderSet(
        height=100, denoms=[STATOM, STRD],
        holders=[h(A_SWEPT, 10_000_000), h(A_REFUNDED, 5_000_000), h(A_PENDING, 2_000_000, keyless=True)],
        excluded=[holders.Excluded(address=A_EXCLUDED, reason="excluded: team: F5", usd=Decimal(99))],
        skipped=[holders.Excluded(address=A_ESCROW, reason="transfer escrow address", usd=Decimal(1000))],
        below_floor=[h(A_DUST, 100_000)],
    )
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=2, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=1, gas_budget=40_000_000, created_at="2026-11-10T09:00:00+00:00")


def t(address: str, denom: str, amount: int) -> ledger.Transfer:
    return ledger.Transfer(address=address, denom=denom, amount=amount, channel="channel-5", receiver="osmo1x")


EVENTS = [
    ledger.submitted(run_id=2, batch_id="002-001", tx_hash="A", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-001", tx_hash="A", at="t", height=5, gas_used=1, transfers=[t(A_SWEPT, "stuatom", 10_000_000), t(A_SWEPT, "ustrd", 1_000_000)], skipped=[]),
    ledger.submitted(run_id=2, batch_id="002-002", tx_hash="B", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-002", tx_hash="B", at="t", height=6, gas_used=1, transfers=[t(A_REFUNDED, "stuatom", 5_000_000), t(A_REFUNDED, "ustrd", 1_000_000)],
                     skipped=[ledger.Skip(address="stride1other", reason="interchain account")]),
]
LIVE = {
    "height": "200",
    "denoms": [{"denom": "stuatom", "symbol": "stATOM", "decimals": "6", "price_usd": "6", "destination": "osmosis-1", "channel": "channel-5"},
               {"denom": "ustrd", "symbol": "STRD", "decimals": "6", "price_usd": "0.05", "destination": "osmosis-1", "channel": "channel-5"}],
    "balances": {
        A_REFUNDED: {"stuatom": "5000000"},  # the stATOM transfer timed out and came back; the STRD landed
        A_PENDING: {"stuatom": "2000000", "ustrd": "1000000"},
        A_EXCLUDED: {"stuatom": "16500000"},
        A_DUST: {"stuatom": "100000"},
        A_NEW: {"stuatom": "3000000"},  # above the floor but not in the plan: appeared since
        A_ESCROW: {"stuatom": "999000000"},
    },
    "operator_strd": "212400000",
}


class ComposeTests(unittest.TestCase):
    def setUp(self) -> None:
        self.data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE, exclusions={}, live_fetched_at="2026-11-10T14:00:00+00:00")

    def test_run_and_batches(self) -> None:
        run = self.data["run"]
        self.assertEqual((run["run_id"], run["floor_usd"], run["height"]), ("2", "10", "100"))
        self.assertEqual(run["batches"], {"total": "3", "confirmed": "2", "submitted": "0", "failed": "0", "lost": "0", "pending": "1"})
        self.assertEqual(run["operator_strd"], "212400000")
        batches = {b["batch_id"]: b for b in self.data["batches"]}
        self.assertEqual(batches["002-001"]["state"], "confirmed")
        self.assertEqual(batches["002-002"]["skipped"], [{"address": "stride1other", "reason": "interchain account"}])
        self.assertEqual(batches["002-002"]["refunded"], "1")
        self.assertEqual((batches["002-003"]["state"], batches["002-003"]["tier"], batches["002-003"]["tx_hash"]), ("pending", "keyless", None))

    def test_address_states_and_totals(self) -> None:
        totals = self.data["totals"]
        self.assertEqual(totals["swept"], {"addresses": "1", "usd": "60.05"})
        self.assertEqual(totals["refunded"], {"addresses": "1", "usd": "30.00"})  # the stATOM that came back, at plan prices
        self.assertEqual(totals["remaining"], {"addresses": "1", "usd": "12.05"})
        self.assertEqual(totals["excluded"], {"addresses": "1", "usd": "99.00"})
        self.assertEqual(totals["below_floor"], {"addresses": "1", "usd": "0.60"})
        self.assertEqual(totals["unplanned"], {"addresses": "1", "usd": "18.00"})
        self.assertEqual(self.data["refunded"], [{"address": A_REFUNDED, "denom": "stuatom", "amount": "5000000", "channel": "channel-5", "batch_id": "002-002"}])

    def test_by_denom(self) -> None:
        by_denom = {d["denom"]: d for d in self.data["by_denom"]}
        self.assertEqual(by_denom["stuatom"]["swept"], {"addresses": "2", "amount": "15000000", "usd": "90.00"})
        self.assertEqual(by_denom["stuatom"]["remaining"], {"addresses": "1", "amount": "2000000", "usd": "12.00"})
        self.assertEqual(by_denom["stuatom"]["refunded_addresses"], "1")
        self.assertEqual(by_denom["stuatom"]["excluded_usd"], "99.00")
        self.assertEqual(by_denom["stuatom"]["below_floor_usd"], "0.60")
        self.assertEqual(by_denom["ustrd"]["destination"], "osmosis-1")

    def test_ladder_counts_live_holders_still_to_sweep(self) -> None:
        # not yet swept and not excluded/skipped: refunded (30), pending (12.05), new (18), dust (0.60)
        self.assertEqual([(r["floor"], r["holders"], r["usd"]) for r in self.data["ladder"]],
                         [("10", "3", "60.05"), ("5", "3", "60.05"), ("1", "3", "60.05"), ("0", "4", "60.65")])

    def test_keyless_and_exclusions_sections(self) -> None:
        self.assertEqual(self.data["keyless"], {"addresses": "1", "usd": "12.05", "swept": "0"})
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE,
                                 exclusions={A_EXCLUDED: holders.Exclusion(address=A_EXCLUDED, section="team", label="F5", reason="by hand")},
                                 live_fetched_at="x")
        self.assertEqual(data["exclusions"], [{"section": "team", "address": A_EXCLUDED, "label": "F5", "reason": "by hand", "live_usd": "99.00"}])


class EdgeTests(unittest.TestCase):
    def test_no_plan(self) -> None:
        data = sweep_tab.compose(plan=None, events=[], live=None, exclusions={}, live_fetched_at=None)
        self.assertIsNone(data["run"])
        self.assertEqual(data["batches"], [])

    def test_no_live_snapshot_keeps_plan_and_ledger_parts(self) -> None:
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=None, exclusions={}, live_fetched_at=None)
        self.assertEqual(data["run"]["batches"]["confirmed"], "2")
        self.assertIsNone(data["totals"])
        self.assertIsNone(data["ladder"])
        self.assertEqual(len(data["batches"]), 3)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `python3 -m unittest discover -s scripts/wind-down/dashboard -p test_sweep_tab.py -v`
Expected: `ModuleNotFoundError: No module named 'sweep_tab'`.

- [ ] **Step 3: Write sweep_tab.py**

```python
"""The Sweep tab: `collect()` is the expensive live read (every holder of every sweep denom, in bulk) and is the tab's
collector, refreshed on demand; `body()` composes it with the plan, the ledger and the exclusions file read from disk
on every request, through the pure `compose()`. Every integer in the payload is a string."""

import pathlib
import sys
from decimal import ROUND_HALF_UP, Decimal
from typing import Any

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package lives beside the dashboard

from sweep import chainio as sweep_chainio  # noqa: E402
from sweep import config as sweep_config  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

CENTS = Decimal("0.01")
STATE_SWEPT = "swept"
STATE_REFUNDED = "refunded"
STATE_REMAINING = "remaining"


# ---- collector


def collect() -> dict[str, Any]:
    denoms = holders.resolve_denoms()
    balances = holders.read_balances(denoms=denoms)
    operator = sweep_chainio.rest_get(path=f"/cosmos/bank/v1beta1/balances/{sweep_config.SWEEP_OPERATOR}/by_denom",
                                      params={"denom": sweep_config.FEE_DENOM})
    return {
        "height": str(sweep_chainio.latest_height()),
        "denoms": [{"denom": d.denom, "symbol": d.symbol, "decimals": str(d.decimals), "price_usd": str(d.price_usd),
                    "destination": d.destination, "channel": d.channel} for d in denoms],
        "balances": {address: {denom: str(amount) for denom, amount in coins.items()} for address, coins in balances.items()},
        "operator_strd": operator["balance"]["amount"],
    }


# ---- route body


def body(view: dict[str, Any]) -> dict[str, Any]:
    """`GET /api/sweep`: plan, ledger and exclusions fresh from disk; the live part from the collector's snapshot."""
    live = None if view.get("loading") else view["data"]
    fetched_at = view.get("fetched_at")
    try:
        plan = planner.load_plan()
        events = ledger.read()
        exclusions = holders.load_exclusions()
    except (planner.PlanError, ledger.LedgerError, holders.ExclusionsError) as error:
        return {"fetched_at": fetched_at, "refreshing": view.get("refreshing", False), "error": str(error), "data": None}
    return {"fetched_at": fetched_at, "refreshing": view.get("refreshing", False), "last_error": view.get("last_error"),
            "data": compose(plan=plan, events=events, live=live, exclusions=exclusions, live_fetched_at=fetched_at)}


# ---- composition


def compose(
    plan: planner.Plan | None,
    events: list[ledger.Event],
    live: dict[str, Any] | None,
    exclusions: dict[str, holders.Exclusion],
    live_fetched_at: str | None,
) -> dict[str, Any]:
    if plan is None:
        return {"run": None, "totals": None, "by_denom": None, "batches": [], "ladder": None, "refunded": [], "exclusions": [], "keyless": None,
                "live_fetched_at": live_fetched_at}

    states = ledger.batch_states(events=events)
    transferred = ledger.confirmed_transfers(events=events)
    balances = _live_balances(live=live)
    by_denom = {d.denom: d for d in plan.denoms}
    address_state = _address_states(plan=plan, transferred=transferred, balances=balances) if live is not None else None
    refunds = _refunds(plan=plan, events=events, balances=balances) if live is not None else []

    return {
        "run": _run(plan=plan, states=states, live=live),
        "totals": _totals(plan=plan, address_state=address_state, transferred=transferred, balances=balances, by_denom=by_denom) if live is not None else None,
        "by_denom": _by_denom(plan=plan, address_state=address_state, transferred=transferred, balances=balances) if live is not None else None,
        "batches": _batches(plan=plan, events=events, states=states, refunds=refunds),
        "ladder": _ladder(plan=plan, address_state=address_state, balances=balances, by_denom=by_denom) if live is not None else None,
        "refunded": [{"address": r.address, "denom": r.denom, "amount": str(r.amount), "channel": r.channel, "batch_id": batch_id} for r, batch_id in refunds],
        "exclusions": [{"section": e.section, "address": e.address, "label": e.label, "reason": e.reason,
                        "live_usd": _money(holders.usd_value(balances=balances.get(e.address, {}), denoms=by_denom)) if live is not None else None}
                       for e in exclusions.values()],
        "keyless": _keyless(plan=plan, address_state=address_state),
        "live_fetched_at": live_fetched_at,
    }


def _live_balances(live: dict[str, Any] | None) -> dict[str, dict[str, int]]:
    if live is None:
        return {}
    return {address: {denom: int(amount) for denom, amount in coins.items()} for address, coins in live["balances"].items()}


def _planned(plan: planner.Plan) -> dict[str, planner.PlannedAddress]:
    return {p.address: p for _, batch in plan.pending_batches() for p in batch.addresses}


def _address_states(plan: planner.Plan, transferred: dict[str, list[ledger.Transfer]], balances: dict[str, dict[str, int]]) -> dict[str, str]:
    """swept: transferred and holding none of it now; refunded: transferred and holding a transferred denom again;
    remaining: no confirmed transfer yet."""
    states: dict[str, str] = {}
    for address in _planned(plan=plan):
        moved = {t.denom for t in transferred.get(address, [])}
        if not moved:
            states[address] = STATE_REMAINING
            continue
        live = balances.get(address, {})
        states[address] = STATE_REFUNDED if any(live.get(denom, 0) > 0 for denom in moved) else STATE_SWEPT
    return states


def _refunds(plan: planner.Plan, events: list[ledger.Event], balances: dict[str, dict[str, int]]) -> list[tuple[ledger.Transfer, str]]:
    planned = _planned(plan=plan)
    refunds: list[tuple[ledger.Transfer, str]] = []
    for event in events:
        if event.kind != ledger.Kind.CONFIRMED:
            continue
        for transfer in event.transfers:
            if transfer.address in planned and balances.get(transfer.address, {}).get(transfer.denom, 0) > 0:
                refunds.append((transfer, event.batch_id))
    return refunds


def _run(plan: planner.Plan, states: dict[str, ledger.BatchState], live: dict[str, Any] | None) -> dict[str, Any]:
    batches = [batch for _, batch in plan.pending_batches()]
    counts = {state.value: 0 for state in ledger.BatchState}
    for batch in batches:
        counts[states.get(batch.id, ledger.BatchState.PENDING).value] += 1
    pending_gas = sum(b.estimated_gas for b in batches if b.id not in states)
    fee = Decimal(pending_gas) * sweep_config.GAS_ADJUSTMENT * sweep_config.GAS_PRICE_USTRD
    return {
        "run_id": str(plan.run_id), "floor_usd": str(plan.floor_usd), "created_at": plan.created_at, "height": str(plan.height), "test": plan.test,
        "batches": {"total": str(len(batches)), **{k: str(v) for k, v in counts.items() if k != "pending"}, "pending": str(counts["pending"])},
        "operator_strd": live["operator_strd"] if live is not None else None,
        "fee_estimate_ustrd": str(int(fee)),
    }


def _totals(plan, address_state, transferred, balances, by_denom) -> dict[str, dict[str, str]]:
    planned = _planned(plan=plan)
    swept = [a for a, s in address_state.items() if s == STATE_SWEPT]
    refunded = [a for a, s in address_state.items() if s == STATE_REFUNDED]
    remaining = [a for a, s in address_state.items() if s == STATE_REMAINING]
    swept_usd = sum((_transfer_usd(t, by_denom) for a in swept for t in transferred[a]), Decimal(0))
    refunded_usd = sum((holders.usd_value(balances={d: v for d, v in balances.get(a, {}).items() if d in {t.denom for t in transferred[a]}}, denoms=by_denom)
                        for a in refunded), Decimal(0))
    remaining_usd = sum((holders.usd_value(balances=balances.get(a, {}), denoms=by_denom) for a in remaining), Decimal(0))
    excluded_usd = sum((holders.usd_value(balances=balances.get(e.address, {}), denoms=by_denom) for e in plan.excluded), Decimal(0))
    known = set(planned) | {e.address for e in plan.excluded} | {e.address for e in plan.skipped}
    others = {a: holders.usd_value(balances=c, denoms=by_denom) for a, c in balances.items() if a not in known}
    below = {a: u for a, u in others.items() if u < plan.floor_usd}
    unplanned = {a: u for a, u in others.items() if u >= plan.floor_usd}
    return {
        "swept": _bucket(len(swept), swept_usd), "refunded": _bucket(len(refunded), refunded_usd), "remaining": _bucket(len(remaining), remaining_usd),
        "excluded": _bucket(len(plan.excluded), excluded_usd), "below_floor": _bucket(len(below), sum(below.values(), Decimal(0))),
        "unplanned": _bucket(len(unplanned), sum(unplanned.values(), Decimal(0))),
    }


def _by_denom(plan, address_state, transferred, balances) -> list[dict[str, Any]]:
    by_denom = {d.denom: d for d in plan.denoms}
    planned = _planned(plan=plan)
    known = set(planned) | {e.address for e in plan.excluded} | {e.address for e in plan.skipped}
    rows = []
    for d in plan.denoms:
        swept_transfers = [t for a, ts in transferred.items() if a in planned for t in ts if t.denom == d.denom]
        swept_amount = sum(t.amount for t in swept_transfers)
        remaining = {a: balances.get(a, {}).get(d.denom, 0) for a, s in address_state.items() if s == STATE_REMAINING}
        remaining = {a: v for a, v in remaining.items() if v > 0}
        refunded = [a for a, s in address_state.items() if s == STATE_REFUNDED and balances.get(a, {}).get(d.denom, 0) > 0]
        excluded_usd = sum((_amount_usd(balances.get(e.address, {}).get(d.denom, 0), d) for e in plan.excluded), Decimal(0))
        below_usd = sum((_amount_usd(c.get(d.denom, 0), d) for a, c in balances.items()
                         if a not in known and holders.usd_value(balances=c, denoms=by_denom) < plan.floor_usd), Decimal(0))
        rows.append({
            "denom": d.denom, "symbol": d.symbol, "decimals": str(d.decimals), "destination": d.destination, "channel": d.channel,
            "swept": {"addresses": str(len({t.address for t in swept_transfers})), "amount": str(swept_amount), "usd": _money(_amount_usd(swept_amount, d))},
            "remaining": {"addresses": str(len(remaining)), "amount": str(sum(remaining.values())), "usd": _money(_amount_usd(sum(remaining.values()), d))},
            "refunded_addresses": str(len(refunded)), "excluded_usd": _money(excluded_usd), "below_floor_usd": _money(below_usd),
        })
    return rows


def _batches(plan, events, states, refunds) -> list[dict[str, Any]]:
    by_batch: dict[str, ledger.Event] = {}
    for event in events:
        by_batch[event.batch_id] = event  # the last line for the batch is its current state
    submitted = {e.batch_id: e for e in events if e.kind == ledger.Kind.SUBMITTED}
    refunded_per_batch: dict[str, int] = {}
    for _, batch_id in refunds:
        refunded_per_batch[batch_id] = refunded_per_batch.get(batch_id, 0) + 1
    rows = []
    for tier, batch in plan.pending_batches():
        last = by_batch.get(batch.id)
        rows.append({
            "run_id": str(plan.run_id), "batch_id": batch.id, "tier": tier.name.value, "addresses": str(len(batch.addresses)), "usd": _money(batch.usd),
            "state": states.get(batch.id, ledger.BatchState.PENDING).value, "tx_hash": last.tx_hash if last else None, "at": last.at if last else None,
            "skipped": [{"address": s.address, "reason": s.reason} for s in (last.skipped if last else [])],
            "refunded": str(refunded_per_batch.get(batch.id, 0)),
        })
    # Batches of earlier runs, from the ledger alone (the plan file only holds the current run)
    current = {batch.id for _, batch in plan.pending_batches()}
    for batch_id, event in by_batch.items():
        if batch_id in current:
            continue
        first = submitted.get(batch_id)
        rows.append({
            "run_id": batch_id.split("-")[0].lstrip("0") or "0", "batch_id": batch_id, "tier": None,
            "addresses": str(first.addresses) if first and first.addresses is not None else None, "usd": None,
            "state": states[batch_id].value, "tx_hash": event.tx_hash, "at": event.at,
            "skipped": [{"address": s.address, "reason": s.reason} for s in event.skipped], "refunded": str(refunded_per_batch.get(batch_id, 0)),
        })
    rows.sort(key=lambda r: r["batch_id"])
    return rows


def _ladder(plan, address_state, balances, by_denom) -> list[dict[str, str]]:
    """Live holders still to sweep: everyone with a balance except the swept, the excluded and the chain-skipped."""
    out = {a for a, s in address_state.items() if s == STATE_SWEPT} | {e.address for e in plan.excluded} | {e.address for e in plan.skipped}
    values = [holders.usd_value(balances=c, denoms=by_denom) for a, c in balances.items() if a not in out]
    floors = [plan.floor_usd] + [f for f in sweep_config.LADDER_FLOORS if f < plan.floor_usd]
    return [{"floor": str(f), "holders": str(sum(1 for v in values if v >= f)), "usd": _money(sum((v for v in values if v >= f), Decimal(0)))} for f in floors]


def _keyless(plan, address_state) -> dict[str, str] | None:
    tier = next((t for t in plan.tiers if t.name == planner.TierName.KEYLESS), None)
    if tier is None:
        return None
    addresses = [p for b in tier.batches for p in b.addresses]
    swept = sum(1 for p in addresses if address_state is not None and address_state.get(p.address) == STATE_SWEPT)
    return {"addresses": str(len(addresses)), "usd": _money(sum((p.usd for p in addresses), Decimal(0))), "swept": str(swept)}


def _transfer_usd(transfer: ledger.Transfer, by_denom: dict[str, holders.SweepDenom]) -> Decimal:
    return _amount_usd(transfer.amount, by_denom[transfer.denom]) if transfer.denom in by_denom else Decimal(0)


def _amount_usd(amount: int, denom: holders.SweepDenom) -> Decimal:
    return Decimal(amount) / Decimal(10) ** denom.decimals * denom.price_usd


def _bucket(count: int, usd: Decimal) -> dict[str, str]:
    return {"addresses": str(count), "usd": _money(usd)}


def _money(value: Decimal) -> str:
    return str(value.quantize(CENTS, rounding=ROUND_HALF_UP))
```

Type the private helpers' parameters fully (`plan: planner.Plan`, `address_state: dict[str, str]`, `transferred: dict[str, list[ledger.Transfer]]`, `balances: dict[str, dict[str, int]]`, `by_denom: dict[str, holders.SweepDenom]`, `events: list[ledger.Event]`, `states: dict[str, ledger.BatchState]`, `refunds: list[tuple[ledger.Transfer, str]]`); the sketch above elides them for width.

- [ ] **Step 4: Run the test to verify it passes**

Run: `python3 -m unittest discover -s scripts/wind-down/dashboard -p test_sweep_tab.py -v`
Expected: 7 tests, OK. If `test_address_states_and_totals` disagrees on `refunded` USD: it is the live balance of the transferred denoms only (5 stATOM × $6 = $30.00), not the whole live balance.

- [ ] **Step 5: Wire the server, config, page shell and index**

`scripts/wind-down/dashboard/config.py`: `REFRESH_INTERVAL_SECONDS = {"channels": 60, "funds": 120, "validators": 300, "pools": 300, "sweep": 1800}`.

`scripts/wind-down/dashboard/server.py`:
- `import sweep_tab` beside the other imports; `COLLECTORS["sweep"] = sweep_tab.collect`.
- In `do_GET`, before the generic `/api/` branch: `elif path == "/api/sweep": self._get_sweep()`.
- Add:

```python
    def _get_sweep(self) -> None:
        """The Sweep tab composes the live holder snapshot with the plan and ledger read from disk on every request."""
        view = CACHES["sweep"].view() if "sweep" in CACHES else {"loading": True}
        body = sweep_tab.body(view=view)
        self._send_json(status=500 if body.get("error") else 200, body=body)
```

- Extend the module docstring's list of exceptions with the Sweep tab in one sentence.

`scripts/wind-down/dashboard/static/app.js`: `TAB_NAMES` becomes `['ops', 'channels', 'validators', 'funds', 'pools', 'multisig', 'sweep']`; add `sweep: 'live holders refresh on demand (button in the tab) · plan and ledger are read from disk'` to `SELF_POLLING_LABELS`.

`scripts/wind-down/dashboard/static/index.html`: add `<button class="tab" data-tab="sweep">Sweep</button>` after the Multisig button, `<section class="view" id="view-sweep"></section>` after the multisig section, and `<script src="/sweep.js"></script>` after `multisig.js`.

- [ ] **Step 6: Write sweep.js**

`scripts/wind-down/dashboard/static/sweep.js`:

```javascript
// Sweep tab: the holder sweep's progress. Like Ops and Multisig it polls its own route, /api/sweep, which composes the
// live holder snapshot (refreshed on demand: the button here, or automatically when the snapshot is missing or older
// than ten minutes while this tab is open) with the plan and ledger read from disk. Every integer is a string.
(() => {

const SWEEP_POLL_MS = 10 * 1000;
const AUTO_REFRESH_AFTER_MS = 10 * 60 * 1000;
const STATE_CLASS = { confirmed: 'ok', submitted: 'warn', failed: 'bad', lost: 'bad', pending: 'idle' };

let lastBody = '';
let refreshRequested = false;

registerSelfPollingTab('sweep', startSweep);

function startSweep() {
  const root = document.getElementById('view-sweep');
  root.innerHTML = '<div id="sweepMessage"></div><div id="sweepHead" class="panel"></div><div id="sweepBody"></div>';
  root.addEventListener('click', (event) => {
    if (event.target.id === 'sweepRefresh') requestRefresh();
  });
  pollSweep();
  setInterval(pollSweep, SWEEP_POLL_MS);
}

async function pollSweep() {
  const response = await fetch('/api/sweep').catch(() => null);
  const raw = response ? await response.text() : '';
  if (!response || response.status !== 200) {
    showMessage(response ? `could not read the sweep state: ${errorText(raw)}` : 'server unreachable');
    return;
  }
  showMessage('');
  const body = JSON.parse(raw);
  maybeAutoRefresh(body);
  if (raw === lastBody) return;
  lastBody = raw;
  draw(body);
}

function maybeAutoRefresh(body) {
  if (location.hash.slice(1).split('/')[0] !== 'sweep' || body.refreshing || refreshRequested) return;
  const age = body.fetched_at ? Date.now() - Date.parse(body.fetched_at) : Infinity;
  if (age > AUTO_REFRESH_AFTER_MS) requestRefresh();
}

async function requestRefresh() {
  refreshRequested = true;
  await fetchJson('/api/refresh/sweep', { method: 'POST' });
  setTimeout(() => { refreshRequested = false; }, 60 * 1000);
}

function errorText(raw) {
  try { return JSON.parse(raw).error; } catch (error) { return raw.slice(0, 200); }
}

function showMessage(text) {
  document.getElementById('sweepMessage').innerHTML = text ? `<div class="errors">${escapeHtml(text)}</div>` : '';
}

// ---- drawing

function draw(body) {
  const data = body.data;
  document.getElementById('sweepHead').innerHTML = headLine(body);
  if (!data || !data.run) {
    document.getElementById('sweepBody').innerHTML = '<div class="placeholder">no plan yet: run `python3 scripts/wind-down/sweep/cli.py plan …`</div>';
    return;
  }
  document.getElementById('sweepBody').innerHTML =
    runLine(data) + tiles(data) + progress(data) + byDenomTable(data) +
    `<div class="two"><div>${batchesPanel(data)}</div><div>${ladderPanel(data)}${refundedPanel(data)}</div></div>` +
    exclusionsPanel(data) + keylessPanel(data);
}

function headLine(body) {
  const age = body.fetched_at ? `live holders read ${formatRelative(body.fetched_at)}` : 'live holders not read yet';
  const refreshing = body.refreshing ? ' · refreshing…' : '';
  const error = body.last_error ? ` · <span class="t-bad">last refresh failed: ${escapeHtml(body.last_error)}</span>` : '';
  return `<div class="ops-bar"><span>${age}${refreshing}${error}</span>
    <button class="refresh" id="sweepRefresh" type="button" ${body.refreshing ? 'disabled' : ''}>Refresh live holders</button></div>`;
}

function runLine(data) {
  const run = data.run;
  const b = run.batches;
  const fees = run.operator_strd === null ? 'operator balance n/a'
    : `operator holds ${formatAmount(run.operator_strd, 6)} STRD, ${formatAmount(run.fee_estimate_ustrd, 6)} STRD needed for the pending batches`;
  return `<div class="callout">Run ${run.run_id}${run.test ? ' (test)' : ''} · floor $${run.floor_usd} · planned ${formatRelative(run.created_at)} at height ${run.height}
    · ${b.total} batches: ${b.confirmed} confirmed, ${b.submitted} submitted, ${b.failed} failed, ${b.lost} lost, ${b.pending} pending · ${fees}</div>`;
}

function tiles(data) {
  const t = data.totals;
  if (!t) return '<div class="placeholder">totals need a live holder read</div>';
  const tile = (cls, label, bucket, sub) =>
    `<div class="tile ${cls}"><div class="k">${label}</div><div class="v">${Number(bucket.addresses).toLocaleString()}</div><div class="s">addresses · $${money(bucket.usd)}${sub ? ' · ' + sub : ''}</div></div>`;
  return `<div class="tiles">${tile('sw-swept', 'Swept', t.swept)}${tile('sw-remaining', 'Remaining in plan', t.remaining)}
    ${tile('sw-refund', 'Refunded', t.refunded, 'transfer timed out; the next plan re-sweeps')}${tile('sw-excluded', 'Excluded', t.excluded)}
    ${tile('sw-floor', 'Below floor', t.below_floor, `${Number(t.unplanned.addresses).toLocaleString()} unplanned above floor ($${money(t.unplanned.usd)})`)}</div>`;
}

function progress(data) {
  const t = data.totals;
  if (!t) return '';
  const parts = [['sw-swept', 'swept', t.swept.usd], ['sw-refund', 'refunded', t.refunded.usd], ['sw-remaining', 'remaining in plan', t.remaining.usd],
                 ['sw-floor', 'below floor', t.below_floor.usd], ['sw-floor', 'unplanned', t.unplanned.usd]];
  const total = parts.reduce((sum, [, , usd]) => sum + Number(usd), 0) || 1;
  const legend = parts.map(([cls, label, usd]) => `<span><i class="${cls}"></i>${label} ${(100 * Number(usd) / total).toFixed(1)}%</span>`).join('');
  const bar = parts.map(([cls, , usd]) => `<span class="${cls}" style="width:${(100 * Number(usd) / total).toFixed(2)}%"></span>`).join('');
  return `<div class="panel"><h2>Progress <span class="sub">by USD at the plan's prices; excluded accounts not counted</span></h2>
    <div class="legend">${legend}</div><div class="progress"><div class="bar">${bar}</div></div></div>`;
}

function byDenomTable(data) {
  if (!data.by_denom) return '';
  const rows = data.by_denom.map((d) => `<tr>
    <td>${escapeHtml(d.symbol)}</td><td class="muted">${escapeHtml(d.destination)} · ${escapeHtml(d.channel)}</td>
    <td class="num">${d.swept.addresses}</td><td class="num">${formatAmount(d.swept.amount, Number(d.decimals))}</td><td class="num">$${money(d.swept.usd)}</td>
    <td class="num">${d.remaining.addresses}</td><td class="num">${formatAmount(d.remaining.amount, Number(d.decimals))}</td><td class="num">$${money(d.remaining.usd)}</td>
    <td class="num ${d.refunded_addresses !== '0' ? 't-warn' : 'muted'}">${d.refunded_addresses}</td>
    <td class="num">$${money(d.excluded_usd)}</td><td class="num">$${money(d.below_floor_usd)}</td></tr>`).join('');
  return `<div class="panel"><h2>By token <span class="sub">live balances of planned and excluded holders; swept amounts from the ledger's transfer events</span></h2>
    <div class="table-scroll"><table><thead><tr><th>Token</th><th>Destination</th><th class="num">Swept addr</th><th class="num">Swept amount</th><th class="num">Swept USD</th>
    <th class="num">Remaining addr</th><th class="num">Remaining amount</th><th class="num">Remaining USD</th><th class="num">Refunded</th><th class="num">Excluded USD</th><th class="num">Below floor USD</th></tr></thead>
    <tbody>${rows}</tbody></table></div></div>`;
}

function batchesPanel(data) {
  let currentRun = null;
  const rows = data.batches.map((b) => {
    const header = b.run_id !== currentRun ? `<tr class="zone"><td colspan="7">Run ${escapeHtml(b.run_id)}</td></tr>` : '';
    currentRun = b.run_id;
    const skips = b.skipped.length ? ` <span class="badge bad" title="${escapeHtml(b.skipped.map((s) => `${s.address}: ${s.reason}`).join('\n'))}">${b.skipped.length} skipped</span>` : '';
    const refunds = b.refunded !== '0' ? ` <span class="badge warn">${b.refunded} refunded</span>` : '';
    return `${header}<tr><td>${escapeHtml(b.batch_id)}</td><td>${b.tier ? escapeHtml(b.tier) : '<span class="muted">earlier run</span>'}</td>
      <td class="num">${b.addresses ?? '<span class="muted">n/a</span>'}</td><td class="num">${b.usd === null ? '<span class="muted">n/a</span>' : '$' + money(b.usd)}</td>
      <td>${pill(STATE_CLASS[b.state] || 'idle', b.state)}${skips}${refunds}</td><td>${b.tx_hash ? addressCell(b.tx_hash, 6, 4) : '<span class="muted">—</span>'}</td>
      <td class="muted">${b.at ? formatRelative(b.at) : '—'}</td></tr>`;
  }).join('');
  return `<div class="panel"><h2>Runs and batches <span class="sub">from state/ledger.jsonl; a batch is one tx</span></h2>
    <table><thead><tr><th>Batch</th><th>Tier</th><th class="num">Addr</th><th class="num">USD</th><th>Status</th><th>Tx</th><th>When</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function ladderPanel(data) {
  if (!data.ladder) return '';
  const rows = data.ladder.map((r, i) => `<tr><td class="${i === 0 ? 'cur' : ''}">$${r.floor}${i === 0 ? ' (current plan)' : ''}</td>
    <td class="num">${Number(r.holders).toLocaleString()}</td><td class="num">$${money(r.usd)}</td></tr>`).join('');
  return `<div class="panel"><h2>Next floor <span class="sub">live holders not yet swept, by floor</span></h2>
    <table class="ladder"><thead><tr><th>Floor</th><th class="num">Holders</th><th class="num">USD</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function refundedPanel(data) {
  if (!data.refunded.length) return '';
  const rows = data.refunded.map((r) => `<tr><td>${addressCell(r.address)}</td><td>${escapeHtml(r.denom)}</td><td class="num">${r.amount}</td>
    <td>${escapeHtml(r.channel)}</td><td>${escapeHtml(r.batch_id)}</td></tr>`).join('');
  return `<div class="panel"><h2>Refunded <span class="sub">swept, then the transfer timed out and the balance is back; the next plan re-sweeps them</span></h2>
    <table><thead><tr><th>Address</th><th>Denom</th><th class="num">Amount (base units)</th><th>Channel</th><th>Batch</th></tr></thead><tbody>${rows}</tbody></table></div>`;
}

function exclusionsPanel(data) {
  const total = data.exclusions.reduce((sum, e) => sum + Number(e.live_usd || 0), 0);
  const rows = data.exclusions.map((e) => `<tr><td>${escapeHtml(e.section)}</td><td>${addressCell(e.address, 14, 8)}</td><td>${escapeHtml(e.label)}</td>
    <td class="muted">${escapeHtml(e.reason)}</td><td class="num">${e.live_usd === null ? 'n/a' : '$' + money(e.live_usd)}</td></tr>`).join('');
  return `<details class="panel"><summary><h2>Excluded <span class="sub">from sweep/exclusions.json · ${data.exclusions.length} addresses · $${money(String(total))} live</span></h2></summary>
    <table><thead><tr><th>Section</th><th>Address</th><th>Label</th><th>Reason</th><th class="num">Live USD</th></tr></thead><tbody>${rows}</tbody></table></details>`;
}

function keylessPanel(data) {
  const k = data.keyless;
  if (!k) return '';
  return `<details class="panel"><summary><h2>Keyless tier <span class="sub">${k.addresses} addresses · $${money(k.usd)} · ${k.swept} swept · never signed; swept last so owners get the most time</span></h2></summary>
    <div class="note">Accounts with no pubkey and sequence 0 at plan time. They sit in the plan's last tier; one that signs a tx before its batch runs moves into the main tier at the next plan.</div></details>`;
}

function money(text) {
  return Number(text).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

})();
```

`scripts/wind-down/dashboard/static/style.css`, appended:

```css
/* sweep */
#view-sweep .callout { margin: 0 0 14px; padding: 8px 12px; border-radius: 8px; background: var(--warn-bg); color: var(--warn); font-weight: 600; }
#view-sweep .tile { border-top: 3px solid var(--line); }
#view-sweep .sw-swept { border-top-color: var(--s-pool); } #view-sweep .sw-remaining { border-top-color: var(--s-staked); }
#view-sweep .sw-refund { border-top-color: var(--s-unbond); } #view-sweep .sw-excluded { border-top-color: var(--s-flight); } #view-sweep .sw-floor { border-top-color: var(--idle-bg); }
#view-sweep .bar .sw-swept, #view-sweep .legend .sw-swept { background: var(--s-pool); }
#view-sweep .bar .sw-remaining, #view-sweep .legend .sw-remaining { background: var(--s-staked); }
#view-sweep .bar .sw-refund, #view-sweep .legend .sw-refund { background: var(--s-unbond); }
#view-sweep .bar .sw-floor, #view-sweep .legend .sw-floor { background: var(--idle); }
#view-sweep .progress { padding: 10px 12px; } #view-sweep .progress .bar { height: 18px; }
#view-sweep .two { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; } #view-sweep .two .panel { margin-bottom: 14px; }
#view-sweep .ladder td.cur { font-weight: 650; color: var(--accent); }
#view-sweep .ops-bar { display: flex; align-items: center; gap: 14px; padding: 10px 12px; }
```

- [ ] **Step 7: Start the server and look at the tab**

Run: `python3 scripts/wind-down/dashboard/server.py` and open `http://localhost:8787/#sweep`. With no plan the tab says so. Create a throwaway plan to see the layout without the network: in a Python shell, build a `planner.Plan` as in `test_sweep_tab.make_plan()` and `planner.write_plan(plan=plan)`; reload; delete `scripts/wind-down/sweep/state/plan.json` and the batch files afterwards (do not commit them). The live read against Polkachu takes a minute or two; the head line shows "refreshing…" while it runs. Stop the server.

- [ ] **Step 8: README section and the full dashboard suite**

Add to `scripts/wind-down/dashboard/README.md`, after the Multisig section:

```markdown
## Sweep tab

The seventh tab (`#sweep`): the holder sweep's progress, from `scripts/wind-down/sweep/` (see its README). Like Ops
and Multisig it polls its own route: `GET /api/sweep` composes the `sweep` collector's snapshot (every holder of every
sweep denom, read in bulk with `denom_owners_by_query`; the operator's STRD) with `sweep/state/plan.json`,
`sweep/state/ledger.jsonl` and `sweep/exclusions.json` read from disk on every request, so a batch the runner just
confirmed shows within ten seconds while the live read stays on demand: the "Refresh live holders" button, or
automatically when the snapshot is missing or older than ten minutes while the tab is open (the collector's interval
is 1,800 s as a backstop). `sweep_tab.compose()` is pure and tested in `test_sweep_tab.py`.

Per planned address: `swept` (a confirmed `sweep_transfer` and none of the transferred denoms held now), `refunded`
(transferred, and a transferred denom is held again: the 24 h timeout refunded it; the next `plan` re-sweeps it),
`remaining` (no confirmed transfer yet). Tiles: swept, remaining, refunded, excluded (the exclusions file, live USD),
below floor (live holders under the plan's floor; the sub-line counts holders above the floor that are not in the
plan, which appeared or were refunded since). The by-token table, the runs-and-batches table (every batch of the
current plan plus earlier runs' batches from the ledger; a red `N skipped` badge means the chain disagreed with the
planner), the next-floor ladder (live holders still to sweep at the plan's floor and at $10, $5, $1, $0), the refunded
list, and the collapsed exclusions and keyless panels. Without a live snapshot the plan and ledger parts render and the
live-dependent ones say so.
```

Run: `python3 -m unittest discover -s scripts/wind-down/dashboard`
Expected: all OK (the existing suites plus 7 new).

- [ ] **Step 9: Commit**

```bash
git add scripts/wind-down/dashboard
git commit -m "wind-down dashboard: the Sweep tab (live holders on demand, plan and ledger from disk)"
```

---

### Task 7: Retire the export builder, update the ops plan and docs

**Files:**
- Delete: `scripts/wind-down/build_sweep_batches.py`, `scripts/wind-down/test_build_sweep_batches.py`
- Modify: `scripts/wind-down/dashboard/ops/plan.json`, `scripts/wind-down/dashboard/test_ops.py`, `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`
- Test: `scripts/wind-down/dashboard/test_ops.py`

**Interfaces:**
- Consumes: the CLI's command names from Task 5's README (`cli.py plan --floor-usd … --canary 3`, `cli.py plan --test`, `cli.py run`, `cli.py status`, `cli.py resolve`); `sweep/config.py` and `sweep/exclusions.json` as the places the finalize step points at.
- Depends on: Tasks 1-4
- Review: no

- [ ] **Step 1: Write the failing test**

Append to `scripts/wind-down/dashboard/test_ops.py` (inside the existing test class that loads the plan, or a new class using `ops.load_plan()`):

```python
class SweepStepsTests(unittest.TestCase):
    def setUp(self) -> None:
        plan = ops.load_plan()
        self.steps = {step["id"]: step for day in plan["days"] for window in day["windows"] for step in window["steps"]}
        self.days = {step["id"]: day["date"] for day in plan["days"] for window in day["windows"] for step in window["steps"]}

    def test_sweep_steps_name_the_runner(self) -> None:
        self.assertNotIn("sweep-export", self.steps)
        self.assertIn("cli.py plan --floor-usd", self.steps["sweep-plan"]["command"])
        self.assertIn("cli.py run", self.steps["sweep-submit"]["command"])
        self.assertIn("sweep/exclusions.json", self.steps["vote-finalize-sweep"]["detail"])
        self.assertNotIn("build_sweep_batches", json.dumps(self.steps))

    def test_sweep_test_step_sits_after_the_upgrade_and_before_the_sweep(self) -> None:
        self.assertIn("cli.py plan --test", self.steps["sweep-test"]["command"])
        self.assertIn("stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg", self.steps["sweep-test"]["text"])
        self.assertEqual(self.days["sweep-test"], "2026-10-13")
```

`import json` at the top of the test file if it is not there.

- [ ] **Step 2: Run the test to verify it fails**

Run: `python3 -m unittest discover -s scripts/wind-down/dashboard -p test_ops.py -k Sweep -v`
Expected: FAIL (`sweep-export` still present, no `sweep-plan`, no `sweep-test`).

- [ ] **Step 3: Edit plan.json**

In the `2026-11-10` day, window labelled "The sweep":

- Replace the `sweep-export` step with:

```json
{
 "id": "sweep-plan",
 "text": "Plan the sweep from live state at the announced floor with three canaries: every holder at or above the floor, every denom on the sweep list (eleven stTokens and ustrd to Osmosis, whitelisted vouchers back one hop); read the excluded, skipped and ladder tables before running.",
 "command": "python3 scripts/wind-down/sweep/cli.py plan --floor-usd <FLOOR> --canary 3",
 "expect": "RESULT: PLANNED with the run id and batch count. The excluded table is exactly sweep/exclusions.json (F5 multisig, relayer keys); the skipped table is escrows, module accounts, ICAs and contracts; the ladder shows holders and USD at the floor and below it. Re-run at a lower floor after each run completes: swept holders have no balance and never reappear; a refunded one does."
}
```

- Replace the `sweep-submit` step with:

```json
{
 "id": "sweep-submit",
 "text": "Run the plan: preflight (chain id, v35 binary, stride-sweeper key in the test keyring, plan age, batch-file hashes, exclusions, channels OPEN, fee balance), then per batch a gas simulation, the signed broadcast and the poll for the result, each appended to sweep/state/ledger.jsonl. Canary tier first, then main by value, keyless last. Watch the Sweep tab; any sweep_skipped event stops the run.",
 "command": "python3 scripts/wind-down/sweep/cli.py run",
 "expect": "Every batch prints `confirmed at <height>: N transfers, 0 skipped`. A skip means the planner and the chain disagree: inspect the ledger line, fix the rule, re-plan. A tx not found within three minutes: `cli.py resolve`. A transfer that times out (24 h) refunds the holder: the Sweep tab lists it under Refunded and the next `plan` includes it again. Commit sweep/state after each session."
}
```

- In the `vote-finalize-sweep` step (`2026-10-06`), replace the `detail` with:

```
scripts/wind-down/sweep/config.py holds the denom list (the eleven stTokens and ustrd to Osmosis, the seven whitelisted vouchers back one hop), the rough USD prices behind the floor, and the gas budget per batch; scripts/wind-down/sweep/exclusions.json holds the F5 multisig and the relayer keys with a reason each (reviewed like the plan). The keyless accounts are swept in a final tier, not excluded. Open items folded in here: whether the two Axelar USDC channels (channel-11, channel-69, about 2 USDC) join SweepUnwindChannels before the binary is cut; confirming the legacy claim module's 2022 airdrops are expired; and notifying the interchain-account and 32-byte-address stToken holders the sweep cannot reach (2.3k stATOM in one ICA) so they move out before the halt.
```

- In the `2026-10-13` day ("After the drain"), append to its window's steps:

```json
{
 "id": "sweep-test",
 "text": "Mainnet rehearsal of the sweep runner on our own address stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg (fund it first with a little stATOM, some STRD and an ATOM voucher so all three destination rules run): plan the one-address batch, run it, then check the same bytes under osmo1… hold the stATOM and STRD and under cosmos1… the ATOM, and that the Sweep tab shows the batch confirmed. This batch calibrates gas per transfer for the real run.",
 "command": "python3 scripts/wind-down/sweep/cli.py plan --test && python3 scripts/wind-down/sweep/cli.py run",
 "expect": "One batch, tier test, `confirmed at <height>: 3 transfers, 0 skipped`; the ledger has a submitted and a confirmed line; the Sweep tab shows run 1 with 1 of 1 confirmed. Commit sweep/state."
}
```

Keep the JSON valid (`python3 -c "import json; json.load(open('scripts/wind-down/dashboard/ops/plan.json'))"`).

- [ ] **Step 4: Delete the old builder and point the design spec at the new one**

```bash
git rm scripts/wind-down/build_sweep_batches.py scripts/wind-down/test_build_sweep_batches.py
```

In `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md`, the "Ops scripts (PRs 5 and 6)" bullet that begins "`build_sweep_batches.py` applies the on-chain skip rules" becomes:

```
- The sweep runner (`scripts/wind-down/sweep/`, design in `2026-10-09-wind-down-sweep-runner-design.md`) replaced
  `build_sweep_batches.py` on 2026-10-09: it plans from live state, signs and submits each batch with the sweep
  operator key, keeps an append-only ledger, and feeds the dashboard's Sweep tab. `coverage_check.py` is the §10 check
  over an export and the vault's Osmosis balances, unit-tested against a synthetic export.
```

Leave the other mentions of the builder in that spec's §7 and §9 as history, except the sentence "(size set from the gas measurement; the builder defaults to 100)" in step 8 of §9, which becomes "(packed under a gas budget by the runner, simulated before each broadcast)".

Check `grep -rn build_sweep_batches scripts docs/wind-down CHANGELOG.md` and leave CHANGELOG and `docs/superpowers/plans/` untouched (history).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `python3 -m unittest discover -s scripts/wind-down/dashboard` and `cd scripts/wind-down && python3 -m unittest discover -s sweep -t .`
Expected: all OK.

- [ ] **Step 6: Commit**

```bash
git add -A scripts/wind-down docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md
git commit -m "wind-down sweep: retire the export builder; ops plan and spec point at the runner"
```
