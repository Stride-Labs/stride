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

4. (optional) Add your validator wallet and 9 other preloaded wallets automatically:

```bash
make localnet-keys
```

- These keys are added to your `--keyring-backend test`
- If the keys are already on your keyring, you will get an `"Error: aborted"`
- Ensure you use the name of the account as listed in the table below, as well as ensure you append the `--keyring-backend test` to your txs
- Example: `strided tx bank send lo-test2 stride1kwax6g0q2nwny5n43fswexgefedge033t5g95j --keyring-backend test --chain-id localstride`

5. You can stop the chain, keeping the state with

```bash
make localnet-stop
```

6. When you are done you can clean up the environment with:

```bash
make localnet-clean
```

## LocalStride Accounts

LocalStride is pre-configured with one validator and 10 accounts with stride balances.

| Account   | Address                                                                                                | Mnemonic                                                                                                                                                                   |
| --------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| lo-val    | `stride1wal8dgs7whmykpdaz0chan2f54ynythkz0cazc`<br/>`stridevaloper1wal8dgs7whmykpdaz0chan2f54ynythkp6upwa` | `deer gaze swear marine one perfect hero twice turkey symbol mushroom hub escape accident prevent rifle horse arena secret endless panel equal rely payment`                    |
| lo-test1  | `stride1u9klnra0d4zq9ffalpnr3nhz5859yc7ckdk9wt`                                                          | `journey envelope color ensure fruit assault soup air ozone math beyond miracle very bring bid retire cargo exhaust garden helmet spread sentence insect treat`                       |
| lo-test2  | `stride1kwax6g0q2nwny5n43fswexgefedge033t5g95j`                                                          | `update minimum pyramid initial napkin guilt minute spread diamond dinosaur force observe lounge siren region forest annual citizen mule pilot style horse prize trophy`              |
| lo-test3  | `stride1dv0ecm36ywdyc6zjftw0q62zy6v3mndrwxde03`                                                          | `between flight suffer century action army insane position egg napkin tumble silent enemy crisp club february lake push coral rice few patch hockey ostrich`        |
| lo-test4  | `stride1z3dj2tvqpzy2l5shx98f9k5486tleah5a00fay`                                                          | `muffin brave clinic miss various width depend sand eager mom vicious spoil verb rain leg lunar blossom always silver funny spot frog half coral` |
| lo-test5  | `stride14khzkfs8luaqymdtplrt5uwzrghrndeh4235am`                                                          | `dismiss verb champion ceiling veteran today owner inch field shock dizzy pool creek problem nuclear cage shift romance venue rabbit flower sign bicycle rocket`        |
| lo-test6  | `stride1qym804u6sa2gvxedfy96c0v9jc0ww7593uechw`                                                          | `until lend canvas brain brief blossom tomato tent drip claw more era click bind shrug surprise universe orchard parrot describe jelly scorpion glove path`                  |
| lo-test7  | `stride1et8cdkxl69yrtmpjhxwey52d88kflwzn5xp4xn`                                                          | `choice holiday audit valley asthma empty visa hood lonely primary aerobic that panda define enrich ankle athlete punch glimpse ridge narrow affair thunder lock`                       |
| lo-test8  | `stride1tcrlyn05q9j590uauncywf26ptfn8se65dvfz6`                                                          | `major eager blame canyon jazz occur curious resemble tragic rack tired choose wolf purity meat dog castle attitude decorate moon echo quote core doctor`                 |
| lo-test9  | `stride14ugekxs6f4rfleg6wj8k0wegv69khfpxkt8yn4`                                                          | `neck devote small animal ready swarm melt ugly bronze opinion fire inmate acquire use mobile party paper clock hour view stool aspect angle demand`       |
| lo-test10 | `stride18htv32r83q2wn2knw5wp9m4nkp4xuzyfhmwpqs`                                                          | `almost turtle mobile bullet figure myself dad depart infant vivid view black purity develop kidney cruel seminar outside disorder attack spoil infant sauce blood`     |
