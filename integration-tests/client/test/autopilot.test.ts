import { sleep } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  USTRD,
  CHAIN_CONFIGS,
  STRIDE_CHAIN_NAME,
  MNEMONICS,
  TEST_CHAINS,
  TRANSFER_CHANNEL,
  DEFAULT_FEE,
} from "./utils/consts";
import { CosmosClient } from "./utils/types";
import { submitTxAndExpectSuccess } from "./utils/txs";
import { waitForChain, assertICAChannelsOpen, assertOpenTransferChannel } from "./utils/startup";
import { waitForBalanceChange } from "./utils/polling";
import { getBalance, getLatestUserRedemptionRecord } from "./utils/queries";
import { StrideClient } from "stridejs";
import {
  createHostClient,
  createStrideClient,
  ensureHostZoneRegistered,
  ensureLiquidStakeExists,
  ensureStTokensOnHost,
} from "./utils/setup";
import { newTransferMsg } from "./utils/msgs";

describe.each(TEST_CHAINS)("Autopilot - %s", (hostChainName) => {
  const HOST_CONFIG = CHAIN_CONFIGS[hostChainName];

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

    strideAccounts["admin"] = await createStrideClient(admin.mnemonic);
    strideAccounts["user"] = await createStrideClient(user.mnemonic);
    hostAccounts["user"] = await createHostClient(HOST_CONFIG, user.mnemonic);

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

  test("Liquid Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;

    const { chainName, hostDenom, stDenom } = HOST_CONFIG;

    // Get the initial native token balance on the host zone and the st token balance on stride
    let initialUserNativeBalance = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    const initialUserStBalance = await getBalance({
      client: stridejs,
      denom: stDenom,
    });

    // Perform autopilot liquid staking
    console.log("Autopilot liquid staking...");
    const channelId = TRANSFER_CHANNEL[chainName][STRIDE_CHAIN_NAME]!;

    const memo = { autopilot: { receiver: stridejs.address, stakeibc: { action: "LiquidStake" } } };
    const autopilotLiquidStake = newTransferMsg({
      channelId,
      coin: `${stakeAmount}${hostDenom}`,
      sender: hostjs.address,
      receiver: stridejs.address,
      memo: JSON.stringify(memo),
    });

    await submitTxAndExpectSuccess(hostjs, [autopilotLiquidStake]);

    // Get final native and st balances
    const finalUserStBalance = await waitForBalanceChange({
      client: stridejs,
      address: stridejs.address,
      denom: stDenom,
      initialBalance: initialUserStBalance,
    });
    const finalUserNativeBalance = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    // Confirm the balance changes
    // Native balance should decrease (sent for staking)
    // StBalance should increase (minted)
    const nativeBalanceDiff = finalUserNativeBalance - initialUserNativeBalance;
    const stBalanceDiff = finalUserStBalance - initialUserStBalance;
    expect(nativeBalanceDiff).to.equal(-(BigInt(stakeAmount) + DEFAULT_FEE), "User native balance change on Stride");
    expect(stBalanceDiff >= 0).to.be.true;
    expect(stBalanceDiff <= stakeAmount).to.be.true; // less than or equal to because of redemption rate
  }, 180_000); // 3 min timeout

  test("Liquid Stake and Forward", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;

    const { chainName, hostDenom, stDenomOnHost } = HOST_CONFIG;

    // Get initial native token balance and st token balance on the host zone
    let initialUserNativeBalance = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    const initialUserStBalance = await getBalance({
      client: hostjs,
      denom: stDenomOnHost,
    });

    // Perform autopilot liquid staking
    console.log("Autopilot liquid staking with forwarding...");

    const channelId = TRANSFER_CHANNEL[chainName][STRIDE_CHAIN_NAME]!;
    const memo = {
      autopilot: { receiver: stridejs.address, stakeibc: { action: "LiquidStake", ibc_receiver: hostjs.address } },
    };

    const autopilotLiquidStake = newTransferMsg({
      channelId,
      coin: `${stakeAmount}${hostDenom}`,
      sender: hostjs.address,
      receiver: stridejs.address,
      memo: JSON.stringify(memo),
    });

    await submitTxAndExpectSuccess(hostjs, [autopilotLiquidStake]);

    // Get final native and st balances
    // We check the stToken balance on the host zone (since it should get transferred back)
    const finalUserStBalance = await waitForBalanceChange({
      client: hostjs,
      address: hostjs.address,
      denom: stDenomOnHost,
      initialBalance: initialUserStBalance,
    });
    const finalUserNativeBalance = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    // Confirm the balance changes
    // Native balance should decrease (sent for staking)
    // StBalance should increase (minted)
    const nativeBalanceDiff = finalUserNativeBalance - initialUserNativeBalance;
    const stBalanceDiff = finalUserStBalance - initialUserStBalance;
    expect(nativeBalanceDiff).to.equal(-(BigInt(stakeAmount) + DEFAULT_FEE), "User native balance change on Stride");
    expect(stBalanceDiff >= 0).to.be.true;
    expect(stBalanceDiff <= stakeAmount).to.be.true; // less than or equal to because of redemption rate
  }, 180_000); // 3 min timeout

  test("Redeem Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;
    const redeemAmount = 1000000;

    const { chainName, chainId, stDenomOnHost } = HOST_CONFIG;

    // Ensure there's enough liquid stake to cover the redemption
    await ensureLiquidStakeExists({
      stridejs,
      hostjs,
      chainId,
      hostChainConfig: HOST_CONFIG,
      minAmount: stakeAmount,
    });

    // Ensure we have stTokens on the host zone
    const initialStBalanceOnHost = await ensureStTokensOnHost({
      stridejs,
      hostjs,
      hostChainConfig: HOST_CONFIG,
      minAmount: redeemAmount,
    });

    // Autopilot redeem stake
    console.log("Autopilot redeeming stake..");

    const channelId = TRANSFER_CHANNEL[chainName][STRIDE_CHAIN_NAME]!;
    const memo = {
      autopilot: { receiver: stridejs.address, stakeibc: { action: "RedeemStake", ibc_receiver: hostjs.address } },
    };

    const autopilotRedeemStake = newTransferMsg({
      channelId,
      coin: `${redeemAmount}${stDenomOnHost}`,
      sender: hostjs.address,
      receiver: stridejs.address,
      memo: JSON.stringify(memo),
    });

    await submitTxAndExpectSuccess(hostjs, [autopilotRedeemStake]);
    await sleep(5000); // wait for transfer to complete

    // Confirm a redemption record was created
    const redemptionRecord = await getLatestUserRedemptionRecord({
      client: stridejs,
      chainId,
      receiver: hostjs.address,
    });
    expect(BigInt(redemptionRecord.stTokenAmount) >= BigInt(redeemAmount)).to.be.true;

    // Confirm the st token balance decreased
    const finalStBalanceOnHost = await getBalance({ client: hostjs, denom: stDenomOnHost });
    const stBalanceDiff = initialStBalanceOnHost - finalStBalanceOnHost;
    expect(stBalanceDiff).to.equal(BigInt(redeemAmount), "st token balance difference");
  }, 120_000); // 2 minute timeout

  test("Redeem Stake Failure", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;
    const redeemAmount = 1000000;

    const { chainName, chainId, hostDenom, stDenomOnHost } = HOST_CONFIG;

    // Ensure there's enough liquid stake to cover the redemption
    await ensureLiquidStakeExists({
      stridejs,
      hostjs,
      chainId,
      hostChainConfig: HOST_CONFIG,
      minAmount: stakeAmount,
    });

    // Ensure we have stTokens on the host zone
    const initialStBalanceOnHost = await ensureStTokensOnHost({
      stridejs,
      hostjs,
      hostChainConfig: HOST_CONFIG,
      minAmount: redeemAmount,
    });

    // Autopilot redeem stake but with a memo that should fail
    console.log("Autopilot redeeming stake failure case..");

    const channelId = TRANSFER_CHANNEL[chainName][STRIDE_CHAIN_NAME]!;
    const memo = {
      autopilot: { receiver: stridejs.address, stakeibc: { action: "RedeemStake", ibc_receiver: "INVALID" } },
    };

    const autopilotRedeemStake = newTransferMsg({
      channelId,
      coin: `${redeemAmount}${stDenomOnHost}`,
      sender: hostjs.address,
      receiver: stridejs.address,
      memo: JSON.stringify(memo),
    });

    await submitTxAndExpectSuccess(hostjs, [autopilotRedeemStake]);
    await sleep(10000); // wait for transfer to complete and tokens to be returned

    // Confirm st tokens were refunded
    const finalStBalanceOnHost = await getBalance({
      client: hostjs,
      denom: stDenomOnHost,
    });
    expect(finalStBalanceOnHost).to.equal(initialStBalanceOnHost);
  }, 120_000); // 2 min timeout
});
