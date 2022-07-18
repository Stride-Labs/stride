package lighttest

import (
	"testing"
)

func TestMakeConnection(t *testing.T) {
	t, coord, _ := InitialBasicSetup(t)

	if coord == nil {
		t.Error("coord is nil")
	}

}
