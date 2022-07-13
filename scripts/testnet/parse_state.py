import json

with open('state.json') as f:
    data = json.load(f)

ibc_denom = 'ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
STATOM_EXCH_RATE = 1.0005

bank_sends = []
genesis = []
for bank_record in data['app_state']['bank']['balances']:
    for coin_record in bank_record['coins']:
        if coin_record['denom'] == ibc_denom:
            bank_sends.append(f"strided tx bank send val2 {bank_record['address']} {coin_record['amount']}{coin_record['denom']}")
        elif coin_record['denom'] == 'ustrd':
            genesis.append(f"strided add-genesis-account {bank_record['address']} {coin_record['amount']}{coin_record['denom']}")
        elif coin_record['denom'] == 'stuatom':
            iamt = int(int(coin_record['amount']) * STATOM_EXCH_RATE)
            bank_sends.append(f"strided tx bank send val2 {bank_record['address']} {iamt}{ibc_denom}")
        else:
            raise Exception(f"Unknown denom {coin_record['denom']}")

bStr = "\n".join(bank_sends)
with open('bank_sends.sh', 'w') as f:
    f.write(bStr)

gStr = "\n".join(genesis)
with open('genesis.sh', 'w') as f:
    f.write(gStr)
