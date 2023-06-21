## How to run consumer chain

### Pre-install

Binaries:

- interchain-security-pd - [Interchain security](https://github.com/cosmos/interchain-security/v3) version: v0.2.1
- strided
- hermes(version: v0.15.0)

### Commands

```sh
rm -rf /Users/admin/.provider1
rm -rf /Users/admin/.provider
rm -rf /Users/admin/.stride1
rm -rf /Users/admin/.stride
sh run.sh
```

### Genesis modification script for consumer chain

```sh
# Add ccv section
if ! ./$PROVIDER_BINARY q provider consumer-genesis "$CONSUMER_CHAIN_ID" --node "$PROVIDER_NODE_ADDRESS" --output json > "$CONSUMER_HOME"/consumer_section.json;
then
       echo "Failed to get consumer genesis for the chain-id '$CONSUMER_CHAIN_ID'! Finalize genesis failed. For more details please check the log file in output directory."
       exit 1
fi

jq -s '.[0].app_state.ccvconsumer = .[1] | .[0]' "$CONSUMER_HOME"/config/genesis.json "$CONSUMER_HOME"/consumer_section.json > "$CONSUMER_HOME"/genesis_consumer.json && \
	mv "$CONSUMER_HOME"/genesis_consumer.json "$CONSUMER_HOME"/config/genesis.json
```

### Initial validator set on consumer chain

```json
      "initial_val_set": [
        {
          "pub_key": {
            "ed25519": "6s4FU4uSsWNjnqhNc9vhyZBqrLjib+z/mfh1LhvkalE="
          },
          "power": "1"
        },
        {
          "pub_key": {
            "ed25519": "JCFnTza2T2jlkTWxC0kY9lczh7F+jQ/bGyhHFFNr7/w="
          },
          "power": "100"
        }
      ],
```
