import { DirectSecp256k1HdWallet, GasPrice, StrideClient, sleep } from "stridejs";
import { stride } from "stridejs";
import { submitTxAndExpectSuccess } from "./txs";
import { getBalance, getHostZoneTotalDelegations } from "./queries";
import { waitForBalanceChange, waitForHostZoneTotalDelegationsChange } from "./polling";
import { ChainConfig, CosmosClient } from "./types";
import {
  STRIDE_CHAIN_NAME,
  USTRD,
  STRIDE_RPC_ENDPOINT,
  DEFAULT_CONNECTION_ID,
  DEFAULT_TRANSFER_CHANNEL_ID,
} from "./consts";
import { ibcTransfer } from "./txs";
import { stringToPath } from "@cosmjs/crypto";
import {
  QueryClient,
  setupAuthExtension,
  setupBankExtension,
  setupIbcExtension,
  setupStakingExtension,
  setupTxExtension,
  SigningStargateClient,
} from "@cosmjs/stargate";
import { Comet38Client } from "@cosmjs/tendermint-rpc";
import { newRegisterHostZoneMsg, newValidator } from "./msgs";

/**
 * Creates a new stride signer account
 * @param mnemonic The account mnemonic
 * @returns The stride client
 */
export async function createStrideClient(mnemonic: string): Promise<StrideClient> {
  const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    prefix: STRIDE_CHAIN_NAME,
  });

  const [{ address }] = await signer.getAccounts();

  return await StrideClient.create(STRIDE_RPC_ENDPOINT, signer, address, {
    gasPrice: GasPrice.fromString(`0.025${USTRD}`),
    broadcastPollIntervalMs: 50,
    resolveIbcResponsesCheckIntervalMs: 50,
  });
}

/**
 * Creates a new host zone signer account
 * @param chainConfig The host chain config
 * @param mnemonic The account mnemonic
 * @returns The host cosmos client
 */
export async function createHostClient(hostConfig: ChainConfig, mnemonic: string): Promise<CosmosClient> {
  const hostSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    prefix: hostConfig.bechPrefix,
    hdPaths: [stringToPath(`m/44'/${hostConfig.coinType}'/0'/0/0`)],
  });
  const [{ address: hostAddress }] = await hostSigner.getAccounts();

  return {
    address: hostAddress,
    denom: hostConfig.hostDenom,
    client: await SigningStargateClient.connectWithSigner(hostConfig.rpcEndpoint, hostSigner, {
      gasPrice: GasPrice.fromString(`1.0${hostConfig.hostDenom}`),
      broadcastPollIntervalMs: 50,
    }),
    query: QueryClient.withExtensions(
      await Comet38Client.connect(hostConfig.rpcEndpoint),
      setupAuthExtension,
      setupBankExtension,
      setupStakingExtension,
      setupIbcExtension,
      setupTxExtension,
    ),
  };
}

/**
 * Checks if a host zone is registered and registers it if it is not
 * @param stridejs The admin stride client
 * @param hostjs A host zone client
 * @param hostConfig The host chain config
 */
export async function ensureHostZoneRegistered({
  stridejs,
  hostjs,
  hostConfig,
}: {
  stridejs: StrideClient;
  hostjs: CosmosClient;
  hostConfig: ChainConfig;
}): Promise<void> {
  const { hostZone } = await stridejs.query.stride.stakeibc.hostZoneAll({});
  const hostZoneRegistered = hostZone.find((hz) => hz.chainId === hostConfig.chainId) !== undefined;

  if (hostZoneRegistered) {
    console.log(`Host zone ${hostConfig.chainName} is already registered...`);
    return;
  }

  console.log(`Registering host zone: ${hostConfig.chainName}...`);
  const registerHostZoneMsg = newRegisterHostZoneMsg({
    sender: stridejs.address,
    connectionId: DEFAULT_CONNECTION_ID,
    transferChannelId: DEFAULT_TRANSFER_CHANNEL_ID,
    hostDenom: hostConfig.hostDenom,
    bechPrefix: hostConfig.bechPrefix,
  });

  const { validators } = await hostjs.query.staking.validators("BOND_STATUS_BONDED");
  const addValidatorsMsg = stride.stakeibc.MessageComposer.withTypeUrl.addValidators({
    creator: stridejs.address,
    hostZone: hostConfig.chainId,
    validators: validators.map((val) =>
      newValidator({
        name: val.description.moniker,
        address: val.operatorAddress,
        weight: 10n,
      }),
    ),
  });

  await submitTxAndExpectSuccess(stridejs, [registerHostZoneMsg, addValidatorsMsg]);
  await sleep(2000);
}

/**
 * Ensures that there are native tokens on stride so that we can liquid stake
 * If there aren't it transfers from the host zone
 * @param stridejs The stride client
 * @param hostjs The host client
 * @param hostChainConfig The host chain config
 * @param minAmount The amount that needs to be liquid staked
 * @returns The balance of native tokens after the transfer
 */
export async function ensureNativeHostTokensOnStride({
  stridejs,
  hostjs,
  hostChainConfig,
  minAmount,
}: {
  stridejs: StrideClient;
  hostjs: CosmosClient;
  hostChainConfig: ChainConfig;
  minAmount: number;
}): Promise<bigint> {
  console.log("Confirming native host tokens are on Stride for liquid staking...");

  const currentBalance = await getBalance({
    client: stridejs,
    denom: hostChainConfig.hostDenomOnStride,
  });

  if (currentBalance >= BigInt(minAmount)) {
    console.log(`Existing balance sufficient, skipping transfer`);
    return currentBalance;
  }

  console.log("Transferring native tokens to Stride...");
  await ibcTransfer({
    client: hostjs,
    sourceChain: hostChainConfig.chainName,
    destinationChain: STRIDE_CHAIN_NAME,
    coin: `${minAmount}${hostChainConfig.hostDenom}`,
    sender: hostjs.address,
    receiver: stridejs.address,
  });

  const finalBalance = await waitForBalanceChange({
    initialBalance: currentBalance,
    client: stridejs,
    address: stridejs.address,
    denom: hostChainConfig.hostDenomOnStride,
  });

  return finalBalance;
}

/**
 * Ensures that there has been a previous liquid stake (identified by there being a delegated balance)
 * on the host zone struct
 * If there isn't, it liquid stakes
 * @param stridejs The stride client
 * @param hostjs The host client
 * @param hostChainConfig The host chain config
 * @param minAmount The amount that needs to be liquid staked
 * @returns The total delegations on the host zone struct after the delegation
 */
export async function ensureLiquidStakeExists({
  stridejs,
  hostjs,
  chainId,
  hostChainConfig,
  minAmount,
}: {
  stridejs: StrideClient;
  hostjs: CosmosClient;
  chainId: string;
  hostChainConfig: ChainConfig;
  minAmount: number;
}): Promise<bigint> {
  console.log("Checking if there are active delegations...");

  // Check if there are existing delegations
  const currentTotalDelegations = await getHostZoneTotalDelegations({
    client: stridejs,
    chainId,
  });

  if (currentTotalDelegations >= BigInt(minAmount)) {
    console.log(`Existing delegations sufficient, skipping setup`);
    return currentTotalDelegations;
  }

  // Ensure there are native tokens that we can liquid stake
  await ensureNativeHostTokensOnStride({
    stridejs,
    hostjs,
    hostChainConfig,
    minAmount,
  });

  // Execute liquid stake
  console.log("Liquid staking...");
  const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
    creator: stridejs.address,
    amount: String(minAmount),
    hostDenom: hostChainConfig.hostDenom,
  });

  await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
  await sleep(2000);

  // Wait for delegation to complete
  console.log("Waiting for delegation to complete...");
  const finalTotalDelegations = await waitForHostZoneTotalDelegationsChange({
    client: stridejs,
    chainId,
    minDelegation: minAmount,
  });

  return finalTotalDelegations;
}
