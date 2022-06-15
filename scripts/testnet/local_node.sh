acct=$(id -un)
strided init $acct --home ~/.stride/.droplet -o --chain-id=droplet 2> /dev/null
cp state/strideTestNode1/config/genesis.json ~/.stride/.droplet/config/genesis.json

sed -i -E "s|seeds = \"\"|seeds = \"\"|g" "~/.stride/.droplet/config/config.toml"