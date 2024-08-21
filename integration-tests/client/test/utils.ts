import { coinsFromString, StrideClient } from "stridejs";

/**
 * Waits for the chain to start by continuously sending transactions until .
 *
 * @param {StrideClient} client The Stride client instance.
 * @param {string} denom The denomination of the coins to send.
 */
export async function waitForChain(
  client: StrideClient,
  denom: string,
): Promise<void> {
  // the best way to ensure a chain is up is to successfully send a tx

  while (true) {
    try {
      const msg =
        client.types.cosmos.bank.v1beta1.MessageComposer.withTypeUrl.send({
          fromAddress: client.address,
          toAddress: client.address,
          amount: coinsFromString(`1${denom}`),
        });

      const tx = await client.signAndBroadcast([msg], 2);

      if (tx.code == 0) {
        break;
      }
    } catch (e) {
      // signAndBroadcast might throw if the RPC is not up yet
    }
  }
}