"""The package's only I/O: Stride REST reads, the strided subprocess, the clock. Tests patch these functions."""

import datetime
import http.client
import json
import random
import subprocess
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from typing import Any

from sweep import config

SDK_NOT_FOUND_CODE = 5
HTTP_TOO_MANY_REQUESTS = 429
HTTP_SERVER_ERROR = 500


class ChainError(Exception):
    """A REST call failed (network, HTTP error, bad JSON)."""


class NotFound(ChainError):
    """The chain answered that the thing does not exist: an unknown account, a tx not yet indexed."""


@dataclass(frozen=True)
class CommandResult:
    returncode: int
    stdout: str
    stderr: str


def rest_get(
    path: str, params: dict[str, str] | None = None, height: int | None = None, base: str = config.REST
) -> Any:
    """GET a Stride REST path and decode the JSON body. HTTP 429, any 5xx, network and transport errors and
    undecodable bodies are
    retried after each of REST_BACKOFF_SECONDS (with jitter), then raise ChainError; a missing thing is NotFound at
    once. With `height` the node answers from that block's state (x-cosmos-block-height). `base` picks the host."""
    query = f"?{urllib.parse.urlencode(params)}" if params else ""
    url = f"{base}{path}{query}"
    headers = {"x-cosmos-block-height": str(height)} if height is not None else {}

    for backoff in (*config.REST_BACKOFF_SECONDS, None):
        try:
            return _fetch_json(url=url, headers=headers)
        except urllib.error.HTTPError as error:
            chain_error = _http_error(path=path, error=error)
            if isinstance(chain_error, NotFound) or not _is_retryable_status(code=error.code) or backoff is None:
                raise chain_error from error
        except (urllib.error.URLError, TimeoutError, http.client.HTTPException, ConnectionError) as error:
            # URLError: no connection; HTTPException: a response cut short (IncompleteRead); ConnectionError: reset
            if backoff is None:
                raise ChainError(f"{path}: {error}") from error
        except json.JSONDecodeError as error:  # a 200 that is an HTML proxy page, not the node
            if backoff is None:
                raise ChainError(f"{path}: {error}") from error
        sleep(backoff * random.uniform(*config.REST_BACKOFF_JITTER))
    raise AssertionError("unreachable: the last attempt returns or raises")


def rest_get_all_pages(
    path: str, key: str, params: dict[str, str] | None = None, base: str = config.REST
) -> list[Any]:
    """Collect `response[key]` across Cosmos `pagination.key` / `next_key` pages."""
    items: list[Any] = []
    page_params = {"pagination.limit": str(config.PAGE_SIZE), **(params or {})}
    while True:
        response = rest_get(path=path, params=page_params, base=base)
        items.extend(response.get(key) or [])
        next_key = (response.get("pagination") or {}).get("next_key")
        if not next_key:
            return items
        page_params = {**page_params, "pagination.key": next_key}


def latest_height(base: str = config.REST) -> int:
    body = rest_get(path="/cosmos/base/tendermint/v1beta1/blocks/latest", base=base)
    return int(body["block"]["header"]["height"])


def strided(args: list[str]) -> CommandResult:
    """Run the strided binary with the given arguments; the caller decides what a non-zero exit means."""
    try:
        completed = subprocess.run(
            [config.STRIDED_BINARY, *args],
            capture_output=True,
            text=True,
            check=False,
            timeout=config.STRIDED_TIMEOUT_SECONDS,
        )
    except subprocess.TimeoutExpired:
        return CommandResult(
            returncode=-1, stdout="", stderr=f"strided timed out after {config.STRIDED_TIMEOUT_SECONDS}s"
        )
    return CommandResult(returncode=completed.returncode, stdout=completed.stdout, stderr=completed.stderr)


def now() -> datetime.datetime:
    return datetime.datetime.now(datetime.UTC)


def sleep(seconds: float) -> None:
    time.sleep(seconds)


def _is_retryable_status(code: int) -> bool:
    return code == HTTP_TOO_MANY_REQUESTS or code >= HTTP_SERVER_ERROR


def _fetch_json(url: str, headers: dict[str, str]) -> Any:
    request = urllib.request.Request(url, headers={"User-Agent": config.USER_AGENT, **headers})
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
