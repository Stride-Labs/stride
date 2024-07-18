import { Coin, StdFee } from "@cosmjs/amino";
import { fromBase64, toHex, toUtf8 } from "@cosmjs/encoding";
import { ripemd160 } from "@noble/hashes/ripemd160";
import { sha256 } from "@noble/hashes/sha256";
import { bech32 } from "bech32";

/**
 * Creates a Coin object from the given string representation of a coin.
 *
 * @example
 * ```
 * coinFromString("1ustrd") => {amount:"1",denom:"ustrd"}
 * ```
 
* @param {string} coinAsString A string representation of a coin in the format "amountdenom"
 * @returns {Coin} A Coin object with the extracted amount and denom
 */
export const coinFromString = (coinAsString: string): Coin => {
  const regexMatch = coinAsString.match(/^([\d\.]+)([a-z]+)$/);

  if (regexMatch === null) {
    throw new Error(`cannot extract denom & amount from '${coinAsString}'`);
  }

  return { amount: regexMatch[1], denom: regexMatch[2] };
};

/**
 * Converts a string of comma-separated coins into an array of Coin objects.
 *
 * @example
 * ```
 * coinsFromString("1ustrd,2uosmo") => => [{amount:"1",denom:"ustrd"},{amount:"2",denom:"uosmo"}]
 * ```
 *
 * @param {string} coinsAsString A string of comma-separated coins in the format "amountdenom,amountdenom,..."
 * @returns {Coin[]} An array of Coin objects
 */
export const coinsFromString = (coinsAsString: string): Coin[] =>
  coinsAsString.split(",").map(coinFromString);

/**
 * Creates a StdFee object from the given gas limit and gas price.
 *
 * @param {number} gasLimit The gas limit to use for the fee calculation.
 * @param {number} [gasPrice=0.025] The gas price to use for the fee calculation. Defaults to 0.025.
 * @returns {StdFee} A StdFee object with the calculated fee amount and gas limit.
 */
export const feeFromGas = (gasLimit: number, gasPrice: number = 0.025): StdFee => ({
  amount: coinsFromString(`${gasLimit * gasPrice}ustrd`),
  gas: String(gasLimit),
});

/**
 * Compute the IBC denom of a token that was sent over IBC.
 *
 * @example
 * To get the IBC denom of STRD on mainnet Osmosis:
 * ```
 * ibcDenom([{incomingPortId: "transfer", incomingChannelId: "channel-326"}], "ustrd")
 * ```
 *
 * @param {Object[]} paths An array of objects containing information about the IBC transfer paths.
 * @param {string} coinMinimalDenom The minimal denom of the coin.
 * @returns {string} The computed IBC denom of the token.
 */
export const ibcDenom = (
  paths: {
    incomingPortId: string;
    incomingChannelId: string;
  }[],
  coinMinimalDenom: string,
): string => {
  const prefixes: string[] = [];
  for (const path of paths) {
    prefixes.push(`${path.incomingPortId}/${path.incomingChannelId}`);
  }

  const prefix = prefixes.join("/");
  const denom = `${prefix}/${coinMinimalDenom}`;

  return "ibc/" + toHex(sha256(toUtf8(denom))).toUpperCase();
};

/**
 * Convert a secp256k1 compressed public key to an address
 *
 * @param {Uint8Array} pubkey The account's pubkey, should be 33 bytes (compressed secp256k1)
 * @param {String} [prefix="stride"] The address' bech32 prefix. Defaults to `"stride"`.
 * @returns the account's address
 */
export function pubkeyToAddress(pubkey: Uint8Array, prefix: string = "stride"): string {
  return bech32.encode(prefix, bech32.toWords(ripemd160(sha256(pubkey))));
}

/**
 * Convert a secp256k1 compressed public key to an address
 *
 * @param {Uint8Array} pubkey The account's pubkey as base64 string, should be 33 bytes (compressed secp256k1)
 * @param {String} [prefix="stride"] The address' bech32 prefix. Defaults to `"stride"`.
 * @returns the account's address
 */
export function base64PubkeyToAddress(pubkey: string, prefix: string = "stride"): string {
  return pubkeyToAddress(fromBase64(pubkey), prefix);
}

/**
 * Convert self delegator address to validator address
 *
 * @param {String} selfDelegator The self delegator bech32 encoded address
 * @param {String} [prefix="stride"] The self delegator address' bech32 prefix. Defaults to `"stride"`.
 * @returns the account's address
 */
export function selfDelegatorAddressToValidatorAddress(
  selfDelegator: string,
  prefix: string = "stride",
): string {
  return bech32.encode(`${prefix}valoper`, bech32.decode(selfDelegator).words);
}

/**
 * Convert self delegator address to validator address
 *
 * @param {String} validator The validator bech32 encoded address
 * @param {String} [prefix="stride"] The self delegator address' bech32 prefix. Defaults to `"stride"`.
 * @returns the account's address
 */
export function validatorAddressToSelfDelegatorAddress(
  validator: string,
  prefix: string = "stride",
): string {
  return bech32.encode(prefix, bech32.decode(validator).words);
}

/**
 * Convert a Tendermint ed25519 public key to a consensus address
 *
 * @param {Uint8Array} pubkey The tendermint pubkey, should be 32 bytes (ed25519)
 * @param {String} [prefix="stride"] The valcons address' bech32 prefix. Defaults to `"stride"`.
 * @returns the valcons account's address
 */
export function tendermintPubkeyToValconsAddress(
  pubkey: Uint8Array,
  prefix: string = "stride",
): string {
  return bech32.encode(`${prefix}valcons`, bech32.toWords(sha256(pubkey).slice(0, 20)));
}

/**
 * Convert a secp256k1 compressed public key to an address
 *
 * @param {Uint8Array} pubkey The account's pubkey as base64 string, should be 33 bytes (compressed secp256k1)
 * @param {String} [prefix="stride"] The address' bech32 prefix. Defaults to `"stride"`.
 * @returns the account's address
 */
export function base64TendermintPubkeyToValconsAddress(
  pubkey: string,
  prefix: string = "stride",
): string {
  return tendermintPubkeyToValconsAddress(fromBase64(pubkey), prefix);
}

/**
 * Sleep for a certain amount of time
 *
 * @param {number} ms The number of milliseconds to sleep for
 * @returns {Promise<void>} A promise that resolves after the sleep
 */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
