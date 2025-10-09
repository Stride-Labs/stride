import argparse
import json
from dataclasses import dataclass
from datetime import datetime


# Classes
@dataclass
class Validator:
    moniker: str
    pubkey: str
    hex_address: str
    operator_address: str
    consensus_address: str


@dataclass
class Account:
    pubkey: str
    address: str


# Constants
TOKEN_INCREASE = 1000000000000000
POWER_INCREASE = 1000000000
BONDED_TOKENS_POOL_MODULE_ADDRESS = "stride1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3ksfndm"
NOT_BONDED_TOKENS_POOL_MODULE_ADDRESS = "stride1tygms3xhhs3yv487phx3dw4a95jn7t7lzs4zm0"

config = {
    "governance_voting_period": "180s",
    "epoch_day_duration": "86400s",
    "epoch_stride_duration": "21600s",
}


def replace(d, old_value, new_value):
    """
    Replace all the occurences of `old_value` with `new_value`
    in `d` dictionary
    """
    for k in d.keys():
        if isinstance(d[k], dict):
            replace(d[k], old_value, new_value)
        elif isinstance(d[k], list):
            for i in range(len(d[k])):
                if isinstance(d[k][i], dict) or isinstance(d[k][i], list):
                    replace(d[k][i], old_value, new_value)
                else:
                    if d[k][i] == old_value:
                        d[k][i] = new_value
        else:
            if d[k] == old_value:
                d[k] = new_value


def replace_validator(genesis, old_validator, new_validator):
    replace(genesis, old_validator.hex_address, new_validator.hex_address)
    replace(genesis, old_validator.consensus_address, new_validator.consensus_address)

    # replace(genesis, old_validator.pubkey, new_validator.pubkey)
    for validator in genesis["consensus"]["validators"]:
        if validator["name"] == old_validator.moniker:
            validator["pub_key"]["value"] = new_validator.pubkey

    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["description"]["moniker"] == old_validator.moniker:
            validator["consensus_pubkey"]["key"] = new_validator.pubkey

    # This creates problems
    # replace(genesis, old_validator.operator_address, new_validator.operator_address)


def replace_account(genesis, old_account, new_account):

    replace(genesis, old_account.address, new_account.address)
    replace(genesis, old_account.pubkey, new_account.pubkey)


def create_parser():

    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter, description="Create a testnet from a state export"
    )

    parser.add_argument(
        "-c", "--chain-id", type=str, default="localstride", help="Chain ID for the testnet \nDefault: localstride\n"
    )

    parser.add_argument(
        "-i", "--input", type=str, default="state_export.json", dest="input_genesis", help="Path to input genesis"
    )

    parser.add_argument(
        "-o", "--output", type=str, default="testnet_genesis.json", dest="output_genesis", help="Path to output genesis"
    )

    parser.add_argument("--validator-hex-address", type=str, help="Validator hex address to replace")

    parser.add_argument("--validator-operator-address", type=str, help="Validator operator address to replace")

    parser.add_argument("--validator-consensus-address", type=str, help="Validator consensus address to replace")

    parser.add_argument("--validator-pubkey", type=str, help="Validator pubkey to replace")

    parser.add_argument("--account-pubkey", type=str, help="Account pubkey to replace")

    parser.add_argument("--account-address", type=str, help="Account address to replace")

    parser.add_argument("--prune-ibc", action="store_true", help="Prune the IBC module")

    parser.add_argument(
        "--pretty-output", action="store_true", help="Properly indent output genesis (increases time and file size)"
    )

    return parser


def main():

    parser = create_parser()
    args = parser.parse_args()

    new_validator = Validator(
        moniker="val",
        pubkey=args.validator_pubkey,
        hex_address=args.validator_hex_address,
        operator_address=args.validator_operator_address,
        consensus_address=args.validator_consensus_address,
    )

    old_validator = Validator(
        moniker="Mendel",
        pubkey="idsN6Oq6FjHf/woVuEo2yQfRqDcO2L3g6uJfDDJtoXo=",
        hex_address="2F811FD9BAD33E72A674DCA98A15EBAF241341A7",
        operator_address="stridevaloper1h2r2k24349gtx7e4kfxxl8gzqz8tn6zym65uxc",
        consensus_address="stridevalcons197q3lkd66vl89fn5mj5c590t4ujpxsd8rus25g",
    )

    new_account = Account(pubkey=args.account_pubkey, address=args.account_address)

    old_account = Account(
        pubkey="Ayyx0UKVV+w9zsTTLTGylpUH0bPON0DVdseetjVNN9eC", address="stride1h2r2k24349gtx7e4kfxxl8gzqz8tn6zyc0sq2a"
    )

    print("📝 Opening {}... (it may take a while)".format(args.input_genesis))
    with open(args.input_genesis, "r") as f:
        genesis = json.load(f)

    # Replace chain-id
    print("🔗 Replace chain-id {} with {}".format(genesis["chain_id"], args.chain_id))
    genesis["chain_id"] = args.chain_id

    # Update gov module
    print("🗳️ Update gov module")
    print(
        "\tModify governance_voting_period from {} to {}".format(
            genesis["app_state"]["gov"]["params"]["voting_period"], config["governance_voting_period"]
        )
    )
    genesis["app_state"]["gov"]["params"]["voting_period"] = config["governance_voting_period"]

    # Update epochs module
    print("⌛ Update epochs module")
    print("\tModify epoch_duration")
    print("\tReset current_epoch_start_time")

    for epoch in genesis["app_state"]["epochs"]["epochs"]:
        if epoch["identifier"] == "day":
            epoch["duration"] = config["epoch_day_duration"]

        elif epoch["identifier"] == "stride_epoch":
            epoch["duration"] = config["epoch_stride_duration"]

        epoch["current_epoch_start_time"] = datetime.now().isoformat() + "Z"

    # Prune IBC
    if args.prune_ibc:

        print("🕸 Pruning IBC module")

        genesis["app_state"]["ibc"]["channel_genesis"]["ack_sequences"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["acknowledgements"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["channels"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["commitments"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["receipts"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["recv_sequences"] = []
        genesis["app_state"]["ibc"]["channel_genesis"]["send_sequences"] = []

        genesis["app_state"]["ibc"]["client_genesis"]["clients"] = []
        genesis["app_state"]["ibc"]["client_genesis"]["clients_consensus"] = []
        genesis["app_state"]["ibc"]["client_genesis"]["clients_metadata"] = []

    # Impersonate validator
    print("🚀 Replace validator")

    # print("\t{:50} -> {}".format(old_validator.moniker, new_validator.moniker))
    print("\t{:20} {}".format("Pubkey", new_validator.pubkey))
    print("\t{:20} {}".format("Consensus address", new_validator.consensus_address))
    print("\t{:20} {}".format("Operator address", new_validator.operator_address))
    print("\t{:20} {}".format("Hex address", new_validator.hex_address))

    replace_validator(genesis, old_validator, new_validator)

    # Impersonate account
    print("🧪 Replace account")
    print("\t{:20} {}".format("Pubkey", new_account.pubkey))
    print("\t{:20} {}".format("Address", new_account.address))

    replace_account(genesis, old_account, new_account)

    # Update staking module
    print("🥩 Update staking module")

    # Replace validator pub key in genesis["app_state"]["staking"]["validators"]
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["description"]["moniker"] == old_validator.moniker:
            # Update delegator shares
            validator["delegator_shares"] = (
                str(int(float(validator["delegator_shares"]) + TOKEN_INCREASE)) + ".000000000000000000"
            )
            print("\tUpdate delegator shares to {}".format(validator["delegator_shares"]))

            # Update tokens
            validator["tokens"] = str(int(validator["tokens"]) + TOKEN_INCREASE)
            print("\tUpdate tokens to {}".format(validator["tokens"]))
            break

    # Update self delegation on operator address
    for delegation in genesis["app_state"]["staking"]["delegations"]:
        if delegation["delegator_address"] == new_account.address:
            delegation["shares"] = str(int(float(delegation["shares"])) + TOKEN_INCREASE) + ".000000000000000000"
            print(
                "\tUpdate {} delegation shares to {} to {}".format(
                    new_account.address, delegation["validator_address"], delegation["shares"]
                )
            )
            break

    # Update genesis["app_state"]["distribution"]["delegator_starting_infos"] on operator address
    for delegator_starting_info in genesis["app_state"]["distribution"]["delegator_starting_infos"]:
        if delegator_starting_info["delegator_address"] == new_account.address:
            delegator_starting_info["starting_info"]["stake"] = (
                str(int(float(delegator_starting_info["starting_info"]["stake"]) + TOKEN_INCREASE))
                + ".000000000000000000"
            )
            print(
                "\tUpdate {} stake to {}".format(
                    delegator_starting_info["delegator_address"], delegator_starting_info["starting_info"]["stake"]
                )
            )
            break

    print("🔋 Update validator power")

    # Update power in genesis["validators"]
    for validator in genesis["consensus"]["validators"]:
        if validator["name"] == old_validator.moniker:
            validator["power"] = str(int(validator["power"]) + POWER_INCREASE)
            print("\tUpdate {} validator power to {}".format(validator["address"], validator["power"]))
            break

    for validator_power in genesis["app_state"]["staking"]["last_validator_powers"]:
        if validator_power["address"] == old_validator.operator_address:
            validator_power["power"] = str(int(validator_power["power"]) + POWER_INCREASE)
            print(
                "\tUpdate {} last_validator_power to {}".format(
                    old_validator.operator_address, validator_power["power"]
                )
            )
            break

    # Update total power
    genesis["app_state"]["staking"]["last_total_power"] = str(
        int(genesis["app_state"]["staking"]["last_total_power"]) + POWER_INCREASE
    )
    print("\tUpdate last_total_power to {}".format(genesis["app_state"]["staking"]["last_total_power"]))

    # Update bank module
    print("💵 Update bank module")

    # First, update the account balance
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == new_account.address:
            for coin in balance["coins"]:
                if coin["denom"] == "ustrd":
                    coin["amount"] = str(int(coin["amount"]) + TOKEN_INCREASE)
                    print("\tUpdate {} ustrd balance to {}".format(new_account.address, coin["amount"]))
                    break
            break

    # Calculate the correct bonded pool balance from all bonded validators
    total_bonded_tokens = 0
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["status"] == "BOND_STATUS_BONDED":
            total_bonded_tokens += int(validator["tokens"])

    print(f"\tCalculated total bonded tokens: {total_bonded_tokens}")

    # Calculate not bonded tokens from unbonding and unbonded validators
    total_not_bonded_tokens = 0
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["status"] in ["BOND_STATUS_UNBONDING", "BOND_STATUS_UNBONDED"]:
            total_not_bonded_tokens += int(validator["tokens"])

    # Add unbonding delegation tokens
    for ubd in genesis["app_state"]["staking"]["unbonding_delegations"]:
        for entry in ubd["entries"]:
            total_not_bonded_tokens += int(entry["balance"])

    print(f"\tCalculated total not bonded tokens: {total_not_bonded_tokens}")

    # Updated bonded pool balance
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == BONDED_TOKENS_POOL_MODULE_ADDRESS:
            for coin in balance["coins"]:
                if coin["denom"] == "ustrd":
                    old_amount = int(coin["amount"])
                    coin["amount"] = str(total_bonded_tokens)
                    print(
                        "\tUpdate {} (bonded_tokens_pool_module) ustrd balance from {} to {}".format(
                            BONDED_TOKENS_POOL_MODULE_ADDRESS, old_amount, coin["amount"]
                        )
                    )
                    # Calculate the difference for supply adjustment
                    bonded_pool_increase = total_bonded_tokens - old_amount
                    break
            break

    # Update not bonded pool balance
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == NOT_BONDED_TOKENS_POOL_MODULE_ADDRESS:
            for coin in balance["coins"]:
                if coin["denom"] == "ustrd":
                    old_not_bonded_amount = int(coin["amount"])
                    coin["amount"] = str(total_not_bonded_tokens)
                    print(
                        "\tUpdate {} (not_bonded_tokens_pool_module) ustrd balance from {} to {}".format(
                            NOT_BONDED_TOKENS_POOL_MODULE_ADDRESS, old_not_bonded_amount, coin["amount"]
                        )
                    )
                    not_bonded_pool_adjustment = total_not_bonded_tokens - old_not_bonded_amount
                    break
            break

    # Update total supply accounting for both account increase and bonded pool adjustment
    for supply in genesis["app_state"]["bank"]["supply"]:
        if supply["denom"] == "ustrd":
            old_supply = int(supply["amount"])
            new_supply = old_supply + TOKEN_INCREASE + bonded_pool_increase + not_bonded_pool_adjustment
            print(
                "\tUpdate total ustrd supply from {} to {} (account: +{}, bonded pool: +{}, not bonded pool: +{})".format(
                    supply["amount"], new_supply, TOKEN_INCREASE, bonded_pool_increase, not_bonded_pool_adjustment
                )
            )
            supply["amount"] = str(new_supply)
            break

    print("Set governors as validators")
    # TODO: There is a check in baseapp/abci.go to see that
    # validators before init and after are the same, but that
    # isn't true in the post-ics world, in particular for
    # sovereign to consumer changeovers
    # See: https://github.com/cosmos/cosmos-sdk/blob/main/baseapp/abci.go#L114
    init_val_set = [
        {"power": val["power"], "pub_key": {"ed25519": val["pub_key"]["value"]}}
        for val in genesis["consensus"]["validators"]
    ]
    genesis["app_state"]["ccvconsumer"]["provider"]["initial_val_set"] = init_val_set

    # Update provider fee pool addr
    print("🥸  Replace Provider Fee Pool Addr")
    genesis["app_state"]["ccvconsumer"]["params"][
        "provider_fee_pool_addr_str"
    ] = "stride1h2r2k24349gtx7e4kfxxl8gzqz8tn6zyc0sq2a"
    genesis["app_state"]["ccvconsumer"]["params"]["enabled"] = True

    print("📝 Writing {}... (it may take a while)".format(args.output_genesis))
    with open(args.output_genesis, "w") as f:
        if args.pretty_output:
            f.write(json.dumps(genesis, indent=2))
        else:
            f.write(json.dumps(genesis))


if __name__ == '__main__':
    main()
