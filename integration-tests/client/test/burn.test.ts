import { sleep, stride } from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import { USTRD, STRIDE_CHAIN_NAME, MNEMONICS, DEFAULT_FEE } from "./utils/consts";
import { waitForChain } from "./utils/startup";
import { getBalance } from "./utils/queries";
import { StrideClient } from "stridejs";
import { createStrideClient } from "./utils/setup";
import { newBurnStrdMsg, newLinkMsg } from "./utils/msgs";
import { submitTxAndExpectSuccess } from "./utils/txs";

describe("STRD Burner", () => {
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
    await sleep(3_000);
  }, 45_000);

  test("Burn from Single Account", async () => {
    const stridejs = strideAccounts.user;
    const burnAmount1 = BigInt(2_000_000);
    const burnAmount2 = BigInt(3_000_000);

    // Get initial balances
    const initialBalance = await getBalance({
      client: stridejs,
      denom: USTRD,
    });

    // Get initial burn amounts
    const { totalBurned: totalBurnedBefore, totalUserBurned: totalUserBurnedBefore } =
      await stridejs.query.stride.strdburner.totalStrdBurned({});
    const { burnedAmount: userBurnedBefore } = await stridejs.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs.address,
    });

    // Burn tokens
    console.log("Burning tokens...");
    const burnMsg1 = newBurnStrdMsg({ burner: stridejs.address, amount: burnAmount1 });
    await submitTxAndExpectSuccess(stridejs, [burnMsg1]);
    await sleep(2_000);

    // Confirm burn totals
    const { totalBurned: totalBurnedAfter1, totalUserBurned: totalUserBurnedAfter1 } =
      await stridejs.query.stride.strdburner.totalStrdBurned({});
    const { burnedAmount: userBurnedAfter1 } = await stridejs.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs.address,
    });

    expect(BigInt(totalBurnedAfter1)).to.equal(BigInt(totalBurnedBefore) + burnAmount1, "total burn #1");
    expect(BigInt(totalUserBurnedAfter1)).to.equal(BigInt(totalUserBurnedBefore) + burnAmount1, "total user burn #1");
    expect(BigInt(userBurnedAfter1)).to.equal(BigInt(userBurnedBefore) + burnAmount1, "user burn #1");

    // Confirm balance decreased
    const updatedBalance1 = await getBalance({
      client: stridejs,
      denom: USTRD,
    });
    expect(updatedBalance1).to.equal(initialBalance - burnAmount1 - DEFAULT_FEE, "balance after first burn");

    // Burn more
    console.log("Burning more tokens...");
    const burnMsg2 = newBurnStrdMsg({ burner: stridejs.address, amount: burnAmount2 });
    await submitTxAndExpectSuccess(stridejs, [burnMsg2]);
    await sleep(2_000);

    // Confirm burn totals again
    const { totalBurned: totalBurnedAfter2, totalUserBurned: totalUserBurnedAfter2 } =
      await stridejs.query.stride.strdburner.totalStrdBurned({});
    const { burnedAmount: userBurnedAfter2 } = await stridejs.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs.address,
    });

    expect(BigInt(totalBurnedAfter2)).to.equal(BigInt(totalBurnedBefore) + burnAmount1 + burnAmount2, "total burn #2");
    expect(BigInt(totalUserBurnedAfter2)).to.equal(
      BigInt(totalUserBurnedBefore) + burnAmount1 + burnAmount2,
      "total user burn #2",
    );
    expect(BigInt(userBurnedAfter2)).to.equal(BigInt(userBurnedBefore) + burnAmount1 + burnAmount2, "user burn #2");

    // Confirm balance decreased
    const updatedBalance2 = await getBalance({
      client: stridejs,
      denom: USTRD,
    });
    expect(updatedBalance2).to.equal(updatedBalance1 - burnAmount2 - DEFAULT_FEE, "balance after second burn");
  }, 120_000); // 2 min timeout

  test("Burn from Multiple Users", async () => {
    const stridejs1 = strideAccounts.user;
    const stridejs2 = strideAccounts.admin;

    const burnAmount1 = BigInt(2_000_000);
    const burnAmount2 = BigInt(3_000_000);

    // Get initial balances
    const initialBalance1 = await getBalance({
      client: stridejs1,
      denom: USTRD,
    });

    const initialBalance2 = await getBalance({
      client: stridejs2,
      denom: USTRD,
    });

    // Get initial burn amounts
    const { burnedAmount: userBurnedBefore1 } = await stridejs1.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs1.address,
    });
    const { burnedAmount: userBurnedBefore2 } = await stridejs2.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs2.address,
    });

    // Burn from each account
    console.log("Burning tokens...");
    const burnMsg1 = newBurnStrdMsg({ burner: stridejs1.address, amount: burnAmount1 });
    await submitTxAndExpectSuccess(stridejs1, [burnMsg1]);

    const burnMsg2 = newBurnStrdMsg({ burner: stridejs2.address, amount: burnAmount2 });
    await submitTxAndExpectSuccess(stridejs2, [burnMsg2]);
    await sleep(2_000);

    // Confirm burn amount from each
    const { burnedAmount: userBurnedAfter1 } = await stridejs1.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs1.address,
    });
    const { burnedAmount: userBurnedAfter2 } = await stridejs2.query.stride.strdburner.strdBurnedByAddress({
      address: stridejs2.address,
    });

    expect(BigInt(userBurnedAfter1)).to.equal(BigInt(userBurnedBefore1) + burnAmount1, "user 1 burn");
    expect(BigInt(userBurnedAfter2)).to.equal(BigInt(userBurnedBefore2) + burnAmount2, "user 2 burn");

    // Confirm balances
    const updatedBalance1 = await getBalance({
      client: stridejs1,
      denom: USTRD,
    });

    const updatedBalance2 = await getBalance({
      client: stridejs2,
      denom: USTRD,
    });

    expect(updatedBalance1).to.equal(initialBalance1 - burnAmount1 - DEFAULT_FEE, "user 1 balance");
    expect(updatedBalance2).to.equal(initialBalance2 - burnAmount2 - DEFAULT_FEE, "user 2 balance");
  }, 120_000); // 2 min timeout

  test("Link Address", async () => {
    const stridejs = strideAccounts.user;
    const linkedAddress1 = "0x1";
    const linkedAddress2 = "0x2";

    // Link first address
    const linkMsg1 = newLinkMsg({ strideAddress: stridejs.address, linkedAddress: linkedAddress1 });
    await submitTxAndExpectSuccess(stridejs, [linkMsg1]);
    await sleep(2_000);

    // Confirm link
    const { linkedAddress: actualLinkedAddress1 } = await stridejs.query.stride.strdburner.linkedAddress({
      strideAddress: stridejs.address,
    });
    expect(actualLinkedAddress1).to.equal(linkedAddress1, "first address link");

    // Re-link with a new address
    const linkMsg2 = newLinkMsg({ strideAddress: stridejs.address, linkedAddress: linkedAddress2 });
    await submitTxAndExpectSuccess(stridejs, [linkMsg2]);
    await sleep(2_000);

    // Confirm again
    const { linkedAddress: actualLinkedAddress2 } = await stridejs.query.stride.strdburner.linkedAddress({
      strideAddress: stridejs.address,
    });
    expect(actualLinkedAddress2).to.equal(linkedAddress2, "second address link");
  });
});
