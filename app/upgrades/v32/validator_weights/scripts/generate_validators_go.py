"""
Generates validators.go from new_validators.csv and all_validators.csv.

Output: app/upgrades/v32/validators.go
"""

import csv
import sys
from collections import defaultdict
from pathlib import Path

WEIGHTS_DIR = Path(__file__).parent.parent
NEW_VALIDATORS_PATH = WEIGHTS_DIR / "new_validators.csv"
ALL_VALIDATORS_PATH = WEIGHTS_DIR / "all_validators.csv"
OUTPUT_PATH = WEIGHTS_DIR.parent / "validators.go"


def load_csv_by_chain(path: Path) -> dict[str, list[dict]]:
    result: dict[str, list[dict]] = defaultdict(list)
    with path.open() as f:
        for row in csv.DictReader(f):
            result[row["chain_id"]].append(row)
    return result


def build_go_source(
    new_vals: dict[str, list[dict]],
    all_vals: dict[str, list[dict]],
) -> str:
    lines = []
    lines.append("package v32")
    lines.append("")
    lines.append("type ValidatorConfig struct {")
    lines.append("\tAddress string")
    lines.append("\tName    string")
    lines.append("}")
    lines.append("")
    lines.append("type WeightConfig struct {")
    lines.append("\tAddress string")
    lines.append("\tWeight  uint64")
    lines.append("}")
    lines.append("")

    lines.append("var NewValidators = map[string][]ValidatorConfig{")
    for chain_id in sorted(new_vals):
        lines.append(f'\t"{chain_id}": {{')
        for v in new_vals[chain_id]:
            lines.append(f'\t\t{{Address: "{v["validator_address"]}", Name: "{v["validator_name"]}"}},')
        lines.append("\t},")
    lines.append("}")
    lines.append("")

    lines.append("var TargetWeights = map[string][]WeightConfig{")
    for chain_id in sorted(all_vals):
        lines.append(f'\t"{chain_id}": {{')
        for v in all_vals[chain_id]:
            w = int(v["target_weight"])
            lines.append(f'\t\t{{Address: "{v["validator_address"]}", Weight: {w}}},')
        lines.append("\t},")
    lines.append("}")
    lines.append("")

    return "\n".join(lines)


def main() -> int:
    print(f"Loading {NEW_VALIDATORS_PATH} ...")
    new_vals = load_csv_by_chain(NEW_VALIDATORS_PATH)
    print(f"Found {sum(len(v) for v in new_vals.values())} new validators across {len(new_vals)} chains")

    print(f"Loading {ALL_VALIDATORS_PATH} ...")
    all_vals = load_csv_by_chain(ALL_VALIDATORS_PATH)
    print(f"Found {sum(len(v) for v in all_vals.values())} total validators across {len(all_vals)} chains")

    source = build_go_source(new_vals, all_vals)
    OUTPUT_PATH.write_text(source)
    print(f"Wrote {OUTPUT_PATH}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
