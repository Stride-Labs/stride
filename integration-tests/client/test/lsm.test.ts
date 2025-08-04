import {
  cosmosProtoRegistry,
  gaiaProtoRegistry,
  ibcDenom,
  ibcProtoRegistry,
  sleep,
  strideProtoRegistry,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  USTRD,
  CHAIN_CONFIGS,
  STRIDE_CHAIN_NAME,
  MNEMONICS,
  TRANSFER_CHANNEL,
  COSMOSHUB_CHAIN_NAME,
} from "./utils/consts";
import { CosmosClient } from "./utils/types";
import { ibcTransfer, submitTxAndExpectSuccess } from "./utils/txs";
import { waitForChain, assertICAChannelsOpen, assertOpenTransferChannel } from "./utils/startup";
import { waitForBalanceChange, waitForHostZoneTotalDelegationsChange } from "./utils/polling";
import { getBalance, getHostZone, getLatestTokenizeShareRecordId } from "./utils/queries";
import { StrideClient } from "stridejs";
import { createHostClient, createStrideClient, ensureHostZoneRegistered } from "./utils/setup";
import { newDelegateMsg, newLSMLiquidStakeMsg, newTokenizeSharesMsg } from "./utils/msgs";
import { Registry } from "@cosmjs/proto-signing";

describe("LSM", () => {
  const HOST_CONFIG = CHAIN_CONFIGS[COSMOSHUB_CHAIN_NAME];

  let strideAccounts: {
    user: StrideClient;
    admin: StrideClient;
  };

  let hostAccounts: {
    user: CosmosClient;
  };

  // Initialize accounts and wait for the chain to start
  beforeAll(async () => {
    // @ts-expect-error
    strideAccounts = {};
    // @ts-expect-error
    hostAccounts = {};

    const admin = MNEMONICS.admin;
    const user = MNEMONICS.users[0];

    const gaiaRegistry = new Registry([...gaiaProtoRegistry, ...cosmosProtoRegistry, ...ibcProtoRegistry]);

    strideAccounts["admin"] = await createStrideClient(admin.mnemonic);
    strideAccounts["user"] = await createStrideClient(user.mnemonic);
    hostAccounts["user"] = await createHostClient(HOST_CONFIG, user.mnemonic, gaiaRegistry);

    await waitForChain(STRIDE_CHAIN_NAME, strideAccounts.user, USTRD);
    await waitForChain(HOST_CONFIG.chainName, hostAccounts.user, HOST_CONFIG.hostDenom);

    const strideToHostChannel = TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][HOST_CONFIG.chainName];
    const hostToStrideChannel = TRANSFER_CHANNEL[HOST_CONFIG.chainName][STRIDE_CHAIN_NAME];
    await assertOpenTransferChannel(STRIDE_CHAIN_NAME, strideAccounts.user, strideToHostChannel);
    await assertOpenTransferChannel(HOST_CONFIG.chainName, hostAccounts.user, hostToStrideChannel);

    await ensureHostZoneRegistered({
      stridejs: strideAccounts.admin,
      hostjs: hostAccounts.user,
      hostConfig: HOST_CONFIG,
    });

    await assertICAChannelsOpen(strideAccounts.admin, HOST_CONFIG.chainId);
  }, 45_000);

  test("LSM Liquid Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;

    const { hostDenom, stDenom, chainId } = HOST_CONFIG;

    // Delegate normal stake on the host chain
    const { validators } = await getHostZone({ client: stridejs, chainId: chainId });
    const validator = validators[0].address;

    console.log("Delegating native stake...");
    const msgDelegate = newDelegateMsg({
      validator: validator,
      delegator: hostjs.address,
      amount: stakeAmount,
      denom: hostDenom,
    });
    await submitTxAndExpectSuccess(hostjs, [msgDelegate]);
    await sleep(2000);

    // Tokenize those shares
    console.log("Tokenizing shares...");
    const msgTokenize = newTokenizeSharesMsg({
      validator,
      delegator: hostjs.address,
      amount: stakeAmount,
      denom: hostDenom,
    });
    await submitTxAndExpectSuccess(hostjs, [msgTokenize]);
    await sleep(2000);

    // Get the record ID and denoms
    const tokenizeShareRecordId = await getLatestTokenizeShareRecordId();
    const lsmDenomOnHost = `${validator}/${tokenizeShareRecordId}`;

    const channelId = TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][COSMOSHUB_CHAIN_NAME]!;
    const lsmDenomOnStride = ibcDenom([{ incomingPortId: "transfer", incomingChannelId: channelId }], lsmDenomOnHost);

    // Get initial balances
    let initialLsmBalanceOnStride = await getBalance({ client: stridejs, denom: lsmDenomOnStride });
    const initialStBalanceOnStride = await getBalance({ client: stridejs, denom: stDenom });

    // Transfer them to stride
    console.log("Transferring tokenized shares...");
    await ibcTransfer({
      client: hostjs,
      sourceChain: COSMOSHUB_CHAIN_NAME,
      destinationChain: STRIDE_CHAIN_NAME,
      sender: hostjs.address,
      receiver: stridejs.address,
      coin: {
        amount: stakeAmount.toString(),
        denom: lsmDenomOnHost,
      },
    });

    // Wait to receive them on Stride and get initial balances
    console.log("Waiting for transfer to complete...");
    initialLsmBalanceOnStride = await waitForBalanceChange({
      client: stridejs,
      address: stridejs.address,
      denom: lsmDenomOnStride,
      minChange: stakeAmount,
      initialBalance: initialLsmBalanceOnStride,
    });

    console.log("Submitting LSM liquid stakes...");
    const lsmLiquidStakeMsg = newLSMLiquidStakeMsg({
      staker: stridejs.address,
      amount: stakeAmount / 2, // half total amount
      lsmTokenIbcDenom: lsmDenomOnStride,
    });

    await submitTxAndExpectSuccess(hostjs, [lsmLiquidStakeMsg]);
    await sleep(2000);

    // Liquid stake the other half (this one should trigger a slash query)
    await submitTxAndExpectSuccess(hostjs, [lsmLiquidStakeMsg]);
    await sleep(2000);

    // Get final st and tokenize share record balances
    const finalStBalanceOnStride = await getBalance({ client: stridejs, denom: stDenom });
    const finalLsmBalanceOnStride = await getBalance({ client: stridejs, denom: lsmDenomOnStride });

    // Confirm the balance changes
    // LSM balance should decrease (sent for staking)
    // StBalance should increase (minted)
    const lsmBalanceDiff = finalLsmBalanceOnStride - initialLsmBalanceOnStride;
    const stBalanceDiff = finalStBalanceOnStride - initialStBalanceOnStride;
    expect(lsmBalanceDiff).to.equal(-BigInt(stakeAmount), "LSM token balance");
    expect(stBalanceDiff >= 0).to.be.true;
    expect(stBalanceDiff <= stakeAmount).to.be.true; // less than or equal to because of redemption rate

    // Wait for the LSM shares to get converted to native stake
    await waitForHostZoneTotalDelegationsChange({
      client: stridejs,
      chainId: COSMOSHUB_CHAIN_NAME,
      minDelegation: stakeAmount,
    });
  }, 180_000); // 3 min timeout
});
