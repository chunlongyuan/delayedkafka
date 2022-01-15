package xid

import (
	"testing"
)

func TestPrivateIPToMachineID(t *testing.T) {
	mid := privateIPToMachineID()
	if mid <= 0 {
		t.Error("MachineID should be > 0")
	}
	t.Log(mid)
}
