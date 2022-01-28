package ip

import (
	"testing"
)

func TestPrivateIPToMachineID(t *testing.T) {
	mid := PrivateIPToMachineID()
	if mid <= 0 {
		t.Error("MachineID should be > 0")
	}
	t.Log(mid)
}
