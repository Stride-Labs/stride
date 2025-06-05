# LSM Scripts

## Setup

- Ensure there are 3 hub validators specified in `config.sh`

```bash
GAIA_NUM_NODES=3
```

## LSM Staking Fllow

- Tokenize shares

```bash
bash dockernet/scripts/lsm/setup.sh
```

- LSM Liquid stake

```bash
bash dockernet/scripts/lsm/stake.sh
```

- View the LSM Deposit record in `logs/state.log` (it should flow through differnet states and then get deleted when the detokenization completes)
