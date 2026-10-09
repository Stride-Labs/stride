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
