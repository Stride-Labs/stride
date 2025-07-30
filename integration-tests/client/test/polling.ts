import { StrideClient } from "stridejs";
import { REMOVED } from "./consts";
import { CosmosClient } from "./types";
import { sleep } from "stridejs";
import { getBalance, getDelegatedBalance, getHostZone } from "./queries";
import { bigIntAbs } from "./utils";
import { expect } from "vitest";

/**
 * Wait for a balance to change (increase from initial value)
 * @param client The stride client
 * @param address The address to query the balance for
 * @param denom The token denom to check the balance for
 * @param initialBalance Optional initial balance to compare against
 * If not provided, will query the current balance
 * @param minChange Min change between queries before the waiting is over
 */
export async function waitForBalanceChange({
  client,
  address,
  denom,
  initialBalance,
  minChange = 0,
}: {
  client: StrideClient | CosmosClient;
  address: string;
  denom: string;
  initialBalance?: bigint;
  minChange?: number;
}): Promise<bigint> {
  const maxAttempts = 60;
  let attempt = 0;
  let prevBalance = initialBalance === undefined ? await getBalance({ client, address, denom }) : initialBalance;

  while (attempt < maxAttempts) {
    const currBalance = await getBalance({ client, address, denom });
    if (bigIntAbs(currBalance - prevBalance) >= BigInt(minChange)) {
      return currBalance;
    }

    attempt++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for balance change for ${denom} at ${address}`);
}

/**
 * Wait for a delegation to occur on the host zone
 * @param client The stride client
 * @param chainId The chainId of the host zone
 * @param minDelegation The minimum delegation before returning a success
 */
export async function waitForHostZoneTotalDelegationsChange({
  client,
  chainId,
  minDelegation = 0,
}: {
  client: StrideClient;
  chainId: string;
  minDelegation?: number;
}): Promise<bigint> {
  const maxAttempts = 180;
  let attempt = 0;

  while (attempt < maxAttempts) {
    let { totalDelegations: currDelegations } = await getHostZone({ client, chainId });

    if (BigInt(currDelegations) >= BigInt(minDelegation)) {
      return BigInt(currDelegations);
    }

    attempt++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for the host zone struct's delegated balance to reach ${minDelegation}`);
}

/**
 * Wait for the deposit record to change to status
 * @param client The stride client
 * @param depositRecordId The ID of the deposit record to search for
 * @param status The status of the deposit record that we're waiting for
 * If we're waiting for a record to be removed, pass the string "REMOVED"
 */
export async function waitForDepositRecordStatus({
  client,
  depositRecordId,
  status,
}: {
  client: StrideClient;
  depositRecordId: bigint;
  status: any;
}): Promise<void> {
  const maxAttempts = 60;
  let attempt = 0;

  while (attempt < maxAttempts) {
    // If we're checking that the record was removed, query all records and check that the ID is not found
    if (status === REMOVED) {
      const { depositRecord } = await client.query.stride.records.depositRecordAll();
      if (depositRecord.filter((record) => record.id == depositRecordId).length == 0) {
        return;
      }
    } else {
      // Otherwise, if we're checking for a status, query the record by ID and check the status
      const { depositRecord } = await client.query.stride.records.depositRecord({ id: depositRecordId });
      if (depositRecord.status === status) {
        return;
      }
    }

    attempt++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for delegation record to change to status: ${status.toString()}`);
}

/**
 * Wait for the epoch unbonding records to change to the specified status
 * @param client The stride client
 * @param chainId The chain ID of the host zone
 * @param epochNumber The epoch number of the relevant record
 * @param status The target status of the record
 */
export async function waitForUnbondingRecordStatus({
  client,
  chainId,
  epochNumber,
  status,
}: {
  client: StrideClient;
  chainId: string;
  epochNumber: bigint;
  status: any;
}): Promise<void> {
  const maxAttempts = 360;
  let attempt = 0;

  while (attempt < maxAttempts) {
    // If we're checking that the record was removed, query all records and check that the ID is not found
    if (status === REMOVED) {
      const { epochUnbondingRecord: epochUnbondingRecords } =
        await client.query.stride.records.epochUnbondingRecordAll();

      if (epochUnbondingRecords.filter((record) => record.epochNumber == epochNumber).length == 0) {
        return;
      }
    } else {
      // Otherwise, if we're checking for a status, query the record by ID and check the status
      const { epochUnbondingRecord } = await client.query.stride.records.epochUnbondingRecord({ epochNumber });
      const hostZoneUnbondingRecords = epochUnbondingRecord.hostZoneUnbondings.filter(
        (record) => record.hostZoneId == chainId,
      );
      expect(hostZoneUnbondingRecords.length).to.equal(
        1,
        `No unbonding record found for ${chainId} and epoch ${epochNumber}`,
      );

      if (hostZoneUnbondingRecords[0].status === status) {
        return;
      }
    }

    attempt++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for unbonding record to change to status: ${status.toString()}`);
}

/**
 * Wait for a redemption record to be removed after a claim is complete
 * @param client The stride client
 * @param redemptionRecordId The ID of the redemption record
 */
export async function waitForRedemptionRecordRemoval({
  client,
  redemptionRecordId,
}: {
  client: StrideClient;
  redemptionRecordId: string;
}): Promise<void> {
  const maxAttempts = 360;
  let attempt = 0;

  while (attempt < maxAttempts) {
    const { userRedemptionRecord: userRedemptionRecords } = await client.query.stride.records.userRedemptionRecordAll();
    if (userRedemptionRecords.filter((record) => record.id == redemptionRecordId).length == 0) {
      return;
    }

    attempt++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for redemption record ${redemptionRecordId} to be removed`);
}
