
import requests
import json 
import pandas as pd

TESTNET = "internal"

STRIDE_ADDR = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"

STRIDE_BANK_QUERY = f"https://stride-node3.{TESTNET}.stridenet.co/bank/balances/ADDRESS"
GAIA_BANK_QUERY = f"https://gaia.{TESTNET}.stridenet.co/bank/balances/ADDRESS"

GAIA_VALIDATOR_QUERY = "https://gaia.internal.stridenet.co/staking/validators"

SEPARATOR = "\n\n============================================================\n\n"

def query_stride_team_balance():
    url = STRIDE_BANK_QUERY.replace("ADDRESS", STRIDE_ADDR)
    r = requests.get(url=url)
    data = r.json()
    out = {}
    for res in data['result']:
        out[res['denom']] = res['amount']
    return out

def query_stride_team_balance():
    url = STRIDE_BANK_QUERY.replace("ADDRESS", STRIDE_ADDR)
    r = requests.get(url=url)
    data = r.json()
    out = {}
    for res in data['result']:
        out[res['denom']] = res['amount']
    return out

def query_host_zone_info():
    url = f"https://stride-node3.{TESTNET}.stridenet.co/Stride-Labs/stride/stakeibc/host_zone/GAIA"
    r = requests.get(url=url)
    data = r.json()
    return data

def get_host_zone_accts(host_zone):
    accts = ["delegationAccount", "feeAccount", "redemptionAccount", "withdrawalAccount"]
    out = {}
    hz = host_zone['HostZone']
    for acct in accts:
        out[acct] = {
            'Address': hz[acct]['address'],
            'EstBalance': hz[acct]['balance'],
            'target': hz[acct]['target'],
        }
        out['TrueBalance'] = query_gaia
        out['TrueStaked'] = 
    return out

def query_gaia_validators():
    url = f"https://gaia.{TESTNET}.stridenet.co/staking/validators"
    r = requests.get(url=url)
    data = r.json()
    out = {}
    for validator in data['result']:
        out[validator["operator_address"]] = validator['tokens']
    return out

def fmt_dict(d, prefix=''):
    out = ""
    for k, v in d.items():
        out += f"{prefix}{k}: {v}\n"
    return out

def fmt_json(j):
    return json.dumps(j, indent=4, sort_keys=True)

stride_team_balance = query_stride_team_balance()
host_zone_info = query_host_zone_info()


# Get Gaia info
gaia_validators = query_gaia_validators()

print(SEPARATOR)
print("STRIDE TEAM BALANCE")
print(fmt_dict(stride_team_balance))
print(SEPARATOR)
print("HOST ZONE INFO")
print(fmt_json(host_zone_info))
print(SEPARATOR)
print("\nGAIA INFO")
print("\tVALIDATORS")
print(fmt_dict(gaia_validators, prefix="\t\t"))

# gaia_addr_balance = query_stride_team_balance()
