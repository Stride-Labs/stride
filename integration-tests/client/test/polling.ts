import { StrideClient } from "stridejs";
import { REMOVED } from "./consts";
import { CosmosClient } from "./types";
import { sleep } from "stridejs";
import { getBalance, getDelegatedBalance } from "./queries";
import { bigIntAbs } from "./utils";

/**
 * Wait for a balance to change (increase from initial value)
 */
export async function waitForBalanceChange({
  client,
  address,
  denom,
  initialBalance,
  minChange = 0,
  maxAttempts = 60,
}: {
  client: StrideClient | CosmosClient;
  address: string;
  denom: string;
  initialBalance?: bigint;
  minChange?: number;
  maxAttempts?: number;
}): Promise<bigint> {
  let attempts = 0;
  let prevBalance = initialBalance === undefined ? await getBalance({ client, address, denom }) : initialBalance;

  while (attempts < maxAttempts) {
    const currBalance = await getBalance({ client, address, denom });
    if (bigIntAbs(currBalance - prevBalance) >= BigInt(minChange)) {
      return currBalance;
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for balance change for ${denom} at ${address}`);
}

/**
 * Wait for a delegation to occur on the host zone
 * @param client The cosmos client of the host zone
 * @param delegator The delegator's address
 * @param minChange The minimum change before returning a success
 * @param maxAttempts The max number of attempts to try, each spaced by a second
 */
export async function waitForDelegationChange({
  client,
  delegator,
  minChange = 0,
  maxAttempts = 60,
}: {
  client: StrideClient | CosmosClient;
  delegator: string;
  minChange?: number;
  maxAttempts?: number;
}): Promise<bigint> {
  let attempts = 0;
  let prevBalance = await getDelegatedBalance({ client, delegator });

  while (attempts < maxAttempts) {
    const currBalance = await getDelegatedBalance({ client, delegator });
    if (bigIntAbs(currBalance - prevBalance) >= BigInt(minChange)) {
      return currBalance;
    }

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for delegated balance change at ${delegator}`);
}

/**
 * Wait for the deposit record to change to status DELEGATION_QUEUE
 */
export async function waitForDepositRecordStatus({
  client,
  depositRecordId,
  status,
  maxAttempts = 60,
}: {
  client: StrideClient;
  depositRecordId: bigint;
  status: any;
  maxAttempts?: number;
}): Promise<void> {
  let attempts = 0;

  while (attempts < maxAttempts) {
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

    attempts++;
    await sleep(1000); // 1 second
  }

  throw new Error(`Timed out waiting for delegation record to change to status: ${status.toString()}`);
}
