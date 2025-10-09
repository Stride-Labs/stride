# Mainnet State Exported Testnet [Internal Only]

## Testing Upgrades

### Start a local mainnet node

- Download the latest polkachu snapshot from [here](https://polkachu.com/tendermint_snapshots/stride)
- Checkout the mainnet version of stride, build the binary, and then switch back to the working branch

```bash
# Ex: git checkout v28.0.0
git checkout {old-version}

make install

git checkout {working-branch}
```

- [Mac Only] Setup the home directory - this assumes the snapshot above is in you ~/Downloads folder

```bash
make setup-localstride-node
```

- Start the node

```bash
make start-localstride-node
```

- Kill the code once it's caught up to the latest block
- Export the current state

```bash
# Ex: STAGE=before UPGRADE_NAME=v29 make localstride-state-export
STAGE=before UPGRADE_NAME=v{UPGRADE_NAME} make localstride-state-export
```

- Backup the state files in case the upgrade fails, allowing restarting from a checkpoint

```bash
make backup-localstride
```

### Upgrade the Node

- Testnetify the node and start it up again

```bash
# Ex: UPGRADE_NAME=v29 make upgrade-localstride
UPGRADE_NAME=v{UPGRADE_NAME} make upgrade-localstride
```

- Kill the node again
- Export the new state

```bash
STAGE=after UPGRADE_NAME=v{UPGRADE_NAME} localstride-state-export
```

- View the logs and confirm the upgrade passed
- If it fails and you have to restart, kill it, restore the checkpoint and then re-run the upgrade command

```bash
make restore-localstride-backup
```
