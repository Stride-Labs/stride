import {
    QueryClient,
    setupAuthExtension,
    setupBankExtension,
    setupStakingExtension,
    setupTxExtension,
    SigningStargateClient,
} from "@cosmjs/stargate";
import { Comet38Client } from "@cosmjs/tendermint-rpc";
import {
    DirectSecp256k1HdWallet,
    GasPrice,
    sleep,
    stride,
} from "stridejs";
import { beforeAll, describe, expect, test } from "vitest";
import {
    ATOM_DENOM_ON_STRIDE,
    GAIA_CHAIN_ID,
    GAIA_RPC_ENDPOINT,
    STRIDE_RPC_ENDPOINT,
    TRANSFER_CHANNEL,
    UATOM,
    USTRD,
    STRD_DENOM_ON_GAIA,
} from "./consts";
import { CosmosClient } from "./types";
import {
    ibcTransfer,
    waitForChain,
    waitForIbc,
    submitTxAndExpectSuccess,
} from "./utils";
import { StrideClient } from "stridejs";

// Utility function to get balance as a string
async function getBalance({
                              client,
                              address,
                              denom,
                          }: {
    client: StrideClient | CosmosClient;
    address: string;
    denom: string;
}): Promise<string> {
    if (client instanceof StrideClient) {
        const { balance: { amount } = { amount: "0" } } =
            await client.query.cosmos.bank.v1beta1.balance({
                address,
                denom,
            });
        return amount;
    } else {
        const balance = await client.query.bank.balance(address, denom);
        return balance.amount;
    }
}

// Initialize accounts
let strideAccounts: {
    user: StrideClient;
    admin: StrideClient;
    val1: StrideClient;
    val2: StrideClient;
    val3: StrideClient;
};

let gaiaAccounts: {
    user: CosmosClient;
    val1: CosmosClient;
};

const mnemonics: {
    name: "user" | "admin" | "val1" | "val2" | "val3";
    mnemonic: string;
}[] = [
    {
        name: "user",
        mnemonic:
            "brief play describe burden half aim soccer carbon hope wait output play vacuum joke energy crucial output mimic cruise brother document rail anger leaf",
    },
    {
        name: "admin",
        mnemonic:
            "tone cause tribe this switch near host damage idle fragile antique tail soda alien depth write wool they rapid unfold body scan pledge soft",
    },
    {
        name: "val1",
        mnemonic:
            "close soup mirror crew erode defy knock trigger gather eyebrow tent farm gym gloom base lemon sleep weekend rich forget diagram hurt prize fly",
    },
    {
        name: "val2",
        mnemonic:
            "turkey miss hurry unable embark hospital kangaroo nuclear outside term toy fall buffalo book opinion such moral meadow wing olive camp sad metal banner",
    },
    {
        name: "val3",
        mnemonic:
            "tenant neck ask season exist hill churn rice convince shock modify evidence armor track army street stay light program harvest now settle feed wheat",
    },
];

// Initialize accounts and wait for the chain to start
beforeAll(async () => {
    console.log("setting up accounts...");

    // @ts-expect-error
    strideAccounts = {};
    // @ts-expect-error
    gaiaAccounts = {};

    for (const { name, mnemonic } of mnemonics) {
        // setup signer for Stride
        const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
            prefix: "stride",
        });

        const [{ address }] = await signer.getAccounts();

        strideAccounts[name] = await StrideClient.create(
            STRIDE_RPC_ENDPOINT,
            signer,
            address,
            {
                gasPrice: GasPrice.fromString(`0.025${USTRD}`),
                broadcastPollIntervalMs: 50,
                resolveIbcResponsesCheckIntervalMs: 50,
            },
        );

        if (name === "user" || name === "val1") {
            // setup signer for Gaia
            const gaiaSigner = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);
            const [{ address: gaiaAddress }] = await gaiaSigner.getAccounts();

            gaiaAccounts[name] = {
                address: gaiaAddress,
                client: await SigningStargateClient.connectWithSigner(
                    GAIA_RPC_ENDPOINT,
                    gaiaSigner,
                    {
                        gasPrice: GasPrice.fromString(`1.0${UATOM}`),
                        broadcastPollIntervalMs: 50,
                    },
                ),
                query: QueryClient.withExtensions(
                    await Comet38Client.connect(GAIA_RPC_ENDPOINT),
                    setupAuthExtension,
                    setupBankExtension,
                    setupStakingExtension,
                    setupTxExtension,
                ),
            };
        }
    }

    console.log("waiting for stride to start...");
    await waitForChain(strideAccounts.user, USTRD);

    console.log("waiting for gaia to start...");
    await waitForChain(gaiaAccounts.user, UATOM);

    console.log("waiting for stride-gaia ibc...");
    await waitForIbc(
        strideAccounts.user,
        TRANSFER_CHANNEL.STRIDE.GAIA!,
        USTRD,
        "cosmos",
    );
}, 45_000);

describe("Core Tests", () => {
    test("IBC Transfer", async () => {
        const stridejs = strideAccounts.user;
        const gaiajs = gaiaAccounts.user;
        const transferAmount = 50000000;

        // Get initial balances
        const strideInitialStrdBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: USTRD,
        });

        const strideInitialAtomBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: ATOM_DENOM_ON_STRIDE,
        });

        const gaiaInitialStrdBalance = await getBalance({
            client: gaiajs,
            address: gaiajs.address,
            denom: STRD_DENOM_ON_GAIA,
        });

        const gaiaInitialAtomBalance = await getBalance({
            client: gaiajs,
            address: gaiajs.address,
            denom: UATOM,
        });

        console.log("Initial balances:");
        console.log(`Stride USTRD: ${strideInitialStrdBalance}`);
        console.log(`Stride ATOM: ${strideInitialAtomBalance}`);
        console.log(`Gaia STRD: ${gaiaInitialStrdBalance}`);
        console.log(`Gaia ATOM: ${gaiaInitialAtomBalance}`);

        // Perform IBC transfers
        console.log("Transferring USTRD from Stride to Gaia...");
        await ibcTransfer({
            client: stridejs,
            sourceChain: "STRIDE",
            destinationChain: "GAIA",
            coin: `${transferAmount}${USTRD}`,
            sender: stridejs.address,
            receiver: gaiajs.address,
        });

        console.log("Transferring ATOM from Gaia to Stride...");
        await ibcTransfer({
            client: gaiajs,
            sourceChain: "GAIA",
            destinationChain: "STRIDE",
            coin: `${transferAmount}${UATOM}`,
            sender: gaiajs.address,
            receiver: stridejs.address,
        });

        // Wait a bit for transfers to complete
        await sleep(5000);

        // Get final balances
        const strideFinalStrdBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: USTRD,
        });

        const strideFinalAtomBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: ATOM_DENOM_ON_STRIDE,
        });

        const gaiaFinalStrdBalance = await getBalance({
            client: gaiajs,
            address: gaiajs.address,
            denom: STRD_DENOM_ON_GAIA,
        });

        const gaiaFinalAtomBalance = await getBalance({
            client: gaiajs,
            address: gaiajs.address,
            denom: UATOM,
        });

        console.log("Final balances:");
        console.log(`Stride USTRD: ${strideFinalStrdBalance}`);
        console.log(`Stride ATOM: ${strideFinalAtomBalance}`);
        console.log(`Gaia STRD: ${gaiaFinalStrdBalance}`);
        console.log(`Gaia ATOM: ${gaiaFinalAtomBalance}`);

        // Calculate and verify balance changes
        const strideStrdBalanceDiff = BigInt(strideFinalStrdBalance) - BigInt(strideInitialStrdBalance);
        const strideAtomBalanceDiff = BigInt(strideFinalAtomBalance) - BigInt(strideInitialAtomBalance);
        const gaiaStrdBalanceDiff = BigInt(gaiaFinalStrdBalance) - BigInt(gaiaInitialStrdBalance);
        const gaiaAtomBalanceDiff = BigInt(gaiaFinalAtomBalance) - BigInt(gaiaInitialAtomBalance);

        console.log("Balance differences:");
        console.log(`Stride STRD diff: ${strideStrdBalanceDiff}`);
        console.log(`Stride ATOM diff: ${strideAtomBalanceDiff}`);
        console.log(`Gaia STRD diff: ${gaiaStrdBalanceDiff}`);
        console.log(`Gaia ATOM diff: ${gaiaAtomBalanceDiff}`);

        // Verify the transfers worked
        // STRD sent out from Stride → negative balance change
        expect(strideStrdBalanceDiff).toBeLessThanOrEqual(BigInt(-transferAmount));
        expect(strideStrdBalanceDiff).toBeGreaterThan(BigInt(-transferAmount - 1000000)); // gas fee limit

        // ATOM received on Stride → positive balance change
        expect(strideAtomBalanceDiff).toBe(BigInt(transferAmount));

        // STRD received on Gaia → positive balance change
        expect(gaiaStrdBalanceDiff).toBe(BigInt(transferAmount));

        // ATOM sent out from Gaia → negative balance change
        expect(gaiaAtomBalanceDiff).toBeLessThanOrEqual(BigInt(-transferAmount));
        expect(gaiaAtomBalanceDiff).toBeGreaterThan(BigInt(-transferAmount - 1000000)); // gas fee limit
    }, 120_000);

    test("Liquid Stake Mint and Transfer", async () => {
        const stridejs = strideAccounts.user;
        const gaiajs = gaiaAccounts.user;
        const stakeAmount = 10000000;

        // Get initial balances
        const strideInitialAtomBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: ATOM_DENOM_ON_STRIDE,
        });

        const strideInitialStAtomBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: "stuatom",
        });

        // Get delegation address
        const { hostZone } = await stridejs.query.stride.stakeibc.hostZone({
            chainId: GAIA_CHAIN_ID,
        });
        const delegationAddress = hostZone?.delegationIcaAddress || "";

        // Get initial delegation ICA balance
        const delegationInitialBalance = await getBalance({
            client: gaiajs,
            address: delegationAddress,
            denom: UATOM,
        });

        console.log("Initial balances:");
        console.log(`Stride ATOM: ${strideInitialAtomBalance}`);
        console.log(`Stride stATOM: ${strideInitialStAtomBalance}`);
        console.log(`Delegation ICA ATOM: ${delegationInitialBalance}`);

        // Perform liquid staking
        const liquidStakeMsg = stride.stakeibc.MessageComposer.withTypeUrl.liquidStake({
            creator: stridejs.address,
            amount: String(stakeAmount),
            hostDenom: UATOM,
        });

        await submitTxAndExpectSuccess(stridejs, [liquidStakeMsg]);

        // Wait for stTokens to be minted
        let strideFinalStAtomBalance;
        let attempts = 0;
        const maxAttempts = 60; // 30 seconds max wait time

        while (attempts < maxAttempts) {
            const currentStAtomBalance = await getBalance({
                client: stridejs,
                address: stridejs.address,
                denom: "stuatom",
            });

            if (BigInt(currentStAtomBalance) > BigInt(strideInitialStAtomBalance)) {
                strideFinalStAtomBalance = currentStAtomBalance;
                break;
            }

            attempts++;
            await sleep(500);
        }

        if (attempts >= maxAttempts) {
            throw new Error("Timed out waiting for stATOM tokens to be minted");
        }

        // Get final ATOM balance on Stride
        const strideFinalAtomBalance = await getBalance({
            client: stridejs,
            address: stridejs.address,
            denom: ATOM_DENOM_ON_STRIDE,
        });

        // Wait for tokens to be transferred to the delegation account
        let delegationFinalBalance;
        let delegationAttempts = 0;
        const delegationMaxAttempts = 120; // 60 seconds max wait time

        while (delegationAttempts < delegationMaxAttempts) {
            delegationFinalBalance = await getBalance({
                client: gaiajs,
                address: delegationAddress,
                denom: UATOM,
            });

            if (BigInt(delegationFinalBalance) > BigInt(delegationInitialBalance)) {
                break;
            }

            delegationAttempts++;
            await sleep(500);
        }

        if (delegationAttempts >= delegationMaxAttempts) {
            throw new Error("Timed out waiting for tokens to be transferred to delegation account");
        }

        console.log("Final balances:");
        console.log(`Stride ATOM: ${strideFinalAtomBalance}`);
        console.log(`Stride stATOM: ${strideFinalStAtomBalance}`);
        console.log(`Delegation ICA ATOM: ${delegationFinalBalance}`);

        // Calculate balance differences
        const strideAtomBalanceDiff = BigInt(strideFinalAtomBalance) - BigInt(strideInitialAtomBalance);
        const strideStAtomBalanceDiff = BigInt(strideFinalStAtomBalance) - BigInt(strideInitialStAtomBalance);
        const delegationBalanceDiff = BigInt(delegationFinalBalance) - BigInt(delegationInitialBalance);

        console.log("Balance differences:");
        console.log(`Stride ATOM diff: ${strideAtomBalanceDiff}`);
        console.log(`Stride stATOM diff: ${strideStAtomBalanceDiff}`);
        console.log(`Delegation ICA diff: ${delegationBalanceDiff}`);

        // Verify balance changes
        expect(strideAtomBalanceDiff).toBe(BigInt(-stakeAmount)); // ATOM should decrease (sent for staking)
        expect(strideStAtomBalanceDiff).toBe(BigInt(stakeAmount)); // stATOM should increase (minted)
        expect(delegationBalanceDiff).toBe(BigInt(stakeAmount)); // Delegation ICA should receive tokens
    }, 180_000); // 3 minutes timeout
});
