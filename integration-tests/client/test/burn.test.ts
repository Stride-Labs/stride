import { sleep, stride } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import { USTRD, STRIDE_CHAIN_NAME, MNEMONICS } from "./utils/consts";
import { waitForChain } from "./utils/startup";
import { getBalance } from "./utils/queries";
import { StrideClient } from "stridejs";
import { createStrideClient } from "./utils/setup";

describe("STRD Burn", () => {
  let strideAccounts: {
    user: StrideClient;
    admin: StrideClient;
  };

  // Initialize accounts and wait for the chain to start
  beforeAll(async () => {
    // @ts-expect-error
    strideAccounts = {};

    const admin = MNEMONICS.admin;
    const user = MNEMONICS.users[0];

    strideAccounts["admin"] = await createStrideClient(admin.mnemonic);
    strideAccounts["user"] = await createStrideClient(user.mnemonic);

    await waitForChain(STRIDE_CHAIN_NAME, strideAccounts.user, USTRD);
  }, 45_000);

  test("Burn from Single Account", async () => {
    const stridejs = strideAccounts.user;

    // Get initial balances
    const initialStrdBalance1 = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    // Burn tokens

    // Confirm burn totals

    // Burn more

    // Confirm totals again
  }, 120_000); // 2 min timeout

  test("Burn from Multiple Users", async () => {
    const stridejs1 = strideAccounts.user;
    const stridejs2 = strideAccounts.admin;

    // Get initial balances
    const initialStrdBalance1 = await getBalance({
      client: stridejs1,
      denom: USTRD,
    });

    const initialStrdBalance2 = await getBalance({
      client: stridejs2,
      denom: USTRD,
    });

    // Burn from each account

    // Confirm burn totals

    // Confirm burn amount from each
  }, 120_000); // 2 min timeout
});
