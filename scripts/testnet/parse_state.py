import json

with open('state.json') as f:
    data = json.load(f)

IGNORE_ADDRS = {'stride159atdlc3ksl50g0659w5tq42wwer334ajl7xnq', # stride main
                'stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan', # module
                'stride1m3h30wlvsf8llruxtpukdvsy0km2kum8t68ynv', # module
                'stride1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8y5yqan', # module
                'stride1yl6hdjhmkf37639730gffanpzndzdpmhd5k4r0', # module
                'stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl', # module
                'stride1vlthgax23ca9syk7xgaz347xmf4nunefyrktcu', # module
                'stride1mvdq4nlupl39243qjz7sds5ez3rl9mnx253lza', # module
                'stride1lqnk7sldxed4u4pf90cqvk4mkt08dtp7gzdgkl', # module
                'stride1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3ksfndm', # staking pool module? 
                'stride1eqahrt8nu2xx394puzppuy49csmu2688ktuy6v', # faucet 
                'stride1zfrw4r3lnlvw5v3m5fgckayvqlhx5l30yahna8', # base account
                'stride1ad22g9hscw35v7tq3d28c3kek79knn0msjyw7f', # base account
                'stride158pufn0quh57d57lagq5uqnm5ssk92ftultnrv', # ???
}

ibc_denom = 'ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2'
STATOM_EXCH_RATE = 1.0005

bank_sends = []
genesis = []
genesis_local = []

bank_suffix = "--keyring-backend=test -y"

gen_amts = {}

for delegation_record in data['app_state']['staking']['delegations']:
    addr = delegation_record['delegator_address']
    if addr in IGNORE_ADDRS:
        continue
    if addr not in gen_amts:
        gen_amts[addr] = 0
    gen_amts[addr] += int(float(delegation_record['shares']))

for bank_record in data['app_state']['bank']['balances']:
    if bank_record['address'] in IGNORE_ADDRS:
        continue
    for coin_record in bank_record['coins']:
        if int(coin_record['amount']) > 1000000000:
            continue
        if coin_record['denom'] == ibc_denom:
            bank_sends.append(f"strided tx bank send val2 {bank_record['address']} {coin_record['amount']}{coin_record['denom']} {bank_suffix}")
        elif coin_record['denom'] == 'ustrd':
            if bank_record['address'] not in gen_amts:
                gen_amts[bank_record['address']] = 0
            gen_amts[bank_record['address']] += int(coin_record['amount'])
        elif coin_record['denom'] == 'stuatom':
            iamt = int(int(coin_record['amount']) * STATOM_EXCH_RATE)
            bank_sends.append(f"strided tx bank send val2 {bank_record['address']} {iamt}{ibc_denom} {bank_suffix}")
        else:
            raise Exception(f"Unknown denom {coin_record['denom']}")

for addr, amt in gen_amts.items():
    if amt > 1000000000:
        continue
    genesis_local.append(f"$STRIDE_CMD add-genesis-account {addr} {amt}ustrd")
    genesis.append(f"$MAIN_NODE_CMD add-genesis-account {addr} {amt}ustrd")

bStr = "\nsleep 12\n".join(bank_sends)
with open('bank_sends.sh', 'w') as f:
    f.write(bStr)

gStr = "\n".join(genesis_local)
with open('../../scripts-local/genesis.sh', 'w') as f:
    f.write(gStr)

gStr = "\n".join(genesis)
with open('genesis.sh', 'w') as f:
    f.write(gStr)
