package types

type AckResponseStatus int

const (
	AckResponseStatus_SUCCESS AckResponseStatus = iota
	AckResponseStatus_TIMEOUT
	AckResponseStatus_FAILURE
)

type AcknowledgementResponse struct {
	Status       AckResponseStatus
	MsgResponses [][]byte
	Error        string
}
