alias GETBAL="head -n 1 | grep -o -E '[0-9]+'"

echo "amount 1035" | GETBAL



docker-compose --ansi never exec -T gaia1 gaiad q bank balances cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2
docker-compose --ansi never exec -T stride1 strided --home /stride/.strided q bank balances stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided q bank balances stride1ft20pydau82pgesyl9huhhux307s9h3078692y

docker-compose --ansi never exec -T stride2 strided tx bank send val2 stride16vlrvd7lsfqg8q7kyxcyar9v7nt0h99p5arglq 10ustrd --home /stride/.strided --keyring-backend test --chain-id STRIDE -y

docker-compose --ansi never exec -T stride1 strided tx stakeibc liquid-stake 1000 ibc/9117A26BA81E29FA4F78F57DC2BD90CD3D26848101BA880445F119B22A1E254E --keyring-backend test --home /stride/.strided --from val1 --chain-id STRIDE
# ignite scaffold query module-address name:string --response addr --module stakeibc 

# docker-compose --ansi never exec -T stride1 strided tx ibc-transfer transfer channel-1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2: 1000ustrd --home /stride/.strided --keyring-backend test --from val1 --chain-id STRIDE -y

docker-compose --ansi never exec -T stride1 strided --home /stride/.strided --keyring-backend test tx ibc-transfer transfer transfer channel-1 cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2 1000ustrd --from val1 --chain-id STRIDE -y

docker-compose --ansi never exec -T gaia1 gaiad --home /gaia/.gaiad --keyring-backend test tx ibc-transfer transfer transfer channel-1 stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7 1000uatom --from gval1 --chain-id GAIA_1 -y

docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init GAIA_1 STRIDE connection-0 transfer transfer


docker-compose run hermes hermes -c /tmp/hermes.toml tx raw chan-open-init $main_chain $main_gaia_chain connection-0 transfer transfer > /dev/null
