import { ProposalStatus, VoteOption } from "osmojs/cosmos/gov/v1beta1/gov";
import { coinsFromString, cosmos, getValueFromEvents, sleep, stride, StrideClient } from "stridejs";
import { cosmosProtoRegistry, ibcProtoRegistry, osmosisProtoRegistry } from "osmojs";
import { beforeAll, beforeEach, describe, expect, test } from "vitest";
import {
  ATOM_DENOM_ON_OSMOSIS,
  CHAIN_CONFIGS,
  MNEMONICS,
  OSMOSIS_CHAIN_NAME,
  STRIDE_CHAIN_NAME,
  COSMOSHUB_CHAIN_NAME,
  TRANSFER_CHANNEL,
  USTRD,
} from "./utils/consts";
import {
  addConcentratedLiquidityPositionMsg,
  newConcentratedLiquidityPoolMsg,
  newGammPoolMsg,
  newRegisterTokenPriceQueryMsg,
} from "./utils/msgs";
import { CosmosClient } from "./utils/types";
import { ibcTransfer, submitTxAndExpectSuccess } from "./utils/txs";
import { assertICAChannelsOpen, assertOpenTransferChannel, waitForChain, waitForIbc } from "./utils/startup";
import { getBalance, moduleAddress } from "./utils/queries";
import { createHostClient, createStrideClient, ensureHostZoneRegistered } from "./utils/setup";
import { Registry } from "@cosmjs/proto-signing";

describe("Buyback and Burn", () => {
  let strideAccounts: {
    user: StrideClient; // a normal account loaded with 100 STRD
    admin: StrideClient; // the stride admin account loaded with 1000 STRD
    val1: StrideClient;
    val2: StrideClient;
    val3: StrideClient;
  };

  let cosmoshubAccounts: {
    user: CosmosClient; // a normal account loaded with 100 ATOM
    val1: CosmosClient;
  };

  let osmosisAccounts: {
    user: CosmosClient; // a normal account loaded with 1,000,000 OSMO
    val1: CosmosClient;
  };

  const cosmoshubConfig = CHAIN_CONFIGS[COSMOSHUB_CHAIN_NAME];
  const osmosisConfig = CHAIN_CONFIGS[OSMOSIS_CHAIN_NAME];

  // init accounts and wait for chain to start
  beforeAll(async () => {
    // init {,gaia,osmo}Accounts as an empty object, then add the accounts in the loop
    // @ts-expect-error
    strideAccounts = {};
    // @ts-expect-error
    cosmoshubAccounts = {};
    // @ts-expect-error
    osmosisAccounts = {};

    const admin = MNEMONICS.admin;
    const user = MNEMONICS.users[0];
    const val1 = MNEMONICS.validators[0];
    const val2 = MNEMONICS.validators[1];
    const val3 = MNEMONICS.validators[2];

    strideAccounts["admin"] = await createStrideClient(admin.mnemonic);
    strideAccounts["user"] = await createStrideClient(user.mnemonic);
    strideAccounts["val1"] = await createStrideClient(val1.mnemonic);
    strideAccounts["val2"] = await createStrideClient(val2.mnemonic);
    strideAccounts["val3"] = await createStrideClient(val3.mnemonic);

    cosmoshubAccounts["user"] = await createHostClient(cosmoshubConfig, user.mnemonic);
    cosmoshubAccounts["val1"] = await createHostClient(cosmoshubConfig, val1.mnemonic);

    const osmoRegistry = new Registry([...osmosisProtoRegistry, ...cosmosProtoRegistry, ...ibcProtoRegistry]);
    osmosisAccounts["user"] = await createHostClient(osmosisConfig, user.mnemonic, osmoRegistry);
    osmosisAccounts["val1"] = await createHostClient(osmosisConfig, val1.mnemonic, osmoRegistry);

    await waitForChain(STRIDE_CHAIN_NAME, strideAccounts.user, USTRD);
    await waitForChain(cosmoshubConfig.chainName, cosmoshubAccounts.user, cosmoshubConfig.hostDenom);
    await waitForChain(osmosisConfig.chainName, osmosisAccounts.user, osmosisConfig.hostDenom);

    const strideToHubChannel = TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][COSMOSHUB_CHAIN_NAME]!;
    const hubToStrideChannel = TRANSFER_CHANNEL[COSMOSHUB_CHAIN_NAME][STRIDE_CHAIN_NAME]!;
    await assertOpenTransferChannel(STRIDE_CHAIN_NAME, strideAccounts.user, strideToHubChannel);
    await assertOpenTransferChannel(cosmoshubConfig.chainName, cosmoshubAccounts.user, hubToStrideChannel);

    const strideToOsmosisChannel = TRANSFER_CHANNEL[STRIDE_CHAIN_NAME][OSMOSIS_CHAIN_NAME]!;
    const osmosisToStrideChannel = TRANSFER_CHANNEL[OSMOSIS_CHAIN_NAME][STRIDE_CHAIN_NAME]!;
    await assertOpenTransferChannel(STRIDE_CHAIN_NAME, strideAccounts.user, strideToOsmosisChannel);
    await assertOpenTransferChannel(osmosisConfig.chainName, osmosisAccounts.user, osmosisToStrideChannel);

    await ensureHostZoneRegistered({
      stridejs: strideAccounts.admin,
      hostjs: cosmoshubAccounts.user,
      hostConfig: cosmoshubConfig,
    });

    await ensureHostZoneRegistered({
      stridejs: strideAccounts.admin,
      hostjs: osmosisAccounts.user,
      hostConfig: osmosisConfig,
    });

    await assertICAChannelsOpen(strideAccounts.admin, cosmoshubConfig.chainId);
    await assertICAChannelsOpen(strideAccounts.admin, osmosisConfig.chainId);
  }, 45_000);

  beforeEach(async () => {
    // Remove all token prices to not mess up tokenPriceForQuoteDenom query
    const { tokenPrices } = await strideAccounts.admin.query.stride.icqoracle.tokenPrices({});

    if (tokenPrices.length === 0) {
      return;
    }

    submitTxAndExpectSuccess(
      strideAccounts.admin,
      tokenPrices.map((tp) =>
        stride.icqoracle.MessageComposer.withTypeUrl.removeTokenPriceQuery({
          admin: strideAccounts.admin.address,
          baseDenom: tp.tokenPrice.baseDenom,
          quoteDenom: tp.tokenPrice.quoteDenom,
          osmosisPoolId: tp.tokenPrice.osmosisPoolId,
        }),
      ),
    );
  });

  describe("ICQOracle", async () => {
    test("Gamm Pool Price", async () => {
      const stridejs = strideAccounts.user;
      const osmojs = osmosisAccounts.user;

      await ibcTransfer({
        client: stridejs,
        sourceChain: STRIDE_CHAIN_NAME,
        destinationChain: OSMOSIS_CHAIN_NAME,
        coin: `1000000${USTRD}`,
        sender: stridejs.address,
        receiver: osmojs.address,
      });

      const poolMsg = newGammPoolMsg({
        sender: osmojs.address,
        tokens: [`10${osmosisConfig.hostDenom}`, `2${osmosisConfig.strdDenomOnHost}`],
        weights: [1, 1],
      });
      const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

      const osmoStrdPoolId = BigInt(getValueFromEvents(poolTx.events, "pool_created.pool_id"));

      const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
        admin: strideAccounts.admin.address,
        baseDenom: USTRD,
        quoteDenom: osmosisConfig.hostDenomOnStride,
        baseDenomOnOsmosis: osmosisConfig.strdDenomOnHost,
        quoteDenomOnOsmosis: osmosisConfig.hostDenom,
        poolId: osmoStrdPoolId,
      });
      await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);
      await sleep(2000);

      while (true) {
        const {
          tokenPrice: {
            baseDenom,
            quoteDenom,
            osmosisBaseDenom,
            osmosisQuoteDenom,
            osmosisPoolId,
            spotPrice,
            lastRequestTime,
            lastResponseTime,
          },
        } = await stridejs.query.stride.icqoracle.tokenPrice({
          baseDenom: USTRD,
          quoteDenom: osmosisConfig.hostDenomOnStride,
          poolId: osmoStrdPoolId,
        });
        if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
          expect(Number(spotPrice)).toBe(5);

          // Verify base denom matches
          expect(baseDenom).toBe(USTRD);
          expect(osmosisBaseDenom).toBe(osmosisConfig.strdDenomOnHost);

          // Verify quote denom matches
          expect(quoteDenom).toBe(osmosisConfig.hostDenomOnStride);
          expect(osmosisQuoteDenom).toBe(osmosisConfig.hostDenom);

          // Verify pool ID
          expect(osmosisPoolId).toBe(osmoStrdPoolId);

          // Verify query metadata
          expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
          expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
          expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(true);

          break;
        }
        await sleep(500);
      }
    }, 60_000); // 1 minute

    test("Concentrated Liquidity Pool Price", async () => {
      const stridejs = strideAccounts.user;
      const osmojs = osmosisAccounts.user;

      await ibcTransfer({
        client: stridejs,
        sourceChain: STRIDE_CHAIN_NAME,
        destinationChain: OSMOSIS_CHAIN_NAME,
        coin: `1000000${USTRD}`,
        sender: stridejs.address,
        receiver: osmojs.address,
      });

      const poolMsg = newConcentratedLiquidityPoolMsg({
        sender: osmojs.address,
        denom0: osmosisConfig.strdDenomOnHost,
      });
      const poolTx = await submitTxAndExpectSuccess(osmojs, poolMsg);

      const osmoStrdPoolId = BigInt(getValueFromEvents(poolTx.events, "pool_created.pool_id"));

      const addLiquidityMsg = addConcentratedLiquidityPositionMsg({
        sender: osmojs.address,
        poolId: osmoStrdPoolId,
        tokensProvided: coinsFromString(`5${osmosisConfig.strdDenomOnHost},10${osmosisConfig.hostDenom}`),
        tokenMinAmount0: "5",
        tokenMinAmount1: "10",
      });
      await submitTxAndExpectSuccess(osmojs, addLiquidityMsg);

      const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
        admin: strideAccounts.admin.address,
        baseDenom: USTRD,
        quoteDenom: osmosisConfig.hostDenomOnStride,
        baseDenomOnOsmosis: osmosisConfig.strdDenomOnHost,
        quoteDenomOnOsmosis: osmosisConfig.hostDenom,
        poolId: osmoStrdPoolId,
      });
      await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);
      await sleep(2000);

      while (true) {
        const {
          tokenPrice: {
            baseDenom,
            quoteDenom,
            osmosisBaseDenom,
            osmosisQuoteDenom,
            osmosisPoolId,
            spotPrice,
            lastRequestTime,
            lastResponseTime,
          },
        } = await stridejs.query.stride.icqoracle.tokenPrice({
          baseDenom: USTRD,
          quoteDenom: osmosisConfig.hostDenomOnStride,
          poolId: osmoStrdPoolId,
        });
        if (lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
          expect(Number(spotPrice)).toBe(2);

          // Verify base denom matches
          expect(baseDenom).toBe(USTRD);
          expect(osmosisBaseDenom).toBe(osmosisConfig.strdDenomOnHost);

          // Verify quote denom matches
          expect(quoteDenom).toBe(osmosisConfig.hostDenomOnStride);
          expect(osmosisQuoteDenom).toBe(osmosisConfig.hostDenom);

          // Verify pool ID
          expect(osmosisPoolId).toBe(osmoStrdPoolId);

          // Verify query metadata
          expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
          expect(lastRequestTime).not.toBe("0001-01-01T00:00:00.000Z");
          expect(new Date(lastResponseTime) > new Date(lastRequestTime)).toBe(true);

          break;
        }
        await sleep(500);
      }
    }, 60_000); // 1 minute

    test("Update Params", async () => {
      const stridejs = strideAccounts.user;

      const { params } = await stridejs.query.stride.icqoracle.params({});
      params.priceExpirationTimeoutSec += 1n;

      const govAddress = await moduleAddress(stridejs, "gov");

      const tx = await submitTxAndExpectSuccess(stridejs, [
        cosmos.gov.v1.MessageComposer.withTypeUrl.submitProposal({
          messages: [
            stride.icqoracle.MsgUpdateParams.toProtoMsg({
              authority: govAddress,
              params,
            }),
          ],
          initialDeposit: coinsFromString(`10000000${USTRD}`),
          proposer: stridejs.address,
          metadata: "Update icqoracle params",
          title: "Update icqoracle params",
          summary: "Update icqoracle params",
        }),
      ]);
      const proposalId = BigInt(getValueFromEvents(tx.events, "submit_proposal.proposal_id"));
      await sleep(2000);

      const txs = await Promise.all([
        strideAccounts.val1.signAndBroadcast(
          [
            cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
              proposalId: proposalId,
              voter: strideAccounts.val1.address,
              option: VoteOption.VOTE_OPTION_YES,
              metadata: "",
            }),
          ],
          2,
        ),
        strideAccounts.val2.signAndBroadcast(
          [
            cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
              proposalId: proposalId,
              voter: strideAccounts.val2.address,
              option: VoteOption.VOTE_OPTION_YES,
              metadata: "",
            }),
          ],
          2,
        ),
        strideAccounts.val3.signAndBroadcast(
          [
            cosmos.gov.v1.MessageComposer.withTypeUrl.vote({
              proposalId: proposalId,
              voter: strideAccounts.val3.address,
              option: VoteOption.VOTE_OPTION_YES,
              metadata: "",
            }),
          ],
          2,
        ),
      ]);

      for (const tx of txs) {
        if (tx.code !== 0) {
          console.error(tx.rawLog);
        }
        expect(tx.code).toBe(0);
      }

      while (true) {
        const { proposal } = await stridejs.query.cosmos.gov.v1.proposal({
          proposalId,
        });

        if (proposal?.status !== ProposalStatus.PROPOSAL_STATUS_VOTING_PERIOD) {
          expect(proposal?.status).toBe(ProposalStatus.PROPOSAL_STATUS_PASSED);
          break;
        }

        await sleep(500);
      }

      const { params: newParams } = await stridejs.query.stride.icqoracle.params({});
      expect(newParams).toStrictEqual(params);
    }, 60_000); // 1 mintue

    test("Unwrap IBC Denom", async () => {
      const stridejs = strideAccounts.admin;
      const gaiajs = cosmoshubAccounts.user;
      const osmojs = osmosisAccounts.user;

      // Transfer ATOM & OSMO to Stride to register their ibc denoms on Stride's ibc transfer app
      await ibcTransfer({
        client: gaiajs,
        sourceChain: COSMOSHUB_CHAIN_NAME,
        destinationChain: STRIDE_CHAIN_NAME,
        coin: `1${cosmoshubConfig.hostDenom}`,
        sender: gaiajs.address,
        receiver: stridejs.address,
      });

      await ibcTransfer({
        client: osmojs,
        sourceChain: OSMOSIS_CHAIN_NAME,
        destinationChain: STRIDE_CHAIN_NAME,
        coin: `1${osmosisConfig.hostDenom}`,
        sender: osmojs.address,
        receiver: stridejs.address,
      });

      const registerTokenPriceMsg = newRegisterTokenPriceQueryMsg({
        admin: strideAccounts.admin.address,
        baseDenom: cosmoshubConfig.hostDenomOnStride,
        quoteDenom: osmosisConfig.hostDenomOnStride,
        baseDenomOnOsmosis: ATOM_DENOM_ON_OSMOSIS,
        quoteDenomOnOsmosis: osmosisConfig.hostDenom,
        poolId: 1n, // not important for thie TokenPrice to work for the test to work
      });
      await submitTxAndExpectSuccess(strideAccounts.admin, registerTokenPriceMsg);
      await sleep(2000);

      const { baseDenomUnwrapped, quoteDenomUnwrapped } = await stridejs.query.stride.icqoracle.tokenPrice({
        baseDenom: cosmoshubConfig.hostDenomOnStride,
        quoteDenom: osmosisConfig.hostDenomOnStride,
        poolId: 1n,
      });

      expect(baseDenomUnwrapped).toBe(cosmoshubConfig.hostDenom);
      expect(quoteDenomUnwrapped).toBe(osmosisConfig.hostDenom);
    });
  });

  describe("Auction", async () => {
    test(
      "Bid and Burn",
      async () => {
        // - Transfer STRD to Osmosis
        // - Transfer ATOM to Osmosis
        // - Create STRD/OSMO pool
        // - Create ATOM/OSMO pool
        // - Add TokenPrice(base=STRD, quote=OSMO)
        // - Add TokenPrice(base=ATOM, quote=OSMO)
        // - Query for price of ATOM in STRD
        // - Send 10 ATOM to reward_collector account on Stride
        // - Wait for rewards to get swept from reward_collector to x/auction
        // - Create ATOM auction
        // - Buy ATOM with STRD off auction
        // - Verify STRD was burned by x/strdburner and ATOM was sent to user

        const stridejs = strideAccounts.user;
        const gaiajs = cosmoshubAccounts.val1;
        const osmojs = osmosisAccounts.user;

        // Transfer STRD to Osmosis
        await ibcTransfer({
          client: stridejs,
          sourceChain: STRIDE_CHAIN_NAME,
          destinationChain: OSMOSIS_CHAIN_NAME,
          coin: `1000000${USTRD}`,
          sender: stridejs.address,
          receiver: osmojs.address,
        });

        // Transfer ATOM to Osmosis
        await ibcTransfer({
          client: gaiajs,
          sourceChain: COSMOSHUB_CHAIN_NAME,
          destinationChain: STRIDE_CHAIN_NAME,
          coin: `1000000${cosmoshubConfig.hostDenom}`,
          sender: gaiajs.address,
          receiver: stridejs.address, // needs to be valid but ignored by pfm
          memo: JSON.stringify({
            forward: {
              receiver: osmojs.address,
              port: "transfer",
              channel: TRANSFER_CHANNEL.stride.osmosis,
            },
          }),
        });

        // Create STRD/OSMO pool
        const createClPoolTx = await submitTxAndExpectSuccess(
          osmojs,
          newConcentratedLiquidityPoolMsg({
            sender: osmojs.address,
            denom0: osmosisConfig.strdDenomOnHost,
          }),
        );
        await sleep(2000);

        const osmoStrdPoolId = BigInt(getValueFromEvents(createClPoolTx.events, "pool_created.pool_id"));

        await submitTxAndExpectSuccess(
          osmojs,
          addConcentratedLiquidityPositionMsg({
            poolId: osmoStrdPoolId,
            sender: osmojs.address,
            tokensProvided: coinsFromString(`5${osmosisConfig.strdDenomOnHost},10${osmosisConfig.hostDenom}`),
            tokenMinAmount0: "5",
            tokenMinAmount1: "10",
          }),
        );
        await sleep(2000);

        // Create ATOM/OSMO pool
        const createGammPoolTx = await submitTxAndExpectSuccess(
          osmojs,
          newGammPoolMsg({
            sender: osmojs.address,
            tokens: [`10${osmosisConfig.hostDenom}`, `2${ATOM_DENOM_ON_OSMOSIS}`],
            weights: [1, 1],
          }),
        );

        const osmoAtomPoolId = BigInt(getValueFromEvents(createGammPoolTx.events, "pool_created.pool_id"));

        // Add TokenPrice(base=STRD, quote=OSMO)
        await submitTxAndExpectSuccess(
          strideAccounts.admin,
          stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
            admin: strideAccounts.admin.address,
            baseDenom: USTRD,
            quoteDenom: osmosisConfig.hostDenomOnStride,
            osmosisBaseDenom: osmosisConfig.strdDenomOnHost,
            osmosisQuoteDenom: osmosisConfig.hostDenom,
            osmosisPoolId: osmoStrdPoolId,
          }),
        );
        await sleep(2000);

        // Add TokenPrice(base=ATOM, quote=OSMO)
        await submitTxAndExpectSuccess(
          strideAccounts.admin,
          stride.icqoracle.MessageComposer.withTypeUrl.registerTokenPriceQuery({
            admin: strideAccounts.admin.address,
            baseDenom: cosmoshubConfig.hostDenomOnStride,
            quoteDenom: osmosisConfig.hostDenomOnStride,
            osmosisBaseDenom: ATOM_DENOM_ON_OSMOSIS,
            osmosisQuoteDenom: osmosisConfig.hostDenom,
            osmosisPoolId: osmoAtomPoolId,
          }),
        );

        // Wait for both TokenPrices to be updated
        while (true) {
          const { tokenPrice } = await stridejs.query.stride.icqoracle.tokenPrice({
            baseDenom: USTRD,
            quoteDenom: osmosisConfig.hostDenomOnStride,
            poolId: osmoStrdPoolId,
          });
          if (tokenPrice.lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
            expect(Number(tokenPrice.spotPrice)).toBe(2);
            break;
          }
          await sleep(500);
        }
        while (true) {
          const { tokenPrice } = await stridejs.query.stride.icqoracle.tokenPrice({
            baseDenom: cosmoshubConfig.hostDenomOnStride,
            quoteDenom: osmosisConfig.hostDenomOnStride,
            poolId: osmoAtomPoolId,
          });
          if (tokenPrice.lastResponseTime.toISOString() != "0001-01-01T00:00:00.000Z") {
            expect(Number(tokenPrice.spotPrice)).toBe(5);
            break;
          }
          await sleep(500);
        }

        // Query for price of ATOM in STRD
        const { price } = await stridejs.query.stride.icqoracle.tokenPriceForQuoteDenom({
          baseDenom: cosmoshubConfig.hostDenomOnStride,
          quoteDenom: USTRD,
        });

        // Price should be 2.5:
        //
        // TODO: Tind a better way to test this.
        // This will fail if other tests set the price to be something different
        //
        // ATOM/OSMO pool is 2/10 => 1 ATOM is 5 OSMO
        // STRD/OSMO pool is 5/10 => 1 STRD is 2 OSMO
        // =>
        // 2.5 STRD is 5 OSMO
        // =>
        // 1 ATOM is 2.5 STRD
        expect(Number(price)).toBe(2.5);

        const rewardAmount = 10_000000;
        const rewardCollectorAddress = await moduleAddress(stridejs, "reward_collector");
        // Send 10 ATOM to reward_collector account on Stride
        await ibcTransfer({
          // 1740472679470067779
          // 1740472677000000000
          client: gaiajs,
          sourceChain: COSMOSHUB_CHAIN_NAME,
          destinationChain: STRIDE_CHAIN_NAME,
          coin: `${rewardAmount}${cosmoshubConfig.hostDenom}`,
          sender: gaiajs.address,
          receiver: rewardCollectorAddress,
        });

        // Wait for funds to get swept from reward_collector to auction
        const auctionAddress = await moduleAddress(stridejs, "auction");

        let auctionAtomBalance: string;
        while (true) {
          ({ balance: { amount: auctionAtomBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
              address: auctionAddress,
              denom: cosmoshubConfig.hostDenomOnStride,
            }));

          if (BigInt(auctionAtomBalance) > 0n) {
            break;
          }

          const { balance: { amount: rewardCollectorAtomBalance } = { amount: "0" } } =
            await stridejs.query.cosmos.bank.v1beta1.balance({
              address: rewardCollectorAddress,
              denom: cosmoshubConfig.hostDenomOnStride,
            });

          await sleep(500);
        }

        // Create ATOM auction
        const auctionName = "ATOM" + Math.random();
        const { address: strdburnerAddress } = await stridejs.query.stride.strdburner.strdBurnerAddress({});

        await submitTxAndExpectSuccess(
          strideAccounts.admin,
          stride.auction.MessageComposer.withTypeUrl.createAuction({
            admin: strideAccounts.admin.address,
            auctionName,
            auctionType: stride.auction.AuctionType.AUCTION_TYPE_FCFS,
            sellingDenom: cosmoshubConfig.hostDenomOnStride,
            paymentDenom: USTRD,
            enabled: true,
            minPriceMultiplier: "0.95",
            minBidAmount: "1",
            beneficiary: strdburnerAddress,
          }),
        );

        // Buy ATOM with STRD off auction and verify STRD was burned and ATOM was sent to user
        const { totalBurned: totalBurnedBefore } = await stridejs.query.stride.strdburner.totalStrdBurned({});
        const { balance: { amount: userAtomBalanceBefore } = { amount: "0" } } =
          await stridejs.query.cosmos.bank.v1beta1.balance({
            address: strideAccounts.user.address,
            denom: cosmoshubConfig.hostDenomOnStride,
          });

        // price is 2.5 and BigInt doesn't support fractions so we'll multiply by 10 for a price of 25
        // and then divide by 10
        const atomsToBuy = BigInt(auctionAtomBalance) / 100n;
        const strdToPay = (BigInt(Number(price) * 10) * atomsToBuy) / 10n;

        await submitTxAndExpectSuccess(
          strideAccounts.user,
          stride.auction.MessageComposer.withTypeUrl.placeBid({
            bidder: strideAccounts.user.address,
            auctionName,
            sellingTokenAmount: String(atomsToBuy),
            paymentTokenAmount: String(strdToPay),
          }),
        );

        const { totalBurned: totalBurnedAfter } = await stridejs.query.stride.strdburner.totalStrdBurned({});
        const { balance: { amount: userAtomBalanceAfter } = { amount: "0" } } =
          await stridejs.query.cosmos.bank.v1beta1.balance({
            address: strideAccounts.user.address,
            denom: cosmoshubConfig.hostDenomOnStride,
          });

        expect(BigInt(userAtomBalanceAfter)).toBe(BigInt(userAtomBalanceBefore) + atomsToBuy);
        expect(BigInt(totalBurnedAfter)).toBe(BigInt(totalBurnedBefore) + strdToPay);
      },
      5 * 60 * 1000 /* 5min */,
    );

    test("Staking Rewards to x/auction", async () => {
      const stridejs = strideAccounts.admin;
      const gaiajs = cosmoshubAccounts.user;

      const auctionAddress = await moduleAddress(stridejs, "auction");
      const auctionBalanceBefore = await getBalance({
        client: stridejs,
        address: auctionAddress,
        denom: cosmoshubConfig.hostDenomOnStride,
      });

      const stakeAmount = 10_000_000;
      const rewardAmount = 10_000;
      const feeAmount = 1_000;

      // Liquid stake 10 ATOM
      await ibcTransfer({
        client: gaiajs,
        sourceChain: COSMOSHUB_CHAIN_NAME,
        destinationChain: STRIDE_CHAIN_NAME,
        coin: `${stakeAmount}${cosmoshubConfig.hostDenom}`,
        sender: gaiajs.address,
        receiver: stridejs.address,
      });

      const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
        creator: stridejs.address,
        amount: String(stakeAmount),
        hostDenom: cosmoshubConfig.hostDenom,
      });

      await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);
      sleep(1000);

      // Check st tokens
      const stAtomBalance = await getBalance({
        client: stridejs,
        address: stridejs.address,
        denom: cosmoshubConfig.stDenom,
      });
      expect(BigInt(stAtomBalance)).toBeGreaterThan(0n);

      // Send 10% of stake to fee address
      // If we send more, you risk tripping some rate limits
      const {
        hostZone: { withdrawalIcaAddress },
      } = await stridejs.query.stride.stakeibc.hostZone({
        chainId: cosmoshubConfig.chainId,
      });

      await submitTxAndExpectSuccess(gaiajs, [
        cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
          fromAddress: gaiajs.address,
          toAddress: withdrawalIcaAddress,
          amount: coinsFromString(`${rewardAmount}${cosmoshubConfig.hostDenom}`),
        }),
      ]);

      // Wait for funds to get swept from fee address on gaia into x/auction
      while (true) {
        const auctionBalanceAfter = await getBalance({
          client: stridejs,
          address: auctionAddress,
          denom: cosmoshubConfig.hostDenomOnStride,
        });

        if (BigInt(auctionBalanceAfter) >= BigInt(auctionBalanceBefore) + BigInt(feeAmount)) {
          break;
        }

        await sleep(500);
      }
    }, 360_000);
  });
});
