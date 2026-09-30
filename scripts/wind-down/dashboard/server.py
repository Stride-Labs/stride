"""Local, read-only wind-down dashboard server.

    python3 scripts/wind-down/dashboard/server.py        # then open http://localhost:8787

Each registered collector is refreshed on its own interval in a background thread; the page polls the
cached snapshots, so the browser never talks to a chain.
"""

import datetime
import http.server
import json
import mimetypes
import pathlib
import sys
import threading
import time
import traceback
from collections.abc import Callable
from typing import Any

import channels
import config
import validators
import funds

STATIC_DIR = pathlib.Path(__file__).parent / "static"

# Tab name -> collect().
COLLECTORS: dict[str, Callable[[], dict[str, Any]]] = {
    "channels": channels.collect,
    "validators": validators.collect,
    "funds": funds.collect,
}


class TabCache:
    """The latest snapshot for one tab, refreshed by its own thread."""

    def __init__(
        self, name: str, collect: Callable[[], dict[str, Any]], interval_seconds: int
    ) -> None:
        self.name = name
        self._collect = collect
        self._interval_seconds = interval_seconds
        self._lock = threading.Lock()
        self._wake = threading.Event()
        self._snapshot: dict[str, Any] | None = None
        self._refreshing = False
        self._last_error: str | None = None

    def start(self) -> None:
        threading.Thread(
            target=self._run, name=f"refresh-{self.name}", daemon=True
        ).start()

    def request_refresh(self) -> None:
        self._wake.set()

    def view(self) -> dict[str, Any]:
        """The API body: the snapshot, or a "loading" body until the first refresh succeeds."""
        with self._lock:
            if self._snapshot is None:
                return {
                    "loading": True,
                    "refreshing": self._refreshing,
                    "last_error": self._last_error,
                }
            return {**self._snapshot, "refreshing": self._refreshing}

    def _run(self) -> None:
        while True:
            self._wake.clear()
            started = time.monotonic()
            self._refresh()
            # The interval is start-to-start, so a slow refresh is not followed by a full extra wait.
            self._wake.wait(timeout=max(0.0, self._interval_seconds - (time.monotonic() - started)))

    def _refresh(self) -> None:
        with self._lock:
            self._refreshing = True
        started = time.monotonic()

        try:
            data = self._collect()
        except Exception:  # noqa: BLE001 - keep serving the previous snapshot whatever went wrong
            traceback.print_exc(file=sys.stderr)
            with self._lock:
                self._last_error = traceback.format_exc().strip().splitlines()[-1]
                self._refreshing = False
            return

        snapshot = {
            "fetched_at": datetime.datetime.now(datetime.UTC).isoformat(),
            "duration_seconds": round(time.monotonic() - started, 2),
            "data": data,
        }
        with self._lock:
            self._snapshot = snapshot
            self._last_error = None
            self._refreshing = False


CACHES: dict[str, TabCache] = {
    name: TabCache(
        name=name,
        collect=collect,
        interval_seconds=config.REFRESH_INTERVAL_SECONDS[name],
    )
    for name, collect in COLLECTORS.items()
}


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        path = self.path.split("?")[0]
        if path == "/api/config":
            self._send_json(
                status=200, body={"intervals": config.REFRESH_INTERVAL_SECONDS}
            )
        elif path.startswith("/api/"):
            self._get_snapshot(tab=path.removeprefix("/api/"))
        else:
            self._send_static(path=path)

    def do_POST(self) -> None:
        path = self.path.split("?")[0]
        prefix = "/api/refresh/"
        cache = (
            CACHES.get(path.removeprefix(prefix)) if path.startswith(prefix) else None
        )
        if cache is None:
            self._send_json(status=404, body={"error": "unknown tab"})
            return

        cache.request_refresh()
        self._send_json(status=202, body={"refreshing": True})

    def log_message(self, format: str, *args: Any) -> None:  # noqa: A002 - signature fixed by the base class
        """Silence per-request logging; the page polls every few seconds."""

    def _get_snapshot(self, tab: str) -> None:
        cache = CACHES.get(tab)
        if cache is None:
            self._send_json(status=404, body={"error": "not built"})
            return

        body = cache.view()
        self._send_json(status=503 if body.get("loading") else 200, body=body)

    def _send_static(self, path: str) -> None:
        relative = "index.html" if path == "/" else path.lstrip("/")
        target = (STATIC_DIR / relative).resolve()
        # Refuse anything that escapes static/ (e.g. "/../server.py").
        if not target.is_relative_to(STATIC_DIR.resolve()) or not target.is_file():
            self._send_json(status=404, body={"error": "not found"})
            return

        content_type = (
            mimetypes.guess_type(target.name)[0] or "application/octet-stream"
        )
        self._send(status=200, content_type=content_type, payload=target.read_bytes())

    def _send_json(self, status: int, body: dict[str, Any]) -> None:
        self._send(
            status=status,
            content_type="application/json",
            payload=json.dumps(body).encode(),
        )

    def _send(self, status: int, content_type: str, payload: bytes) -> None:
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(payload)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(payload)


def main() -> None:
    for cache in CACHES.values():
        cache.start()

    server = http.server.ThreadingHTTPServer((config.HOST, config.PORT), Handler)
    print(f"Wind-down dashboard on http://localhost:{config.PORT}", file=sys.stderr)
    server.serve_forever()


if __name__ == "__main__":
    main()
