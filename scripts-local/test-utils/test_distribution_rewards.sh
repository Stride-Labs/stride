### LIQ STAKE + EXCH RATE TEST
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# import dependencies
source ${SCRIPT_DIR}/../account_vars.sh

# ADDRESSES 

# valA stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# bonded 1000000000ustrd or 1000strd 
# valB stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f (unused)
# bonded 1000000000ustrd or 1000strd

# striderly is not stride1z56v8wqvgmhm3hmnffapxujvd4w4rkw6mdrmjg
# balance: 500000000000ustrd or 500000strd
# val1 is blacklisted stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# balance: 400000000000ustrd or 400000strd

# query balances
# $STRIDE_CMD q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7
# exit

# --------------------------------------------------------------------------------
# TEST 0
# --------------------------------------------------------------------------------
# delegate from whitelisted address to valA
# $STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from striderly
# exit

# query tx
# $STRIDE_CMD q tx 94035F50CBEE933785862AC29909974387D9FC1FE549AA0130A9E472280BEC03
# exit

# query outstanding rewards (whitelisted)
# $STRIDE_CMD q distribution rewards stride1z56v8wqvgmhm3hmnffapxujvd4w4rkw6mdrmjg stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# exit
# accruing
# rewards:
# - amount: "6403050.000000000000000000"
#   denom: ustrd

# --------------------------------------------------------------------------------
# TEST 1
# --------------------------------------------------------------------------------

# delegate from blacklisted address to valA
# $STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from val1
# exit

# query outstanding rewards (blacklisted)
# $STRIDE_CMD q distribution rewards stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# exit

# rewards do not accrue
# rewards: []


# --------------------------------------------------------------------------------
# TEST 3
# --------------------------------------------------------------------------------

# wait for the block to turn
# $STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# exit

# block n
echo start block n
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# stake from blacklisted address (val1) to valA
$STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from val1 -y
# query outstanding rewards (for the validator)
$STRIDE_CMD query distribution validator-outstanding-rewards stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# query the block height
echo end block n
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# sleep 15 seconds
sleep 15

# block n+1
echo start block n+1
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# stake from blacklisted address (val1) to valA
$STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from val1 -y
# query outstanding rewards (for the validator)
$STRIDE_CMD query distribution validator-outstanding-rewards stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# query the block height
echo end block n+1
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# sleep 15 seconds
sleep 15

# block n+2
echo start block n+2
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# stake from blacklisted address (val1) to valA
$STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from val1 -y
# query outstanding rewards (for the validator)
$STRIDE_CMD query distribution validator-outstanding-rewards stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# query the block height
echo end block n+2
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# sleep 15 seconds
sleep 15

# block n+3
echo start block n+3
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# stake from blacklisted address (val1) to valA
$STRIDE_CMD tx staking delegate stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w 1000000000ustrd --from val1 -y
# query outstanding rewards (for the validator)
$STRIDE_CMD query distribution validator-outstanding-rewards stridevaloper1py0fvhdtq4au3d9l88rec6vyda3e0wttx9x92w
# query the block height
echo end block n+3
$STRIDE_CMD q tendermint-validator-set | head -n 1 | tr -dc '0-9'
# sleep 15 seconds
sleep 15
# have rewards stayed the same across blocks?
# 318541500
# 321889500
# 325237500
# 328585500
# yes!

# start block n
# 111code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: 8EB0901A59A2203869CCDAE192B6C68638F797A7BAD6ACEF7E726B4618F61C14
# rewards:
# - amount: "318541500.000000000000000000"
#   denom: ustrd
# end block n
# 111start block n+1
# 112code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: 6D7E657E9FEAF1A379C4CFDC6B58F7423801000C2277BE788ECD1D0FC4291C46
# rewards:
# - amount: "321889500.000000000000000000"
#   denom: ustrd
# end block n+1
# 112start block n+2
# 113code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: DD2AAD8B5ADB988206F4B7E3968AD94C4812EEA8179DD7E9690F9910C60C102B
# rewards:
# - amount: "325237500.000000000000000000"
#   denom: ustrd
# end block n+2
# 113start block n+3
# 114code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: 980DD7A15BA797FBABA133343529EEA3CC3A7EB260B0449DB970027C7BA7E8AC
# rewards:
# - amount: "328585500.000000000000000000"
#   denom: ustrd
# end block n+3
# 114%                                                                                                                                                                     



# start block n
# 12code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: 94035F50CBEE933785862AC29909974387D9FC1FE549AA0130A9E472280BEC03
# rewards:
# - amount: "18832500.000000000000000000"
#   denom: ustrd
# end block n
# 12start block n+1
# 13code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: D2858AB3DF0D94C0EA6AF4A58A10A5DD4CA2C3E675058FC9A79BCE9BBA966A3E
# rewards:
# - amount: "20925000.000000000000000000"
#   denom: ustrd
# end block n+1
# 13start block n+2
# 14code: 0
# codespace: ""
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: '[]'
# timestamp: ""
# tx: null
# txhash: 8EB0901A59A2203869CCDAE192B6C68638F797A7BAD6ACEF7E726B4618F61C14
# rewards:
# - amount: "23017500.000000000000000000"
#   denom: ustrd
# end block n+2
# 14start block n+3
# 15code: 19
# codespace: sdk
# data: ""
# events: []
# gas_used: "0"
# gas_wanted: "0"
# height: "0"
# info: ""
# logs: []
# raw_log: ""
# timestamp: ""
# tx: null
# txhash: 8EB0901A59A2203869CCDAE192B6C68638F797A7BAD6ACEF7E726B4618F61C14
# rewards:
# - amount: "23017500.000000000000000000"
#   denom: ustrd
# end block n+3
# 15%    
# now, let's try querying the blacklisted power from block n+2 at block n+3 (query from block n-1)
# 18832500
# 20925000
# 23017500
# 23017500
# 6:23PM ERR CONSENSUS FAILURE!!! err="negative coin amount"
