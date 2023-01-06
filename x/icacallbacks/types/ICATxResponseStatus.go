package types

import sdk "github.com/cosmos/cosmos-sdk/types"


type ICATxResponseStatus int
const (
    SUCCESS ICATxResponseStatus = iota
    TIMEOUT
    FAILURE
)
type ICATxResponse struct {
	Status ICATxResponseStatus // enum of SUCCESS, TIMEOUT, FAILURE
	Data sdk.TxMsgData // TxMsgData
}