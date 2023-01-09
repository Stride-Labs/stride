#!/bin/bash

### IBC TRANSFER
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

GETBAL() {
    head -n 1 | grep -o -E '[0-9]+' || "0"
}

acct="cosmos10uxaa5gkxpeungu2c9qswx035v6t3r24w6v2r6dxd858rq2mzknqj8ru28"

# ##################################
# ### WHICH CASE ARE WE STUDYING ###
# ##################################

# TRANSFER CASE (20211218_8PM)
IBC_TRANSFER_AMT=170670514
STAKE_AMT=2040984630
REINVEST_AMT=65246817
START=13307049 # 2022-12-18 19:59:55
END=13307069 # 2022-12-18 20:02:03
# after running this case, tighten the bounds to make it run faster
START=13307050 # 2022-12-18 19:59:55
END=13307060 # 2022-12-18 20:02:03


###############################
### TRANSFER & REINVESTMENT ###
###############################

echo "\n\n\nSearching blocks $START to $END for $acct"
echo "We expect: \n \
\t + $IBC_TRANSFER_AMT (ibc-transfer) \n \
\t + $REINVEST_AMT (reinvest) \n \
\t - $STAKE_AMT (staked) \n \
\t and (maybe) some undelegations, amt unknown \n"

# for i in {$S..$E}
echo "\nWe see:"
for (( c=$START; c<=$END; c++ ))
do
    # echo $c
    pre_block_bal=$($GAIA_MAIN_CMD q bank balances $acct \
    --height "$c" --node https://gaia-fleet.main.stridenet.co:443 \
    --denom "uatom" | GETBAL)

    post_block_bal=$($GAIA_MAIN_CMD q bank balances $acct \
    --height "$((c+1))" --node https://gaia-fleet.main.stridenet.co:443 \
    --denom "uatom" | GETBAL)

    # echo "block $c: dif=$(($post_block_bal-$pre_block_bal)) ($pre_block_bal => $post_block_bal)"

    # if balance is not equal, print the difference
    if [ "$pre_block_bal" -ne "$post_block_bal" ]; then
        echo "\tdiff: $(($post_block_bal - $pre_block_bal)) (block $c)"
    fi
done


###############
### STAKING ###
###############

# make the bounds tighter in this case so it runs faster
START=13307051 # 2022-12-18 19:59:55
END=13307051 # 2022-12-18 20:02:03



echo "\n\n\nSearching blocks $START to $END for $acct"
echo "We expect: \n \
\t + $IBC_TRANSFER_AMT (ibc-transfer) \n \
\t + $REINVEST_AMT (reinvest) \n \
\t - $STAKE_AMT (staked) \n \
\t and (maybe) some undelegations, amt unknown \n"

# from the logs, line 77-108, obtain:
declare -a vals_at_start=("cosmosvaloper130mdu9a0etmeuw52qfxk73pn0ga6gawkxsrlwf" "cosmosvaloper1we6knm8qartmmh2r0qfpsz6pq0s7emv3e0meuw" "cosmosvaloper16fnz0v4cnv5dpnj0p3gaft2q2kzx8z5hfrx6v5" "cosmosvaloper1ehkfl7palwrh6w2hhr2yfrgrq8jetgucudztfe" "cosmosvaloper1y0us8xvsvfvqkk9c6nt5cfyu5au5tww2ztve7q" "cosmosvaloper1vvwtk805lxehwle9l4yudmq6mn0g32px9xtkhc" "cosmosvaloper1n229vhepft6wnkt5tjpwmxdmcnfz55jv3vp77d" "cosmosvaloper1x88j7vp2xnw3zec8ur3g4waxycyz7m0mahdv3p" "cosmosvaloper1sd4tl9aljmmezzudugs7zlaya7pg2895ws8tfs" "cosmosvaloper1kgddca7qj96z0qcxr2c45z73cfl0c75p7f3s2e" "cosmosvaloper1rpgtz9pskr5geavkjz02caqmeep7cwwpv73axj" "cosmosvaloper10e4vsut6suau8tk9m6dnrm0slgd6npe3jx5xpv" "cosmosvaloper1grgelyng2v6v3t8z87wu3sxgt9m5s03xfytvz7" "cosmosvaloper1eh5mwu044gd5ntkkc2xgfg8247mgc56fz4sdg3" "cosmosvaloper14kn0kk33szpwus9nh8n87fjel8djx0y070ymmj" "cosmosvaloper1ma02nlc7lchu7caufyrrqt4r6v2mpsj90y9wzd" "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d" "cosmosvaloper1zqgheeawp7cmqk27dgyctd80rd8ryhqs6la9wc" "cosmosvaloper1g48268mu5vfp4wk7dk89r0wdrakm9p5xk0q50k" "cosmosvaloper16k579jk6yt2cwmqx9dz5xvq9fug2tekvlu9qdv" "cosmosvaloper1qwl879nx9t6kef4supyazayf7vjhennyh568ys" "cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qxerarrl" "cosmosvaloper132juzk0gdmwuxvx4phug7m3ymyatxlh9734g4w" "cosmosvaloper15urq2dtp9qce4fyc85m6upwm9xul3049e02707" "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn" "cosmosvaloper1vf44d85es37hwl9f4h9gv0e064m0lla60j9luj" "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3" "cosmosvaloper1ey69r37gfxvxg62sh4r0ktpuc46pzjrm873ae8" "cosmosvaloper1v5y0tg0jllvxf5c3afml8s3awue0ymju89frut" "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c" "cosmosvaloper14lultfckehtszvzw4ehu0apvsr77afvyju5zzy" "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0")
declare -a val_del_msgs=("22411551" "23591107" "23984292" "24180884" "24377477" "24770662" "26736588" "27129773" "27522958" "28702513" "29882068" "30668439" "33027549" "37116675" "37352586" "42070807" "44429918" "44823103" "45216288" "46395843" "47182214" "52686805" "73132431" "76277912" "80209764" "86893911" "118348720" "121494201" "128571533" "191087967" "203276706" "217431385")
# validated the valset above at START using this query: $STRIDE_MAIN_CMD q stakeibc show-host-zone "cosmoshub-4" --height "$START" --node https://stride-fleet.main.stridenet.co:443 

TOTAL_STAKE_ADDED_ACTUAL=0
echo "\nWe see:"
# iter the validators
for (( val_idx=0; val_idx<=$((${#val_del_msgs[@]}-1)); val_idx++ ))
do
    echo "\texpected diff ${val_del_msgs[val_idx]} for val $val_idx (${vals_at_start[val_idx]})"
    #iter the blocks
    for (( c=$START; c<=$END; c++ ))
    do

        # pre_block_bal=$($GAIA_MAIN_CMD q satking delegations $acct \
        # --height "$c" --node https://gaia-fleet.main.stridenet.co:443 \
        # --denom "uatom" | GETBAL)
        pre_block_stake=$($GAIA_MAIN_CMD q staking delegation "$acct" "${vals_at_start[val_idx]}" \
                                --height "$c" \
                                --node https://gaia-fleet.main.stridenet.co:443 | grep amount | awk '{print $2}' | tr -d '"')

        post_block_stake=$($GAIA_MAIN_CMD q staking delegation "$acct" "${vals_at_start[val_idx]}" \
                                --height "$((c+1))" \
                                --node https://gaia-fleet.main.stridenet.co:443 | grep amount | awk '{print $2}' | tr -d '"')

        # echo "block $c: dif=$(($post_block_bal-$pre_block_bal)) ($pre_block_bal => $post_block_bal)"

        # if balance is not equal, print the difference
        if [ "$pre_block_stake" -ne "$post_block_stake" ]; then
            diff_val_stake=$(($post_block_stake - $pre_block_stake))
            TOTAL_STAKE_ADDED_ACTUAL=$(($TOTAL_STAKE_ADDED_ACTUAL + $diff_val_stake))
            # if the diff does not equal the diff in the array, print the diff
            if [ "$diff_val_stake" -ne "${val_del_msgs[val_idx]}" ]; then
                echo "\t\tactual diff: $(($post_block_stake - $pre_block_stake)) (block $c)"
            fi
        fi
    done
done
echo "\nTOTAL_STAKE_ADDED_ACTUAL: $TOTAL_STAKE_ADDED_ACTUAL"