

docker-compose --ansi never exec -T stride1 strided tx ibc-transfer transfer transfer channel-1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000ustrd --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE 
docker-compose exec hermes hermes -c /tmp/hermes.toml query
docker-compose exec hermes hermes -c /tmp/hermes.toml query channel end STRIDE transfer channel-1
docker-compose exec hermes hermes -c /tmp/hermes.toml query channel end GAIA transfer channel-1

docker-compose --ansi never exec -T gaia1 gaiad tx ibc-transfer transfer transfer channel-1 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 5000uatom --from gval1 --chain-id GAIA -y --keyring-backend test --home /gaia/.gaiad
docker-compose --ansi never exec -T stride1 strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

# host zones
docker-compose --ansi never exec -T stride1 strided tx stakeibc register-host-zone connection-0 uatom ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9 --chain-id STRIDE --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
docker-compose --ansi never exec -T stride1 strided q stakeibc list-host-zone

docker-compose --ansi never exec -T stride1 strided tx ibc-transfer transfer transfer channel-1 cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv 1000ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --from val1 --chain-id STRIDE -y --keyring-backend test --home /stride/.strided/

docker-compose --ansi never exec -T gaia1 gaiad q bank balances cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv


docker-compose exec hermes hermes -c /tmp/hermes.toml query channel end GAIA transfer channel-1
docker-compose exec hermes hermes -c /tmp/hermes.toml query channels GAIA

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --chain-id STRIDE q stakeibc module-address stakeibc

docker-compose --ansi never exec -T gaia1 gaiad q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2
docker-compose --ansi never exec -T stride1 strided --home /stride/.strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided q bank balances stride1ft20pydau82pgesyl9huhhux307s9h3078692y

docker-compose --ansi never exec -T stride2 strided tx bank send val2 stride16vlrvd7lsfqg8q7kyxcyar9v7nt0h99p5arglq 10ustrd --home /stride/.strided --keyring-backend test --chain-id STRIDE -y

docker-compose --ansi never exec -T stride1 strided tx stakeibc liquid-stake 1000 ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --keyring-backend test --home /stride/.strided --from val1 --chain-id STRIDE
# ignite scaffold query module-address name:string --response addr --module stakeibc 



### LIQ STAKE + EXCH RATE TEST

## GAIA
#  ibc over atoms to stride
gaiad tx ibc-transfer transfer transfer channel-0 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 100000uatom --from gval1 --chain-id GAIA -y --keyring-backend test --home /gaia/.gaiad


## STRIDE
strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

#fund test account
# strided tx bank send val1 stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4 2500000ustrd --home /stride/.strided
#  --keyring-backend test --chain-id STRIDE -y

strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

#register gaia as host zone
strided tx stakeibc register-host-zone connection-0  uatom 'ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9' --chain-id STRIDE --home /stride/.strided --keyring-backend test --from val1 --gas 500000 -y
strided q stakeibc list-host-zone

# liquid stake atoms from stride
strided tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --home /stride/.strided --chain-id STRIDE
strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

# query balance exch rate
strided tx interchainquery query-balance GAIA cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv uatom --connection-id connection-1 --home /stride/.strided --keyring-backend test --chain-id STRIDE -y --from val1
# query exch rate
strided tx interchainquery query-exchangerate GAIA --home /stride/.strided --keyring-backend test --chain-id STRIDE -y --from val1
# query delegated balance
strided tx interchainquery query-delegatedbalance GAIA --home /stride/.strided --keyring-backend test --chain-id STRIDE -y --from val1