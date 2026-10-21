import io
import json
import unittest
from unittest import mock

import pool_amounts
import test_multisig as fixtures


def run(*contracts: str) -> tuple[str, str]:
    """Runs the CLI on cosmoshub-4 against the fixture pools as the fresh read; returns (stdout, stderr)."""
    stdout, stderr = io.StringIO(), io.StringIO()
    argv = ["pool_amounts.py", fixtures.COSMOS, *contracts]
    with (
        mock.patch("sys.argv", argv),
        mock.patch("pools.collect", return_value=fixtures.fake_pools()),
        mock.patch("sys.stdout", stdout),
        mock.patch("sys.stderr", stderr),
    ):
        pool_amounts.main()
    return stdout.getvalue(), stderr.getvalue()


class PoolAmountsTest(unittest.TestCase):
    def test_prints_each_listed_pools_rest_join_from_the_fresh_read(self) -> None:
        stdout, stderr = run(fixtures.CANONICAL_CONTRACT, fixtures.ROUTE_CONTRACT)

        self.assertEqual(json.loads(stdout), {fixtures.CANONICAL_CONTRACT: "4000000", fixtures.ROUTE_CONTRACT: "2000000"})
        self.assertIn(f"{fixtures.ROUTE_CONTRACT}: 2000000", stderr)

    def test_a_pool_list_that_disagrees_with_the_fresh_read_exits_with_nothing_on_stdout(self) -> None:
        with self.assertRaises(SystemExit) as raised:
            run(fixtures.CANONICAL_CONTRACT)

        self.assertIn("refresh the dashboard and copy the close tx again", str(raised.exception.code))

    def test_a_close_that_is_not_ready_on_the_fresh_read_exits(self) -> None:
        unjoined = fixtures.fake_pools(pools=fixtures.unjoined_pools())
        with mock.patch("pools.collect", return_value=unjoined), self.assertRaises(SystemExit) as raised:
            with mock.patch("sys.argv", ["pool_amounts.py", fixtures.COSMOS, fixtures.CANONICAL_CONTRACT]):
                pool_amounts.main()

        self.assertIn("does not show the test join yet", str(raised.exception.code))


if __name__ == "__main__":
    unittest.main()
