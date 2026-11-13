"""Prints a zone's rest-join amounts from a fresh Pools read, for the close tx on the Multisig tab.

The close's generate step runs this right before it writes the unsigned tx, so every rest join uses the escrow and
vault balances of that moment, not the dashboard's last snapshot: a route pool's allocation moves with its channel's
escrow, the canonical pool's with the vault. It prints {contract: amount} as JSON on stdout and each amount on
stderr, and exits non-zero with nothing on stdout when the fresh read disagrees with the pools the tx lists.

Usage, from the repo root: python3 scripts/wind-down/dashboard/pool_amounts.py <chain_id> <contract>...
"""

import argparse
import json
import sys

import config
import multisig
import pools


def main() -> None:
    parser = argparse.ArgumentParser(description="A zone's rest-join amounts from a fresh Pools read.")
    parser.add_argument("chain_id", choices=sorted(config.ZONES_BY_CHAIN_ID))
    parser.add_argument("contracts", nargs="+", help="the pools the close tx rest-joins, as it lists them")
    args = parser.parse_args()

    print(f"Reading the pools fresh for {args.chain_id}...", file=sys.stderr)
    rest_joins = multisig.rest_join_amounts(chain_id=args.chain_id, pools_data=pools.collect())
    if rest_joins.reason:
        sys.exit(f"{args.chain_id}: {rest_joins.reason}")

    # The tx was written from the dashboard's snapshot: a pool it lists must still need its rest join, and no other.
    if set(rest_joins.amounts) != set(args.contracts):
        sys.exit(
            f"{args.chain_id}: the fresh read rest-joins {sorted(rest_joins.amounts)}, the tx lists {sorted(args.contracts)}: "
            "refresh the dashboard and copy the close tx again"
        )
    if any(amount <= 0 for amount in rest_joins.amounts.values()):
        sys.exit(f"{args.chain_id}: a rest join is not positive: {rest_joins.amounts}")

    for contract in args.contracts:
        print(f"  {contract}: {rest_joins.amounts[contract]}", file=sys.stderr)
    print(json.dumps({contract: str(rest_joins.amounts[contract]) for contract in args.contracts}))


if __name__ == "__main__":
    main()
