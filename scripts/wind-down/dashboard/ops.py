"""The Ops tab's data: the dated wind-down checklist (`ops/plan.json`) and the team's ticks (`ops/status.json`).

Both files are read from disk on every call, so an edit to the plan shows on reload. Ticks are written atomically
with sorted keys so each one is a one-line diff, and the file is committed like any other.
"""

import datetime
import zoneinfo
import json
import os
import pathlib
import tempfile
import threading
from dataclasses import dataclass
from typing import Any

OPS_DIR = pathlib.Path(__file__).parent / "ops"
PLAN_PATH = OPS_DIR / "plan.json"
STATUS_PATH = OPS_DIR / "status.json"

MAX_NAME_LENGTH = 80

# Serialises the read-modify-write in record_check; the server handles requests on threads.
_status_lock = threading.Lock()


class InvalidCheckError(ValueError):
    """The body of a tick request is malformed."""


class UnknownStepError(InvalidCheckError):
    """The tick names an id that is not in the plan."""


@dataclass(frozen=True)
class CheckRequest:
    check_id: str
    done: bool
    by: str


def load_plan(plan_path: pathlib.Path | None = None) -> dict[str, Any]:
    return _read_json(path=plan_path or PLAN_PATH)


def load_status(status_path: pathlib.Path | None = None) -> dict[str, Any]:
    return _read_json(path=status_path or STATUS_PATH)


PLAN_TIMEZONE = zoneinfo.ZoneInfo("America/New_York")


def today() -> str:
    """The US Eastern date, the clock the plan's blocks are written in (the team works in ET)."""
    return datetime.datetime.now(PLAN_TIMEZONE).date().isoformat()


def step_ids(plan: dict[str, Any]) -> list[str]:
    """Every tickable id in plan order: each step's own id, then `<id>:<zone>` for each of its zones."""
    ids: list[str] = []
    for day in plan["days"]:
        for window in day["windows"]:
            for step in window["steps"]:
                ids.append(step["id"])
                ids.extend(f"{step['id']}:{zone}" for zone in step.get("zones", []))
    return ids


def parse_check_request(body: Any) -> CheckRequest:
    """Validate a `POST /api/ops/check` body; raises InvalidCheckError naming the first bad field."""
    if not isinstance(body, dict):
        raise InvalidCheckError("body must be a JSON object")

    check_id = body.get("id")
    if not isinstance(check_id, str) or not check_id:
        raise InvalidCheckError("id must be a non-empty string")

    done = body.get("done")
    if not isinstance(done, bool):
        raise InvalidCheckError("done must be a boolean")

    by = body.get("by")
    if not isinstance(by, str) or not by.strip():
        raise InvalidCheckError("by must be a non-empty string")
    if len(by) > MAX_NAME_LENGTH:
        raise InvalidCheckError(f"by must be at most {MAX_NAME_LENGTH} characters")

    return CheckRequest(check_id=check_id, done=done, by=by.strip())


def record_check(
    check_id: str,
    done: bool,
    by: str,
    plan_path: pathlib.Path | None = None,
    status_path: pathlib.Path | None = None,
) -> dict[str, Any]:
    """Tick (or untick) one id and return the full updated status. Raises UnknownStepError for an id not in the plan."""
    if check_id not in step_ids(plan=load_plan(plan_path=plan_path)):
        raise UnknownStepError(f"unknown step id: {check_id}")

    target = status_path or STATUS_PATH
    with _status_lock:
        status = load_status(status_path=target)
        if done:
            status[check_id] = {"done": True, "at": datetime.datetime.now(datetime.UTC).isoformat(), "by": by}
        else:
            status.pop(check_id, None)
        _write_json_atomic(path=target, body=status)

    return status


def _read_json(path: pathlib.Path) -> dict[str, Any]:
    return json.loads(path.read_text())


def _write_json_atomic(path: pathlib.Path, body: dict[str, Any]) -> None:
    # The temp file lives beside the target so os.replace stays on one filesystem and readers never see a partial file.
    with tempfile.NamedTemporaryFile(mode="w", dir=path.parent, suffix=".tmp", delete=False) as handle:
        handle.write(json.dumps(body, indent=2, sort_keys=True) + "\n")
    os.replace(src=handle.name, dst=path)
