SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [ -z "$1" ]
then
    echo "Error, you must pass testnet name in. E.g. \"sh setup_testnet_state.sh droplet\""
    exit 1
fi

echo "Setting up chain $1"

GAIA_FOLDER="${STATE}/gaia"
GAIA_CMD="docker run -v ${STATE}/gaia:/gaia/.gaiad gcr.io/stride-nodes/testnet:tub_gaia gaiad --home /gaia/.gaiad"

GAIA_TOKENS=500000000uatom
GAIA_STAKE_TOKENS=300000000uatom
GAIA_ENDPOINT=gaia.$STRIDE_CHAIN.stridelabs.co
GAIA_CHAIN="GAIA_${STRIDE_CHAIN}"

HERMES_CMD="docker run -v ${STATE}/hermes.toml:/tmp/hermes.toml -v ${STATE}:/hermes/.hermes/keys gcr.io/stride-nodes/testnet:tub_hermes hermes -c /tmp/hermes.toml"
ICQ_CMD="docker run -v ${STATE}/icq:/hermes/.hermes/keys gcr.io/stride-nodes/testnet:tub_icq icq"

GETKEY() {
  grep -i -A 10 "\- name: $1" "$STATE/keys.txt" | tail -n 1
}

GETRLY2() {
  cat internal/state/keys.txt | tail -1
}

ICQ_STRIDE_KEY="helmet say goat special plug umbrella finger night flip axis resource tuna trigger angry shove essay point laundry horror eager forget depend siren alarm"
ICQ_GAIA_KEY="capable later bamboo snow drive afraid cheese practice latin brush hand true visa drama mystery bird client nature jealous guess tank marriage volume fantasy"

ICQ_ADDRESS_STRIDE="stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"
ICQ_ADDRESS_GAIA="cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"