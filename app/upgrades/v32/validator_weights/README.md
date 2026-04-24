# Validator Weights (v32 Upgrade)

This directory contains the data and scripts used to compute the validator weight
updates for the v32 upgrade handler.

## Process

1. **Fetch current on-chain weights** — `scripts/fetch_current_weights.py` queries the
   Stride host zone API and writes `current_weights.csv`.

2. **Parse target weights from spreadsheet** — The raw spreadsheet exports live in
   `target/raw/` (one CSV per chain). `scripts/build_target_weights.py` parses the
   "New Weight" column from each, converts percentages to basis points (out of 10000),
   and normalizes per-chain totals to sum to exactly 10000 by adjusting the largest
   validator's weight.

3. **Identify new validators** — Any validator in the target that doesn't exist on-chain
   yet is written to `new_validators_raw.csv`. This was manually duplicated to
   `new_validators.csv` with cleaned-up names (lowercase, no spaces, no emojis) matching
   the existing on-chain naming convention.

4. **Build the full validator set** — The union of current and target validators is written
   to `all_validators.csv`, with a target weight of 0 for any current validator not in
   the target.

5. **Validation** — The build script verifies that (a) target weights sum to 10000 per
   chain, and (b) every non-zero entry in `all_validators.csv` matches `target_weights.csv`.

6. **Generate Go code** — `scripts/generate_validators_go.py` reads `new_validators.csv`
   and `all_validators.csv` to produce `../validators.go`, which contains the
   `NewValidators` and `TargetWeights` data used by the upgrade handler.

## Files

| File | Description |
|------|-------------|
| `current_weights.csv` | Current on-chain validator weights fetched from the Stride API |
| `target_weights.csv` | Target weights (basis points) for validators with non-zero allocation |
| `all_validators.csv` | Union of current + target validators; 0 weight means the validator should be zeroed out |
| `new_validators_raw.csv` | Validators to add (not yet on-chain) — auto-generated names from the spreadsheet |
| `new_validators.csv` | Same as above with manually cleaned names for the upgrade handler |
| `target/raw/*.csv` | Raw spreadsheet exports with delegation data and new weight assignments |

## Scripts

| Script | Description |
|--------|-------------|
| `scripts/fetch_current_weights.py` | Fetches current validator weights from the Stride host zone API |
| `scripts/build_target_weights.py` | Parses raw CSVs, normalizes weights, builds all output CSVs, and runs validation |
| `scripts/generate_validators_go.py` | Generates `../validators.go` from `new_validators.csv` and `all_validators.csv` |

## Usage

```bash
# Step 1: Fetch current on-chain state
python3 scripts/fetch_current_weights.py

# Step 2: Build target weights and validate
python3 scripts/build_target_weights.py

# Step 3: Generate validators.go for the upgrade handler
python3 scripts/generate_validators_go.py
```
