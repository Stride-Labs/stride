#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../../config.sh

AIRDROP_NAME="sttia"

# --------------------------------------------------------------------------------

echo -e "\n>>> Query Airdrop Allocation to stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv | head -n 30
echo -e "\n>>> Query Balances for stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q bank balances stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv
sleep 3


echo -e "\n>>> Claiming daily..."
$STRIDE_MAIN_CMD tx airdrop claim-daily $AIRDROP_NAME \
 --from val3 -y --gas 1000000 | TRIM_TX
sleep 5

echo -e "\n>>> Query Airdrop Allocation to stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv | head -n 30
echo -e "\n>>> Query Balances for stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q bank balances stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv
sleep 3

# --------------------------------------------------------------------------------


echo -e "\n>>> Query Airdrop Allocation to stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv | head -n 30
echo -e "\n>>> Query Balances for stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q bank balances stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv
sleep 3


echo -e "\n>>> Claiming early..."
$STRIDE_MAIN_CMD tx airdrop claim-early $AIRDROP_NAME \
 --from val3 -y --gas 1000000 | TRIM_TX
sleep 5

echo -e "\n>>> Query Airdrop Allocation to stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q airdrop user-allocation $AIRDROP_NAME stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv | head -n 30
echo -e "\n>>> Query Balances for stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv..."
$STRIDE_MAIN_CMD q bank balances stride1nnurja9zt97huqvsfuartetyjx63tc5zq8s6fv
sleep 3

