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

# Contants
BONDED_TOKENS_POOL_MODULE_ADDRESS = "stride1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3ksfndm"

config = {
    "governance_voting_period": "180s",
    "epoch_day_duration": '3600s',
    "epoch_stride_duration": "3600s",
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
    for validator in genesis["validators"]:
        if validator['name'] == old_validator.moniker:
            validator['pub_key']['value'] = new_validator.pubkey

    for validator in genesis['app_state']['staking']['validators']:
        if validator['description']['moniker'] == old_validator.moniker:
            validator['consensus_pubkey']['key'] = new_validator.pubkey

    # This creates problems
    # replace(genesis, old_validator.operator_address, new_validator.operator_address)

def replace_account(genesis, old_account, new_account):

    replace(genesis, old_account.address, new_account.address)
    replace(genesis, old_account.pubkey, new_account.pubkey)

def create_parser():

    parser = argparse.ArgumentParser(
    formatter_class=argparse.RawDescriptionHelpFormatter,
    description='Create a testnet from a state export')

    parser.add_argument(
        '-c',
        '--chain-id',
        type = str,
        default="localstride",
        help='Chain ID for the testnet \nDefault: localstride\n'
    )

    parser.add_argument(
        '-i',
        '--input',
        type = str,
        default="state_export.json",
        dest='input_genesis',
        help='Path to input genesis'
    )

    parser.add_argument(
        '-o',
        '--output',
        type = str,
        default="testnet_genesis.json",
        dest='output_genesis',
        help='Path to output genesis'
    )

    parser.add_argument(
        '--validator-hex-address',
        type = str,
        help='Validator hex address to replace'
    )

    parser.add_argument(
        '--validator-operator-address',
        type = str,
        help='Validator operator address to replace'
    )

    parser.add_argument(
        '--validator-consensus-address',
        type = str,
        help='Validator consensus address to replace'
    )

    parser.add_argument(
        '--validator-pubkey',
        type = str,
        help='Validator pubkey to replace'
    )

    parser.add_argument(
        '--account-pubkey',
        type = str,
        help='Account pubkey to replace'
    )

    parser.add_argument(
        '--account-address',
        type = str,
        help='Account address to replace'
    )

    parser.add_argument(
        '--prune-ibc',
        action='store_true',
        help='Prune the IBC module'
    )

    parser.add_argument(
        '--pretty-output',
        action='store_true',
        help='Properly indent output genesis (increases time and file size)'
    )

    return parser

def main():

    parser = create_parser()
    args = parser.parse_args()

    new_validator = Validator(
        moniker = "val",
        pubkey = args.validator_pubkey,
        hex_address = args.validator_hex_address,
        operator_address = args.validator_operator_address,
        consensus_address = args.validator_consensus_address
    )

    old_validator = Validator(
        moniker = "Mendel",
        pubkey = "idsN6Oq6FjHf/woVuEo2yQfRqDcO2L3g6uJfDDJtoXo=",
        hex_address = "2F811FD9BAD33E72A674DCA98A15EBAF241341A7",
        operator_address = "stridevaloper1h2r2k24349gtx7e4kfxxl8gzqz8tn6zym65uxc",
        consensus_address = "stridevalcons197q3lkd66vl89fn5mj5c590t4ujpxsd8rus25g"
    )

    new_account = Account(
        pubkey = args.account_pubkey,
        address = args.account_address
    )

    old_account = Account(
        pubkey = "Ayyx0UKVV+w9zsTTLTGylpUH0bPON0DVdseetjVNN9eC",
        address = "stride1h2r2k24349gtx7e4kfxxl8gzqz8tn6zyc0sq2a"
    )

    print("📝 Opening {}... (it may take a while)".format(args.input_genesis))
    with open(args.input_genesis, 'r') as f:
        genesis = json.load(f)

    # Replace chain-id
    print("🔗 Replace chain-id {} with {}".format(genesis['chain_id'], args.chain_id))
    genesis['chain_id'] = args.chain_id

    # Update gov module
    print("🗳️ Update gov module")
    print("\tModify governance_voting_period from {} to {}".format(
            genesis['app_state']['gov']['voting_params']['voting_period'],
            config["governance_voting_period"]))
    genesis['app_state']['gov']['voting_params']['voting_period'] = config["governance_voting_period"]

    # Update epochs module
    print("⌛ Update epochs module")
    print("\tModify epoch_duration")
    print("\tReset current_epoch_start_time")

    for epoch in genesis['app_state']['epochs']['epochs']:
        if epoch['identifier'] == "day": 
            epoch['duration'] = config["epoch_day_duration"]

        elif epoch['identifier'] == "stride_epoch":
            epoch['duration'] = config["epoch_stride_duration"]

        epoch['current_epoch_start_time'] = datetime.now().isoformat() + 'Z'

    # Prune IBC
    if args.prune_ibc:

        print("🕸 Pruning IBC module")

        genesis['app_state']["ibc"]["channel_genesis"]["ack_sequences"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["acknowledgements"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["channels"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["commitments"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["receipts"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["recv_sequences"] = []
        genesis['app_state']["ibc"]["channel_genesis"]["send_sequences"] = []

        genesis['app_state']["ibc"]["client_genesis"]["clients"] = []
        genesis['app_state']["ibc"]["client_genesis"]["clients_consensus"] = []
        genesis['app_state']["ibc"]["client_genesis"]["clients_metadata"] = []

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

    # Replace validator pub key in genesis['app_state']['staking']['validators']
    for validator in genesis['app_state']['staking']['validators']:
        if validator['description']['moniker'] == old_validator.moniker:

            # Update delegator shares
            validator['delegator_shares'] = str(int(float(validator['delegator_shares']) + 1000000000000000)) + ".000000000000000000"
            print("\tUpdate delegator shares to {}".format(validator['delegator_shares']))

            # Update tokens
            validator['tokens'] = str(int(validator['tokens']) + 1000000000000000)
            print("\tUpdate tokens to {}".format(validator['tokens']))
            break

    # Update self delegation on operator address
    for delegation in genesis['app_state']['staking']['delegations']:
        if delegation['delegator_address'] == new_account.address:

            # delegation['validator_address'] = new_validator.operator_address
            delegation['shares'] = str(int(float(delegation['shares'])) + 1000000000000000) + ".000000000000000000"
            print("\tUpdate {} delegation shares to {} to {}".format(new_account.address, delegation['validator_address'], delegation['shares']))
            break

    # Update genesis['app_state']['distribution']['delegator_starting_infos'] on operator address
    for delegator_starting_info in genesis['app_state']['distribution']['delegator_starting_infos']:
        if delegator_starting_info['delegator_address'] == new_account.address:
            delegator_starting_info['starting_info']['stake'] = str(int(float(delegator_starting_info['starting_info']['stake']) + 1000000000000000))+".000000000000000000"
            print("\tUpdate {} stake to {}".format(delegator_starting_info['delegator_address'], delegator_starting_info['starting_info']['stake']))
            break

    print("🔋 Update validator power")

    # Update power in genesis["validators"]
    for validator in genesis["validators"]:
        if validator['name'] == old_validator.moniker:
            validator['power'] = str(int(validator['power']) + 1000000000)
            print("\tUpdate {} validator power to {}".format(validator['address'], validator['power']))
            break

    for validator_power in genesis['app_state']['staking']['last_validator_powers']:
        if validator_power['address'] == old_validator.operator_address:
            validator_power['power'] = str(int(validator_power['power']) + 1000000000)
            print("\tUpdate {} last_validator_power to {}".format(old_validator.operator_address, validator_power['power']))
            break

    # Update total power
    genesis['app_state']['staking']['last_total_power'] = str(int(genesis['app_state']['staking']['last_total_power']) + 1000000000)
    print("\tUpdate last_total_power to {}".format(genesis['app_state']['staking']['last_total_power']))

    # Update bank module
    print("💵 Update bank module")

    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == new_account.address:
            for coin in balance['coins']:
                if coin['denom'] == "ustrd":
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000)
                    print("\tUpdate {} ustrd balance to {}".format(new_account.address, coin["amount"]))
                    break
            break

    # Add 1 BN ustrd to bonded_tokens_pool module address
    for balance in genesis['app_state']['bank']['balances']:
        if balance['address'] == BONDED_TOKENS_POOL_MODULE_ADDRESS:
            # Find ustrd
            for coin in balance['coins']:
                if coin['denom'] == "ustrd":
                    coin["amount"] = str(int(coin["amount"]) + 1000000000000000)
                    print("\tUpdate {} (bonded_tokens_pool_module) ustrd balance to {}".format(BONDED_TOKENS_POOL_MODULE_ADDRESS, coin["amount"]))
                    break
            break

    # Update bank balance
    for supply in genesis['app_state']['bank']['supply']:
        if supply["denom"] == "ustrd":
            print("\tUpdate total ustrd supply from {} to {}".format(supply["amount"], str(int(supply["amount"]) + 2000000000000000)))
            supply["amount"] = str(int(supply["amount"]) + 2000000000000000)
            break

    print("📝 Writing {}... (it may take a while)".format(args.output_genesis))
    with open(args.output_genesis, 'w') as f:
        if args.pretty_output:
            f.write(json.dumps(genesis, indent=2))
        else:
            f.write(json.dumps(genesis))

if __name__ == '__main__':
    main()
