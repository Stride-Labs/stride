package types

type AcknowledgementResponseStatus int

const (
	SUCCESS AcknowledgementResponseStatus = iota
	TIMEOUT
	FAILURE
)

type AcknowledgementResponse struct {
	Status       AcknowledgementResponseStatus
	MsgResponses [][]byte
	Error        string
}
