import { Coin } from "@cosmjs/amino";

export const coinFromString = (coinAsString: string): Coin => {
  const regexMatch = coinAsString.match(/^([\d\.]+)([a-z]+)$/);

  if (regexMatch === null) {
    throw new Error(`cannot extract denom & amount from '${coinAsString}'`);
  }

  return { amount: regexMatch[1], denom: regexMatch[2] };
};

export const coinsFromString = (coinsAsString: string): Coin[] =>
  coinsAsString.split(",").map(coinFromString);
