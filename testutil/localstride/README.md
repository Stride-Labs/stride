# LocalStride

Inspired by LocalOsmosis, LocalStride is a complete Stride testnet containerized with Docker and orchestrated with a simple docker-compose file. LocalStride comes pre-configured with opinionated, sensible defaults for a standard testing environment.

## Prerequisites

Ensure you have docker and docker-compose installed:

```sh
# Docker
sudo apt-get remove docker docker-engine docker.io
sudo apt-get update
sudo apt install docker.io -y

# Docker compose
sudo apt install docker-compose -y
```

## 1. LocalStride - No Initial State

The following commands must be executed from the root folder of the Stride repository.

1. Make any change to the stride code that you want to test

2. Initialize LocalStride:

```bash
make localnet-init
```

The command:

- Builds a local docker image with the latest changes
- Cleans the `$HOME/.stride` folder

3. Start LocalStride:

```bash
make localnet-start
```

> Note
>
> You can also start LocalStride in detach mode with:
>
> `make localnet-startd`

4. (optional) Add your validator wallet and 10 other preloaded wallets automatically:

```bash
make localnet-keys
```

- These keys are added to your `--keyring-backend test`
- If the keys are already on your keyring, you will get an `"Error: aborted"`
- Ensure you use the name of the account as listed in the table below, as well as ensure you append the `--keyring-backend test` to your txs
- Example: `strided tx bank send ls-test2 stride1kwax6g0q2nwny5n43fswexgefedge033t5g95j --keyring-backend test --chain-id localstride`

5. You can stop the chain, keeping the state with

```bash
make localnet-stop
```

6. When you are done you can clean up the environment with:

```bash
make localnet-clean
```

## 2. LocalStride - With Mainnet State

A few things to note before getting started. The below method will only work if you are using the same version as mainnet. In other words,
if mainnet is on v8.0.0 and you try to do this on a v9.0.0 tag or on main, you will run into an error when initializing the genesis. What you can do though is run localstride on the mainnet version, then go through the upgrade process to ensure the upgrade with mainnet state goes smoothly.

### Create a mainnet state export

1. Set up a node on mainnet

2. Ensure your node is caught up to the head of the network, or whatever block you want to start your testnet from

3. Stop your Stride daemon

4. Take a state export snapshot with the following command:

```sh
strided export > state_export.json
```

This will create a file called `state_export.json` which is a snapshot of the current mainnet state.

### Use the state export in LocalStride

1. Copy the `state_export.json` to the `localstride/state_export` folder within the stride repo

```sh
cp $HOME/state_export.json $HOME/stride/testutil/localstride/state_export/
```

6. Ensure you have docker and docker-compose installed:

```sh
# Docker
sudo apt-get remove docker docker-engine docker.io
sudo apt-get update
sudo apt install docker.io -y

# Docker compose
sudo apt install docker-compose -y
```

7. Build the `local:stride` docker image (select yes if prompted to recursively remove):

```bash
make localnet-state-export-init
```

The command:

- Builds a local docker image with the latest changes
- Cleans the `$HOME/.stride` folder

3. Start LocalStride:

```bash
make localnet-state-export-start
```

> Note
>
> You can also start LocalStride in detach mode with:
>
> `make localnet-state-export-startd`

When running this command for the first time, `local:stride` will:

- Modify the provided `state_export.json` to create a new state suitable for a testnet
- Start the chain

You will then go through the genesis initialization process and hit the first block (not block 1, but the block number after your snapshot was taken)

During this process, you may see only p2p logs and no blocks. **This could be the case for the next 30 minutes**, but will eventually start hitting blocks.

9. On your host machine, add this specific wallet which holds a large amount of stride funds

```sh
MNEMONIC="deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment"
echo $MNEMONIC | strided keys add val --recover --keyring-backend test
```

You now are running a validator with a majority of the voting power with the same state as mainnet state (at the time you took the snapshot)

10. On your host machine, you can now query the state-exported testnet:

```sh
strided status
```

11. Here is an example command to ensure complete understanding:

```sh
strided tx bank send val stride1qym804u6sa2gvxedfy96c0v9jc0ww7593uechw 10000000ustrd --chain-id localstride --keyring-backend test
```

12. You can stop chain, keeping the state with

```bash
make localnet-state-export-stop
```

13. When you are done you can clean up the environment with:

```bash
make localnet-state-export-clean
```

Note: At some point, all the validators (except yours) will get jailed at the same block due to them being offline.

When this happens, it may take a little bit of time to process. Once all validators are jailed, you will continue to hit blocks as you did before.
If you are only running the validator for a short time (< 24 hours) you will not experience this.

## LocalStride Accounts

LocalStride is pre-configured with one validator and 10 accounts with stride balances.

| Account   | Address                                                                                                | Mnemonic                                                                                                                                                                   |
| --------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| val    | `stride1wal8dgs7whmykpdaz0chan2f54ynythkz0cazc`<br/>`stridevaloper1wal8dgs7whmykpdaz0chan2f54ynythkp6upwa` | `deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment`                    |
| ls-test1  | `stride1u9klnra0d4zq9ffalpnr3nhz5859yc7ckdk9wt`                                                          | `journey envelope color ensure fruit assault soup air ozone math beyond miracle very bring bid retire cargo exhaust garden helmet spread sentence insect treat`                       |
| ls-test2  | `stride1kwax6g0q2nwny5n43fswexgefedge033t5g95j`                                                          | `update minimum pyramid initial napkin guilt minute spread diamond dinosaur force observe lounge siren region forest annual citizen mule pilot style horse prize trophy`              |
| ls-test3  | `stride1dv0ecm36ywdyc6zjftw0q62zy6v3mndrwxde03`                                                          | `between flight suffer century action army insane position egg napkin tumble silent enemy crisp club february lake push coral rice few patch hockey ostrich`        |
| ls-test4  | `stride1z3dj2tvqpzy2l5shx98f9k5486tleah5a00fay`                                                          | `muffin brave clinic miss various width depend sand eager mom vicious spoil verb rain leg lunar blossom always silver funny spot frog half coral` |
| ls-test5  | `stride14khzkfs8luaqymdtplrt5uwzrghrndeh4235am`                                                          | `dismiss verb champion ceiling veteran today owner inch field shock dizzy pool creek problem nuclear cage shift romance venue rabbit flower sign bicycle rocket`        |
| ls-test6  | `stride1qym804u6sa2gvxedfy96c0v9jc0ww7593uechw`                                                          | `until lend canvas brain brief blossom tomato tent drip claw more era click bind shrug surprise universe orchard parrot describe jelly scorpion glove path`                  |
| ls-test7  | `stride1et8cdkxl69yrtmpjhxwey52d88kflwzn5xp4xn`                                                          | `choice holiday audit valley asthma empty visa hood lonely primary aerobic that panda define enrich ankle athlete punch glimpse ridge narrow affair thunder lock`                       |
| ls-test8  | `stride1tcrlyn05q9j590uauncywf26ptfn8se65dvfz6`                                                          | `major eager blame canyon jazz occur curious resemble tragic rack tired choose wolf purity meat dog castle attitude decorate moon echo quote core doctor`                 |
| ls-test9  | `stride14ugekxs6f4rfleg6wj8k0wegv69khfpxkt8yn4`                                                          | `neck devote small animal ready swarm melt ugly bronze opinion fire inmate acquire use mobile party paper clock hour view stool aspect angle demand`       |
| ls-test10 | `stride18htv32r83q2wn2knw5wp9m4nkp4xuzyfhmwpqs`                                                          | `almost turtle mobile bullet figure myself dad depart infant vivid view black purity develop kidney cruel seminar outside disorder attack spoil infant sauce blood`     |
