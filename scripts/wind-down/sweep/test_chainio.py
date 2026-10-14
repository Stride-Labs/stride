"""chainio: the REST wrapper's error mapping and pagination, with urllib patched out."""

import io
import json
import subprocess
import unittest
import urllib.error
from unittest import mock

from sweep import chainio


def _http_error(code: int, body: dict) -> urllib.error.HTTPError:
    return urllib.error.HTTPError(url="u", code=code, msg="m", hdrs=None, fp=io.BytesIO(json.dumps(body).encode()))


class StridedTests(unittest.TestCase):
    def test_timeout_becomes_a_failed_command_result(self) -> None:
        expired = subprocess.TimeoutExpired(cmd="strided", timeout=1)
        with mock.patch.object(chainio.subprocess, "run", side_effect=expired):
            result = chainio.strided(args=["version"])
        self.assertEqual(result.returncode, -1)
        self.assertIn("timed out", result.stderr)


class RestGetTests(unittest.TestCase):
    def test_404_and_sdk_not_found_bodies_raise_not_found(self) -> None:
        for code, body in ((404, {"message": "no"}), (500, {"code": 5, "message": "account stride1x not found"})):
            with self.subTest(code=code), mock.patch.object(chainio.urllib.request, "urlopen", side_effect=_http_error(code, body)):
                with self.assertRaises(chainio.NotFound):
                    chainio.rest_get(path="/cosmos/auth/v1beta1/accounts/stride1x")

    def test_other_http_errors_raise_chain_error(self) -> None:
        with mock.patch.object(chainio.urllib.request, "urlopen", side_effect=_http_error(502, {"message": "bad gateway"})), \
                mock.patch.object(chainio, "sleep"):
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


class _Raw:
    def __init__(self, raw: bytes) -> None:
        self._raw = raw

    def __enter__(self) -> io.BytesIO:
        return io.BytesIO(self._raw)

    def __exit__(self, *exc: object) -> None:
        return None


class _Body:
    def __init__(self, payload: dict) -> None:
        self._raw = json.dumps(payload).encode()

    def __enter__(self) -> io.BytesIO:
        return io.BytesIO(self._raw)

    def __exit__(self, *exc: object) -> None:
        return None


class BackoffTests(unittest.TestCase):
    def test_429_then_success_sleeps_with_backoff(self) -> None:
        outcomes = [_http_error(429, {}), _http_error(503, {}), urllib.error.URLError("reset"), _Body({"ok": 1})]
        with mock.patch.object(chainio.urllib.request, "urlopen", side_effect=outcomes), \
                mock.patch.object(chainio, "sleep") as sleep, \
                mock.patch.object(chainio.random, "uniform", return_value=1.0):
            body = chainio.rest_get(path="/x")
        self.assertEqual(body, {"ok": 1})
        self.assertEqual([call.args[0] for call in sleep.call_args_list], [1, 2, 4])

    def test_gives_up_after_five_attempts(self) -> None:
        with mock.patch.object(chainio.urllib.request, "urlopen", side_effect=lambda *a, **k: (_ for _ in ()).throw(TimeoutError())) as urlopen, \
                mock.patch.object(chainio, "sleep") as sleep, \
                mock.patch.object(chainio.random, "uniform", return_value=1.0):
            with self.assertRaises(chainio.ChainError):
                chainio.rest_get(path="/x")
        self.assertEqual(urlopen.call_count, 5)
        self.assertEqual([call.args[0] for call in sleep.call_args_list], [1, 2, 4, 8])

    def test_html_200_is_retried_then_decoded(self) -> None:
        outcomes = [_Raw(b"<html>proxy</html>"), _Body({"ok": 1})]
        with mock.patch.object(chainio.urllib.request, "urlopen", side_effect=outcomes), \
                mock.patch.object(chainio, "sleep") as sleep, \
                mock.patch.object(chainio.random, "uniform", return_value=1.0):
            body = chainio.rest_get(path="/x")
        self.assertEqual(body, {"ok": 1})
        self.assertEqual([call.args[0] for call in sleep.call_args_list], [1])

    def test_not_found_and_client_errors_are_not_retried(self) -> None:
        for code, body, expected in ((404, {}, chainio.NotFound), (500, {"code": 5, "message": "not found"}, chainio.NotFound),
                                     (400, {"message": "bad"}, chainio.ChainError)):
            with self.subTest(code=code), mock.patch.object(chainio.urllib.request, "urlopen", side_effect=_http_error(code, body)) as urlopen, \
                    mock.patch.object(chainio, "sleep") as sleep:
                with self.assertRaises(expected):
                    chainio.rest_get(path="/x")
            self.assertEqual(urlopen.call_count, 1)
            sleep.assert_not_called()

    def test_height_sets_the_pinned_header_only_when_given(self) -> None:
        with mock.patch.object(chainio.urllib.request, "urlopen", return_value=_Body({})) as urlopen:
            chainio.rest_get(path="/x", height=123)
            chainio.rest_get(path="/x")
        pinned, unpinned = (call.args[0] for call in urlopen.call_args_list)
        self.assertEqual(pinned.get_header("X-cosmos-block-height"), "123")
        self.assertIsNone(unpinned.get_header("X-cosmos-block-height"))


if __name__ == "__main__":
    unittest.main()
