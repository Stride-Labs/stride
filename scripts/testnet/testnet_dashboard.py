
import requests
import json 
import pandas as pd

TESTNET = "poolparty"

STRIDE_TEAM_ADDR = "stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq"

STRIDE_BANK_QUERY = f"https://stride-node3.{TESTNET}.stridenet.co/bank/balances/ADDRESS"
GAIA_BANK_QUERY = f"https://gaia.{TESTNET}.stridenet.co/bank/balances/ADDRESS"

GAIA_VALIDATOR_QUERY = "https://gaia.internal.stridenet.co/staking/validators"

SEPARATOR = "\n\n============================================================\n\n"

def query_balance(addr, net="stride"):
    if net == 'stride':
        url = STRIDE_BANK_QUERY.replace("ADDRESS", addr)
    elif net == 'gaia':
        url = GAIA_BANK_QUERY.replace("ADDRESS", addr)
    else:
        raise Exception(f"Unknown network {net}")
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
    if "HostZone" not in host_zone:
        return {}
    hz = host_zone['HostZone']
    for acct in accts:
        if (acct not in hz) or (hz[acct] is None):
            out[acct] = {}
            continue 
        out[acct] = {
            'Address': hz[acct]['address'],
            'EstBalance': hz[acct]['balance'],
            'target': hz[acct]['target'],
        }
        def parse_balance(addr):
            qb = query_balance(addr, net='gaia')
            if 'uatom' not in qb:
                qb['uatom'] = 0
            return qb['uatom']
        out[acct]['TrueBalance'] = parse_balance(out[acct]['Address'])
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

stride_team_balance = query_balance(STRIDE_TEAM_ADDR)
host_zone_info = query_host_zone_info()


# Get Gaia info
gaia_validators = query_gaia_validators()
gaia_accts = pd.DataFrame(get_host_zone_accts(host_zone_info)).T

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
print("\tACCOUNTS")
print(gaia_accts)

# gaia_addr_balance = query_stride_team_balance()
