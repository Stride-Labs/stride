import { StrideClient, sleep } from "stridejs";
import { stride } from "stridejs";
import { submitTxAndExpectSuccess } from "./txs";
import { getBalance, getHostZoneTotalDelegations } from "./queries";
import { waitForBalanceChange, waitForHostZoneTotalDelegationsChange } from "./polling";
import { ChainConfig, CosmosClient } from "./types";
import { STRIDE_CHAIN_NAME, HOST_CHAIN_NAME } from "./consts";
import { ibcTransfer } from "./txs";

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
    console.log(`Existing balance (${currentBalance}) sufficient, skipping transfer`);
    return currentBalance;
  }

  console.log("Transferring native tokens to Stride...");
  await ibcTransfer({
    client: hostjs,
    sourceChain: HOST_CHAIN_NAME,
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
    console.log(`Existing delegations (${currentTotalDelegations}) sufficient, skipping setup`);
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
