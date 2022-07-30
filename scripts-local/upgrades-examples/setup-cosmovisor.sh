
# PATH="/Users/rileyedmunds/stride/scripts-local/state/stride"
# STRIDED_PATH="/Users/rileyedmunds/stride/build/strided"

# mkdir -p $PATH/cosmovisor/upgrades
# mkdir -p $PATH/cosmovisor/genesis/bin/
# cp $STRIDED_PATH $PATH/cosmovisor/genesis/bin/

mkdir -p /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/upgrades
mkdir -p /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/genesis/bin/
cp /Users/rileyedmunds/stride/build/strided /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/genesis/bin/


# mkdir -p /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/current
# echo '{"name":"test1","height":100}' > /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/current/upgrade-info.json


mkdir -p /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/upgrades/test1
echo '{"name":"test1","height":100}' > /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/upgrades/test1/upgrade-info.json

mkdir -p /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/upgrades/test1/bin/
cp ~/Downloads/strided /Users/rileyedmunds/stride/scripts-local/state/stride/cosmovisor/upgrades/test1/bin


source ~/.profile

cosmovisor version

