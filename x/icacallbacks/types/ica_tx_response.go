package types

type ICATxResponseStatus int

const (
	SUCCESS ICATxResponseStatus = iota
	TIMEOUT
	FAILURE
)

type ICATxResponse struct {
	Status       ICATxResponseStatus
	MsgResponses [][]byte
	Error        string
}
