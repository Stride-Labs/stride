package types

import (
	"encoding/binary"
)

var _ binary.ByteOrder

const (
	// CallbackDataKeyPrefix is the prefix to retrieve all CallbackData
	CallbackDataKeyPrefix = "CallbackData/value/"
)

// CallbackDataKey returns the store key to retrieve a CallbackData from the index fields
func CallbackDataKey(
	callbackKey string,
) []byte {
	var key []byte

	callbackKeyBytes := []byte(callbackKey)
	key = append(key, callbackKeyBytes...)
	key = append(key, []byte("/")...)

	return key
}
