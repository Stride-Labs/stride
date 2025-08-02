import { sleep, stride } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
  USTRD,
  DEFAULT_FEE,
  REMOVED,
  CHAIN_CONFIGS,
  STRIDE_CHAIN_NAME,
  MNEMONICS,
  TEST_CHAINS,
  TRANSFER_CHANNEL,
} from "./utils/consts";
import { CosmosClient } from "./utils/types";
import { ibcTransfer, submitTxAndExpectSuccess } from "./utils/txs";
import { waitForChain, assertICAChannelsOpen, assertOpenTransferChannel } from "./utils/startup";
import {
  waitForDepositRecordStatus,
  waitForUnbondingRecordStatus,
  waitForRedemptionRecordRemoval,
  waitForRedemptionRateChange,
} from "./utils/polling";
import {
  getBalance,
  getDelegatedBalance,
  getHostZoneTotalDelegations,
  getHostZoneUnbondingRecord,
  getLatestDepositRecord,
  getLatestHostZoneUnbondingRecord,
  getRedemptionAccountBalance,
  getUserRedemptionRecord,
} from "./utils/queries";
import { StrideClient } from "stridejs";
import { Decimal } from "decimal.js";
import { getHostZone } from "./utils/queries";
import {
  createHostClient,
  createStrideClient,
  ensureHostZoneRegistered,
  ensureLiquidStakeExists,
  ensureNativeHostTokensOnStride,
} from "./utils/setup";

describe.each(TEST_CHAINS)("Core Liquid Staking - %s", (hostChainName) => {
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

  test("IBC Transfer", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const transferAmount = BigInt(50000000);

    const { chainName, hostDenom, strdDenomOnHost, hostDenomOnStride } = HOST_CONFIG;

    // Get initial balances
    // We'll send STRD from Stride -> Host Zone
    const initialStrdBalanceOnStride = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const initialStrdBalanceOnHost = await getBalance({
      client: hostjs,
      denom: strdDenomOnHost,
    });

    // As well as host denom (e.g. ATOM) from Host chain -> Stride
    const initialHostBalanceOnHost = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    const initialHostBalanceOnStride = await getBalance({
      client: stridejs,
      denom: hostDenomOnStride,
    });

    // Perform IBC transfers
    console.log("Transferring USTRD from Stride to host zone...");
    await ibcTransfer({
      client: stridejs,
      sourceChain: STRIDE_CHAIN_NAME,
      destinationChain: chainName,
      coin: `${transferAmount}${USTRD}`,
      sender: stridejs.address,
      receiver: hostjs.address,
    });

    console.log("Transferring native host token from host zone to Stride...");
    await ibcTransfer({
      client: hostjs,
      sourceChain: chainName,
      destinationChain: STRIDE_CHAIN_NAME,
      coin: `${transferAmount}${hostDenom}`,
      sender: hostjs.address,
      receiver: stridejs.address,
    });

    // Wait a bit for transfers to complete.
    console.log("Waiting for transfers to complete...");
    await sleep(5000);

    // Get final balances
    const finalStrdBalanceOnStride = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    const finalStrdBalanceOnHost = await getBalance({
      client: hostjs,
      denom: strdDenomOnHost,
    });

    const finalHostBalanceOnHost = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    const finalHostBalanceOnStride = await getBalance({
      client: stridejs,
      denom: hostDenomOnStride,
    });

    // Calculate and verify balance changes
    const strideStrdBalanceDiff = finalStrdBalanceOnStride - initialStrdBalanceOnStride;
    const strideHostBalanceDiff = finalHostBalanceOnStride - initialHostBalanceOnStride;
    const hostStrdBalanceDiff = finalStrdBalanceOnHost - initialStrdBalanceOnHost;
    const hostHostBalanceDiff = finalHostBalanceOnHost - initialHostBalanceOnHost;

    // Verify the transfers worked
    // STRD sent out from Stride → negative balance change + fee
    // STRD received on host → positive balance change
    expect(strideStrdBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Stride STRD balance change");
    expect(hostStrdBalanceDiff).to.equal(transferAmount, "Host STRD balance change");

    // Host denom sent out from host zone → negative balance change + fee
    // Host denom received on Stride → positive balance change
    expect(hostHostBalanceDiff).to.equal(-(transferAmount + DEFAULT_FEE), "Host native balance change");
    expect(strideHostBalanceDiff).to.equal(transferAmount, "Stride host denom balance change");
  }, 120_000); // 2 min timeout

  test("Liquid Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;

    const { chainId, hostDenom, stDenom, hostDenomOnStride } = HOST_CONFIG;

    // Get initial balances on Stride
    let initialUserNativeBalance = await getBalance({
      client: stridejs,
      denom: hostDenomOnStride,
    });

    const initialUserStBalance = await getBalance({
      client: stridejs,
      denom: stDenom,
    });

    // Get the initial delegated balance
    const initialDelegatedBalance = await getDelegatedBalance({
      stridejs,
      hostjs,
      chainId,
    });

    // Ensure there's enough native host tokens to liquid stake, if not transfer
    initialUserNativeBalance = await ensureNativeHostTokensOnStride({
      stridejs,
      hostjs,
      hostChainConfig: HOST_CONFIG,
      minAmount: stakeAmount,
    });

    // Perform liquid staking
    console.log("Liquid staking...");
    const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
      creator: stridejs.address,
      amount: String(stakeAmount),
      hostDenom,
    });

    const tx = await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
    await sleep(2000); // sleep to make sure block finalized

    // Get final native and st balances
    const finalUserStBalance = await getBalance({ client: stridejs, address: stridejs.address, denom: stDenom });
    const finalUserNativeBalance = await getBalance({
      client: stridejs,
      address: stridejs.address,
      denom: hostDenomOnStride,
    });

    // Get the redemption rate at the time of the liquid stake
    const { redemptionRate } = await getHostZone({ client: stridejs, chainId, blockHeight: tx.height });

    // Confirm the balance changes
    // Native balance should decrease (sent for staking)
    // StBalance should increase (minted)
    const nativeBalanceDiff = finalUserNativeBalance - initialUserNativeBalance;
    const stBalanceDiff = finalUserStBalance - initialUserStBalance;
    const expectedStTokensMinted = BigInt(
      Decimal(stakeAmount.toString()).div(Decimal(redemptionRate)).floor().toString(),
    );
    expect(stBalanceDiff).to.equal(expectedStTokensMinted, "User st balance change on Stride");
    expect(nativeBalanceDiff).to.equal(BigInt(-stakeAmount), "User native balance change on Stride");

    // Get the deposit record that was used for the liquid stake
    // We grab the latest TRANSFER_QUEUE record
    const depositRecord = await getLatestDepositRecord({
      client: stridejs,
      chainId,
      blockHeight: tx.height,
      status: stride.records.DepositRecord_Status.TRANSFER_QUEUE,
    });
    const depositRecordId = depositRecord.id;
    expect(BigInt(depositRecord.amount) >= BigInt(stakeAmount)).to.be.true;

    // Wait for the the transfer to complete by checking for when the deposit record
    // changes to state DELEGATION_QUEUE
    console.log("Waiting for transfer to complete...");
    await waitForDepositRecordStatus({
      client: stridejs,
      depositRecordId,
      status: stride.records.DepositRecord_Status.DELEGATION_QUEUE,
    });

    // Wait for delegation to complete by checking the ICA account's delegations
    console.log("Waiting for delegation to complete...");
    await waitForDepositRecordStatus({
      client: stridejs,
      depositRecordId,
      status: REMOVED,
    });

    // Confirm at least our staked amount was delegated (it could be more if there was reinvestment)
    const updatedDelegatedBalance = await getDelegatedBalance({ stridejs, hostjs, chainId });
    expect(updatedDelegatedBalance - initialDelegatedBalance >= BigInt(stakeAmount)).to.be.true;

    // Confirm the host zone updated
    const hostZoneTotalDelegations = await getHostZoneTotalDelegations({ client: stridejs, chainId });
    expect(hostZoneTotalDelegations).to.equal(updatedDelegatedBalance, "Updated delegated balance");
  }, 180_000); // 3 min timeout

  test("Redeem Stake", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 10000000;
    const redeemAmount = 1000000;

    const { chainId, hostDenom, stDenom } = HOST_CONFIG;

    // Ensure there's enough liquid stake to cover the redemption
    await ensureLiquidStakeExists({
      stridejs,
      hostjs,
      chainId,
      hostChainConfig: HOST_CONFIG,
      minAmount: stakeAmount,
    });

    // Get the initial delegated balance both internally and the ground truth
    const initialTotalDelegations = await getHostZoneTotalDelegations({ client: stridejs, chainId });
    const initialDelegatedBalance = await getDelegatedBalance({ stridejs, hostjs, chainId });

    // Before redeeming, get the initial st balance
    const initialUserStBalance = await getBalance({
      client: stridejs,
      denom: stDenom,
    });

    // Submit redeem stake tx
    console.log("Redeeming stake...");
    const redeemStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.redeemStake({
      creator: stridejs.address,
      amount: String(redeemAmount),
      hostZone: chainId,
      receiver: hostjs.address,
    });

    const tx = await submitTxAndExpectSuccess(stridejs, [redeemStakeMsg]);
    await sleep(2000); // sleep to make sure block finalized

    // Confirm the st balance was decremented as the tokens were burned
    const finalUserStBalance = await getBalance({ client: stridejs, address: stridejs.address, denom: stDenom });
    const stBalanceDiff = initialUserStBalance - finalUserStBalance;
    expect(stBalanceDiff).to.equal(BigInt(redeemAmount), "User st balance change after redemption");

    // Get the epoch number from the unbonding record that corresponds to this redemption
    const { epochNumber: unbondingEpoch, hostZoneUnbonding: unbondingRecord } = await getLatestHostZoneUnbondingRecord({
      client: stridejs,
      chainId,
      status: stride.records.HostZoneUnbonding_Status.UNBONDING_QUEUE,
      blockHeight: tx.height,
    });

    // Get the redemption rate at the time of the claim and calculate the native tokens expected
    const { redemptionRate } = await getHostZone({ client: stridejs, chainId, blockHeight: tx.height });
    const expectedNativeAmount = BigInt(
      Decimal(redeemAmount.toString()).mul(Decimal(redemptionRate)).floor().toString(),
    );

    // Confirm a user redemption record was created with the proper amounts
    const redemptionRecord = await getUserRedemptionRecord({
      client: stridejs,
      chainId,
      epochNumber: unbondingEpoch,
      receiver: hostjs.address,
    });
    expect(BigInt(redemptionRecord.stTokenAmount)).to.equal(BigInt(redeemAmount), "Redemption record st amount");
    expect(BigInt(redemptionRecord.nativeTokenAmount)).to.equal(
      expectedNativeAmount,
      "Redemption record native amount",
    );

    // Wait for the undelegation to submit (by waiting for the record to be in status EXIT_TRANSFER_QUEUE)
    console.log("Waiting for undelegation...");
    await waitForUnbondingRecordStatus({
      client: stridejs,
      chainId,
      epochNumber: unbondingEpoch,
      status: stride.records.HostZoneUnbonding_Status.EXIT_TRANSFER_QUEUE,
    });

    // Confirm the delegated balance changed (both on the host zone struct and on the actual host chain)
    // There should simulateously be reinvestment delegations, so we have to relax the check to within a tolerance
    let updatedTotalDelegations = await getHostZoneTotalDelegations({ client: stridejs, chainId });
    let updatedDelegatedBalance = await getDelegatedBalance({ stridejs, hostjs, chainId });

    const expectedDelegationChange = BigInt(unbondingRecord.nativeTokenAmount);
    const delegationChangeLowerBound = BigInt(Decimal(expectedDelegationChange).mul(0.98).floor().toString()); // within 2%
    const delegationChangeUpperBound = BigInt(expectedDelegationChange);

    const totalDelegationsDiff = initialTotalDelegations - updatedTotalDelegations;
    const delegatedBalanceDiff = initialDelegatedBalance - updatedDelegatedBalance;

    expect(totalDelegationsDiff >= delegationChangeLowerBound).to.be.true;
    expect(totalDelegationsDiff <= delegationChangeUpperBound).to.be.true;
    expect(delegatedBalanceDiff >= delegationChangeLowerBound).to.be.true;
    expect(delegatedBalanceDiff <= delegationChangeUpperBound).to.be.true;

    // Fetch the unbonding record and user redemption record again
    // These get updated when the undelegation is submitted with the true number of native tokens owed
    // (factoring in re-investment that occurs in between redeem and unbond)
    const updatedUnbondingRecord = await getHostZoneUnbondingRecord({
      client: stridejs,
      chainId,
      epochNumber: unbondingEpoch,
    });
    const updatedRedemptionRecord = await getUserRedemptionRecord({
      client: stridejs,
      chainId,
      epochNumber: unbondingEpoch,
      receiver: hostjs.address,
    });

    // Get the initial redemption account balance
    const initialRedemptionICABalance = await getRedemptionAccountBalance({ stridejs, hostjs, chainId });

    // Wait for the tokens to be sent to the redemption account
    console.log("Waiting for redemption sweep...");
    await waitForUnbondingRecordStatus({
      client: stridejs,
      chainId,
      epochNumber: unbondingEpoch,
      status: stride.records.HostZoneUnbonding_Status.CLAIMABLE,
    });

    // Confirm the redemption account balance increased
    const redemptionBalanceAfterSweep = await getRedemptionAccountBalance({ stridejs, hostjs, chainId });
    const redemptionBalanceDiffAfterSweep = redemptionBalanceAfterSweep - initialRedemptionICABalance;
    expect(redemptionBalanceDiffAfterSweep).to.equal(
      BigInt(updatedUnbondingRecord.nativeTokenAmount),
      "Redemption ICA balance change",
    );

    // Get the user's native balance before the claim
    const initialUserNativeBalanceOnHost = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });

    // Claim the unbonded tokens
    console.log("Claiming redeemed tokens...");
    const claimMsg = stride.stakeibc.MessageComposer.withTypeUrl.claimUndelegatedTokens({
      creator: stridejs.address,
      hostZoneId: chainId,
      epoch: unbondingEpoch,
      receiver: hostjs.address,
    });
    await submitTxAndExpectSuccess(stridejs, [claimMsg]);

    // Wait for the redemption record to be removed after the claim - indicating that the tokens
    // have been successfully transferred
    await waitForRedemptionRecordRemoval({ client: stridejs, redemptionRecordId: updatedRedemptionRecord.id });

    // Confirm the user received those tokens
    const finalUserNativeBalanceOnHost = await getBalance({
      client: hostjs,
      denom: hostDenom,
    });
    const nativeBalanceDiff = finalUserNativeBalanceOnHost - initialUserNativeBalanceOnHost;

    expect(nativeBalanceDiff).to.equal(BigInt(updatedRedemptionRecord.nativeTokenAmount));
    expect(nativeBalanceDiff >= expectedNativeAmount).to.be.true;
    expect(nativeBalanceDiff <= BigInt(Decimal(expectedNativeAmount.toString()).mul(1.001).floor().toString())).to.be
      .true;

    // Confirm the redemption ICA decremented
    const redemptionBalanceAfterClaim = await getRedemptionAccountBalance({ stridejs, hostjs, chainId });
    const redemptionBalanceDiffAfterClaim = redemptionBalanceAfterSweep - redemptionBalanceAfterClaim;
    expect(redemptionBalanceDiffAfterClaim).to.equal(
      BigInt(updatedRedemptionRecord.nativeTokenAmount),
      "Redemption balance after claim",
    );
  }, 600_000); // 10 minute timeout

  test("Reinvestment", async () => {
    const stridejs = strideAccounts.user;
    const hostjs = hostAccounts.user;
    const stakeAmount = 1000000;

    const { chainId } = HOST_CONFIG;

    // Ensure there's an active liquid stake for reinvestment testing
    await ensureLiquidStakeExists({
      stridejs,
      hostjs,
      chainId,
      hostChainConfig: HOST_CONFIG,
      minAmount: stakeAmount,
    });

    // Wait for the redemption rate to change after the reinvestment kicks off
    await waitForRedemptionRateChange({ client: stridejs, chainId });

    // Check that the redemption rate is greater than 1 - meaning there was reinvestment
    const { redemptionRate } = await getHostZone({ client: stridejs, chainId });
    expect(Decimal(redemptionRate).greaterThan(Decimal(1))).to.be.true;
  }, 600_000); // 10 min timeout
});
