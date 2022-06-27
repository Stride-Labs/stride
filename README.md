# Stride
Bringing interchain liquid staking to Cosmos

## Making changes to this repository
The easiest way to develop cosmos-sdk applications is by using the ignite cli to scaffold code. Ignite (developed by the core cosmos team at Tendermint) makes it possible to scaffold new chains, run relayers, build cosmos related proto files, add messages/queries, add new data structures and more. The drawback of creating thousands of lines of code using ignite is that it is difficult to discern which changes were made by the ignite cli and which changes were made manually by developers. To make it easier to review code written using ignite and to make it easier to retrace our steps if something breaks later, add a commit for each ignite command directly after executing it.

For example, adding a new message type and updating the logic of that message would be two commits.
```
// add the new data type
>>> ignite scaffold list loan amount fee collateral deadline state borrower lender
>>> git add . && git commit -m 'ignite scaffold list loan amount fee collateral deadline state borrower lender'
// make some updates to the keeper method in your code editor
>>> git add . && git commit -m 'update loan list keeper method'
```

An example of a PR using this strategy can be found [here](https://github.com/Stride-Labs/stride/pull/1). Notice, it's easy to differentiate between changes made by ignite and those made manually by reviewing commits. For example, in commit fd3e254bc0, it's easy to see that [a few lines were changes manually](https://github.com/Stride-Labs/stride/pull/1/commits/fd3e254bc0844fe65f5e98f12b366feef2a285f9) even though nearly ~300k LOC were scaffolded.

## Code review format
Opening a PR will automatically create Summary and Test plan fields in the PR description. In the summary, add a high-level summary of what the change entails and the ignite commands run.
**Summary**
Updating some code.
**Test plan**
Add test plan here.


## What is Stride?

**stride** is a blockchain built using Cosmos SDK and Tendermint and created with [Starport](https://starport.com).

## Get started

```
starport chain serve
```

`serve` command installs dependencies, builds, initializes, and starts your blockchain in development.

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Starport docs](https://docs.starport.com).

### Web Frontend

Starport has scaffolded a Vue.js-based web app in the `vue` directory. Run the following commands to install dependencies and start the app:

```
cd vue
npm install
npm run serve
```

The frontend app is built using the `@starport/vue` and `@starport/vuex` packages. For details, see the [monorepo for Starport front-end development](https://github.com/tendermint/vue).

## Release

To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

### Install

To install the latest version of your blockchain node's binary, execute the following command on your machine:

```
curl https://get.starport.com/Stride-Labs/stride@latest! | sudo bash
```

`Stride-Labs/stride` should match the `username` and `repo_name` of the Github repository to which the source code was pushed. Learn more about [the install process](https://github.com/allinbits/starport-installer).

## Learn more

- [Starport](https://starport.com)
- [Tutorials](https://docs.starport.com/guide)
- [Starport docs](https://docs.starport.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.gg/H6wGTY8sxw)

## Build

Please run `make init` to build and serve 3 Stride nodes, 3 Gaia nodes, and 1 Hermes relayer on docker images. 
* `ignite chain build` is run by default each time
* You can optionally pass build arguments to specify which docker images to rebuild
Optional Build Arguments:
1. `s` This will re-build the Stride image (default)
2. `g` This will re-build the Gaia image
3. `h` This will re-build the Hermes image
4. `i` This will re-build the ICQ image

Example:  `make init build=si` will 
1. Run `ignite chain build` to build the Stride binary
2. Rebuild the Stride and ICQ docker images
3. Spin up the 7 docker containers and start all processes

Or, if you just want to re-serve, run `make init build=` to 
1. Use existing Stride binary
2. Use existing docker images 
3. Spin up the 7 docker containers and start all processes

## Local Development
Install the required git submodule dependencies (gaia, hermes, icq)
```
git submodule update --init
make local-install
make local-init build=sghi
```
Install the required packages for each module
```
make local-install

# Or to install to a specific module
make stride-local-install
make gaia-local-install
make icq-local-install
```
Build executables, initialize state, and start the network with `make local-init`
* You can optionally pass build arguments to specify which binary to rebuild
1. `s` This will re-build the Stride binary (default)
2. `g` This will re-build the Gaia binary
3. `h` This will re-build the Hermes binary
4. `i` This will re-build the ICQ binary
* You can optionally pass `cache=true` to restore the state from a backup instead of re-intializing it. 
* Example: `make local-init build=si`, this will:
    * Rebuild the Stride and Gaia binarys
    * Start 1 Stride and 1 Gaia node in the background
    * Start Hermes and ICQ Relayers
Run `make stop` to bring down the chain
## Testing

To run the full test suite, run `make init`, then `sh scripts/tests/run_all_tests.sh`.

If bats testing subdirectories are not populated, run `git submodule update --init`.
