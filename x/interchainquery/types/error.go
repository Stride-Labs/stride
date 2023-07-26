package types

import "errors"

var (
	ErrAlreadyFulfilled    = errors.New("query already fulfilled")
	ErrSucceededNoDelete   = errors.New("query succeeded; do not not execute default behavior")
	ErrInvalidICQProof     = errors.New("icq query response failed")
	ErrICQCallbackNotFound = errors.New("icq callback id not found")
	ErrInvalidICQRequest   = errors.New("invalid interchain query request")
	ErrFailedToRetryQuery  = errors.New("failed to retry query")
)
