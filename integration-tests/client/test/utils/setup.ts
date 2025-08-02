import { DirectSecp256k1HdWallet, GasPrice, StrideClient, sleep } from "stridejs";
import { stride } from "stridejs";
import { submitTxAndExpectSuccess } from "./txs";
import { getBalance, getHostZoneTotalDelegations } from "./queries";
import { waitForBalanceChange, waitForHostZoneTotalDelegationsChange } from "./polling";
import { ChainConfig, CosmosClient } from "./types";
import { STRIDE_CHAIN_NAME, USTRD, STRIDE_RPC_ENDPOINT, TRANSFER_CHANNEL } from "./consts";
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
import { newRegisterHostZoneMsg, newTransferMsg, newValidator } from "./msgs";
import { Registry } from "@cosmjs/proto-signing";

/**
 * Creates a new stride signer account
 * @param mnemonic The account mnemonic
 * @returns The stride client
 */
export async function createStrideClient(mnemonic: string): Promise<StrideClient> {
  // IMPORTANT: We're using Secp256k1HdWallet from @cosmjs/amino because sending amino txs tests both amino and direct.
  // That's because the tx contains the direct encoding anyway, and also attaches a signature on the amino encoding.
  // The mempool then converts from direct to amino to verify the signature.
  // Therefore if the signature verification passes, we can be sure that both amino and direct are working properly.
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
export async function createHostClient(
  hostConfig: ChainConfig,
  mnemonic: string,
  registry?: Registry,
): Promise<CosmosClient> {
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
      registry,
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
    connectionId: hostConfig.connectionId,
    transferChannelId: hostConfig.transferChannelId,
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
 * Ensures that there are st tokens on the host zone so that we can autopilot redemption
 * If there aren't it liquid stakes and transfers them
 * @param stridejs The stride client
 * @param hostjs The host client
 * @param hostChainConfig The host chain config
 * @param minAmount The amount that needs to be liquid staked
 * @returns The balance of native tokens after the transfer
 */
export async function ensureStTokensOnHost({
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
  console.log("Confirming st tokens are on the host zone for autopilot redemption...");

  const stBalanceOnHost = await getBalance({
    client: hostjs,
    denom: hostChainConfig.stDenomOnHost,
  });

  if (stBalanceOnHost >= BigInt(minAmount)) {
    console.log(`Existing balance sufficient, skipping transfer`);
    return stBalanceOnHost;
  }

  // If there are stTokens on Stride, just send those
  const stBalanceOnStride = await getBalance({
    client: stridejs,
    denom: hostChainConfig.stDenom,
  });
  if (stBalanceOnStride >= BigInt(minAmount)) {
    console.log("Transferring st tokens to host zone...");
    await ibcTransfer({
      client: stridejs,
      sourceChain: STRIDE_CHAIN_NAME,
      destinationChain: hostChainConfig.chainName,
      coin: `${minAmount}${hostChainConfig.stDenom}`,
      sender: stridejs.address,
      receiver: hostjs.address,
    });

    return await waitForBalanceChange({
      client: hostjs,
      address: hostjs.address,
      denom: hostChainConfig.stDenom,
      initialBalance: stBalanceOnStride,
      minChange: minAmount,
    });
  }

  // Finally if there aren't any stTokens yet, do an autopilot liquid stake and forward
  const channelId = TRANSFER_CHANNEL[hostChainConfig.chainName][STRIDE_CHAIN_NAME]!;
  const memo = {
    autopilot: { receiver: stridejs.address, stakeibc: { action: "LiquidStake", ibc_receiver: hostjs.address } },
  };

  const autopilotLiquidStake = newTransferMsg({
    channelId,
    coin: `${minAmount}${hostChainConfig.hostDenom}`,
    sender: hostjs.address,
    receiver: stridejs.address,
    memo: JSON.stringify(memo),
  });
  await submitTxAndExpectSuccess(hostjs, [autopilotLiquidStake]);

  return await waitForBalanceChange({
    initialBalance: stBalanceOnStride,
    client: hostjs,
    address: hostjs.address,
    denom: hostChainConfig.stDenomOnHost,
  });
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
