# LocalStride

Inspired by LocalOsmosis, LocalStride is a complete Stride testnet containerized with Docker and orchestrated with a simple docker-compose file. LocalStride comes pre-configured with opinionated, sensible defaults for a standard testing environment.

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

A few things to note before getting started:
 * The local version of stride must match the version of stride on mainnet. I.e. if mainnet is on v11.0.0 and you try following these instructions for v12.0.0 or main, it will fail when initializing the genesis. What can be done instead is starting with v11.0.0 and running the upgrade from v11 to v12 to ensure it goes smoothly.  
 * Running localstride with mainnet state is very memory intensive. It is recommended to have 128GB of memory available, possibly through a VM. If partway through the instructions, the docker service exits with code 137, this means it ran out of memory.
 * It is recommended that you wait for a snapshot from the last twelve hours to be present at `https://polkachu.com/tendermint_snapshots/stride` before proceeding.  
 * The commands should be run from the root of stride repository unless stated otherwise.  

### Create a mainnet state export

0. If using a VM, be sure to SSH in directly through a local terminal, using a command like:  
```sh
gcloud compute ssh --zone "us-central1-a" "biggie-smalls" --project "stride-nodes"
```
Also, replace `/home/kentgang/` in the two volumes sections with your home directory in `stride/testutil/localstride/state-export/docker-compose.yml`

1. Set up a node on mainnet. The following command will download a wizard to guide you through the process.  

```sh
bash -c "$(curl -sSL node.stride.zone/install)"
```

2. Check that the local node is caught up to the mainnet bloc kheight, before killing the Stride daemon. The current block height can be found at `https://www.mintscan.io/stride/blocks`.  

3. Export the mainnet state with the following command:

```sh
strided export > export_state.json
```

### Bootstrap LocalStride with mainnet state

4. Copy or move the mainnet state to the `localstride/state-export` folder.  

```sh
cp export_state.json testutil/localstride/state-export/
```

5. If running from scratch, ignore this step. If a past run was executed, delete the home directories for the `stride` and `stride2` services. These may be located at `~/.stride` and `~/.stride2` respectively.  

6. Build the `local:stride` docker image. 

```sh
make localnet-state-export-build
```

7. Start LocalStride, and kill the process once it's completed preparing the genesis files.

```sh
make localnet-state-export-start
```

This command will begin by consuming the `export_state.json` file from steps 1-4, make modifications to allocate nearly all power to the local private validator key,  and output several files.  

It will take a few moments to open and write the json file. The process should be killed after the json is written and before/while tendermint is setting up.

Once you see something akin to the following effect, the process can be killed:  

```sh
state-export-stride-1   | 	Update total ustrd supply from 86896976915130 to 2086896976915130
state-export-stride-1   | Set governors as validators
state-export-stride-1   | 🥸  Replace Provider Fee Pool Addr
state-export-stride-1   | 📝 Writing /root/.stride/config/genesis.json... (it may take a while)
state-export-stride2-1  | /root/.stride
state-export-stride2-1  | 836fd688c92f31dc84dfc138cd006c6a5083abee
state-export-stride-1   | /root/.stride
state-export-stride-1   | c59c5cf7730a2ebc3a6b9259f91d3e795a90d521
state-export-stride2-1  | 4:16PM INF starting node with ABCI Tendermint in-process module=server
state-export-stride-1   | 4:16PM INF starting node with ABCI Tendermint in-process module=server
```

Note, there are two services being initialized, the primary node `stride` and a kicker node companion `stride2`. The hashes `836fd688c92f31dc84dfc138cd006c6a5083abee` and `c59c5cf7730a2ebc3a6b9259f91d3e795a90d521` in the example above are the node ids that must be added as persistent peers in the configuration files for the two nodes.  

8. Modify the two driver scripts to modify the configurations with the correct node id of the opposite node.  

In `stride/testutil/localstride/state-export/scripts/start1.sh`, modify lines 44-45 by replacing `{NODE_ID_TWO}` with the node id of the service `stride2`. For the example above:  

```sh
    dasel put string -f $CONFIG_FOLDER/config.toml '.persistent_peers' "836fd688c92f31dc84dfc138cd006c6a5083abee@stride2:26658"
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.persistent_peers' "836fd688c92f31dc84dfc138cd006c6a5083abee@stride2:26658"
}
```

Repeat the previous instruction for `start2.sh` resulting in:
```sh
    dasel put string -f $CONFIG_FOLDER/config.toml '.persistent_peers' "c59c5cf7730a2ebc3a6b9259f91d3e795a90d521@stride:26656"
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.persistent_peers' "c59c5cf7730a2ebc3a6b9259f91d3e795a90d521@stride:26656"
}
```

Note, the node ID for `stride` is used in `start2.sh` and the node ID for `stride2` is used in `start1.sh`.  

9. Open a second terminal. In both terminals navigate to `stride/testutil/localstride/state-export`  

Around the same time, run these two commands in their own terminals: 

```sh
docker-compose up stride
```

```sh
docker-compose up stride2
```

10. Wait for it to chug, and start producing blocks (eta: 5 min).  

Some key checkpoints and expected behavior are:  

* The node is nominally initialized, and its address is declared  

```sh
state-export-stride-1  | 4:22PM INF This node is a validator addr=8A1C116786FE88D48E1B7092A3C76727BD085179 module=consensus pubKey=krLZ4b5DKXjeGarvm3s7kSZ6HXsJ9WZmf3iQqTKGOeU=
```

* The persistent peer is recognized and added to the address book  

```sh
state-export-stride-1  | 4:22PM INF Adding persistent peers addrs=["836fd688c92f31dc84dfc138cd006c6a5083abee@stride2:26658"] module=p2p
state-export-stride-1  | 4:22PM INF Adding unconditional peer ids ids=[] module=p2p
state-export-stride-1  | 4:22PM INF Add our address to book addr={"id":"c59c5cf7730a2ebc3a6b9259f91d3e795a90d521","ip":"0.0.0.0","port":26656} book=/root/.stride/config/addrbook.json module=p2p
```

* A proposal is processed, and the initial block the mainnet state export is on is completed 

```sh
state-export-stride-1  | 4:22PM INF received proposal module=consensus proposal={"Type":32,"block_id":{"hash":"3311D9A96272B4125DC895A0E437DA68830E119D28986291F68864FE9EB419EF","parts":{"hash":"9723CBAA04AB11FF9758BE8D0053F42A65E75DC556D8B7F45D956F3AC3A04815","total":1}},"height":5037413,"pol_round":-1,"round":0,"signature":"NvjaOCu5Klsd0fuVBzq6Kez3/KQVx0psuJEf/Kr/TEz2n+OWDLd3vCcFmaTU5/EWUeM4/vXeq50oTqr18R4RDg==","timestamp":"2023-08-18T16:22:30.865176394Z"}
state-export-stride-1  | 4:22PM INF received complete proposal block hash=3311D9A96272B4125DC895A0E437DA68830E119D28986291F68864FE9EB419EF height=5037413 module=consensus
state-export-stride-1  | 4:22PM INF finalizing commit of block hash={} height=5037413 module=consensus num_txs=0 root=E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855
```

Note, this does *not* indicate that the blockchain is fully operational.

* Some errors may be raised. The nodes will connect to its peer, while also complaining it needs other peers. The node is believed to be consuming the genesis state in the background.

```sh
state-export-stride-1  | 4:24PM ERR Connection failed @ sendRoutine conn={"Logger":{"Logger":{}}} err="pong timeout" module=p2p peer={"id":"836fd688c92f31dc84dfc138cd006c6a5083abee","ip":"172.18.0.3","port":60706}
state-export-stride-1  | 4:24PM INF service stop impl={"Logger":{"Logger":{}}} module=p2p msg={} peer={"id":"836fd688c92f31dc84dfc138cd006c6a5083abee","ip":"172.18.0.3","port":60706}
state-export-stride-1  | 4:24PM ERR Stopping peer for error err="pong timeout" module=p2p peer={"Data":{},"Logger":{"Logger":{}}}
state-export-stride-1  | 4:24PM INF service stop impl={"Data":{},"Logger":{"Logger":{}}} module=p2p msg={} peer={"id":"836fd688c92f31dc84dfc138cd006c6a5083abee","ip":"172.18.0.3","port":60706}
state-export-stride-1  | 4:24PM INF service start impl="Peer{MConn{172.18.0.3:51340} 836fd688c92f31dc84dfc138cd006c6a5083abee in}" module=p2p msg={} peer={"id":"836fd688c92f31dc84dfc138cd006c6a5083abee","ip":"172.18.0.3","port":51340}
state-export-stride-1  | 4:24PM INF service start impl=MConn{172.18.0.3:51340} module=p2p msg={} peer={"id":"836fd688c92f31dc84dfc138cd006c6a5083abee","ip":"172.18.0.3","port":51340}
state-export-stride-1  | 4:24PM INF Saving AddrBook to file book=/root/.stride/config/addrbook.json module=p2p size=0
```

* Eventually, more proposals will be received and the reported block height will increment:

```sh
state-export-stride-1  | 4:26PM INF executed block height=5037415 module=state num_invalid_txs=0 num_valid_txs=0
state-export-stride-1  | 4:26PM INF commit synced commit=436F6D6D697449447B5B34312034392032343920313620333520313236203232352038362035322038332035392036382032343020323720313930203135342038352031303220313535203131352031333720313230203136342031373920353720313934203633203132203130302032313020323330203136375D3A3443444436377D module=server
state-export-stride-1  | 4:26PM INF committed state app_hash=2931F910237EE15634533B44F01BBE9A55669B738978A4B339C23F0C64D2E6A7 height=5037415 module=state num_txs=0
state-export-stride-1  | 4:26PM INF indexed block exents height=5037415 module=txindex
state-export-stride-1  | 4:26PM INF Ensure peers module=pex numDialing=0 numInPeers=1 numOutPeers=0 numToDial=10
state-export-stride-1  | 4:26PM INF We need more addresses. Sending pexRequest to random peer module=pex peer={"Data":{},"Logger":{"Logger":{}}}
state-export-stride-1  | 4:26PM INF No addresses to dial. Falling back to seeds module=pex
state-export-stride-1  | 4:26PM INF Timed out dur=4916.365002 height=5037416 module=consensus round=0 step=1
```

In the example above, we see the height has gone from 415 -> 416. The blockchain is fully operational at this time.

10. Exit out of the kicker node.  

11. You can now query the status of LocalStride:

```sh
strided status
```

Additionally, you can send tokens:  

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

### Testing the upgrade
* Once localstride starts churning blocks, you are ready to test the upgrade. Run the following to submit and vote on the upgrade:
```bash 
# Check the localstride logs to determine the current block and propose the upgrade at a height at least 75 blocks in the future
#  Ex: make localnet-state-export-upgrade upgrade_name=v5 upgrade_height=1956500
make localnet-state-export-upgrade upgrade_name={upgrade_name} upgrade_height={upgrade_height}
```
* Wait for the upgrade height and confirm the node crashed. Run the following to take down the container:
```
make localnet-state-export-stop
```
* Switch the repo back to the version we're upgrading to and re-build the stride image **without clearing the state**:
```bash
git checkout {latest_branch}
make localnet-state-export-build
```
* Finally, start the node back up with the updated binary
```bash
make localnet-state-export-start
```
* Check the localstride logs and confirm the upgrade succeeded
* If the upgrade passes, but then the chain hangs without producing blocks, you may need to add a second "kicker" node to jump start the network. To do this:
  * Copy the `~/.stride/data` directory over to another machine
  * Get the node ID of each machine by running `strided tendermint show-node-id`
  * Add each node as a persistent peer of the other by adding `persistent_peers = "{NODE_ID}@{IP}:26656"` to `~/.stride/config/config.toml`
  * Start both nodes until they start committing blocks
  * Once they're rolling, you can shut the second node off

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
