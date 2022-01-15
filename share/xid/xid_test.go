package xid

import (
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	var lastId uint64
	for i := 1; i <= 10000; i++ {
		id := Get()
		if id <= lastId {
			t.Errorf("not sequence \nlast_id(%d) id(%d) \nlast_id(%064b) id(%064b)", lastId, id, lastId, id)
			t.FailNow()
		}
		lastId = id
		if i%1000 == 0 {
			fmt.Printf("-->\n    id(%d)\n    id(%064b)\n", id, id)
		}
	}
}

func TestGetByKey(t *testing.T) {
	var lastId uint64
	for i := 1; i <= 10000; i++ {
		id := GetByKey(123456789)
		if id <= lastId {
			t.Errorf("not sequence \nlast_id(%d) id(%d) \nlast_id(%064b) id(%064b)", lastId, id, lastId, id)
		}
		lastId = id
		if i%1000 == 0 {
			fmt.Printf("-->\n    id(%d)\n    id(%064b)\n", id, id)
		}
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}

func TestNextMillisecond(t *testing.T) {
	t1 := timestamp()
	t2 := nextMillisecond(t1)

	if t2 <= t1 {
		t.Errorf("time was not advanced to next millisecond")
	}
}
