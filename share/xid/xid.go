package xid

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"dk/share/ip"
)

/** 协议说明

参考 snowflake 算法
B为毫秒时间戳
C为同一毫秒内的有序序列
D为机器信息
E为业务信息

|<---------------------- uint64 64 bits ----------------------->|
|---A--|----------B---------|-----C-----|-----D-----|-----E-----|
+------+--------------------+-----------+-----------+-----------+
|  1   |     41 bits        |  8 bits   |  6 bits   |  8 bits   |
| bit  | millisecond (msec) | sequence  | machineId |key(0-255) |
|unused|   	                |  0-255    |eg:ip(0-63)|eg:store_id|
+---------+-----------------+-----------+-----------+-----------+
*/

const (
	timestampOffset = 22
	sequenceOffset  = 14
	machineIdOffset = 8
	maxTimestamp    = 1<<41 - 1
	maxSequence     = 1<<8 - 1
	maxMachineId    = 1<<6 - 1
	maxKey          = 1<<8 - 1
)

var (
	// 为了降低生成的 id 的大小而减去无用的时间戳
	since = time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	//
	defKey uint64
	//
	machineId uint64
	//
	xidDebug = os.Getenv("XID_DEBUG") == "true"
	// protect the generator function
	genMux sync.Mutex
	//
	last = timestamp()
	seq  = uint64(0)
)

type idStruct struct {
	timestamp uint64
	seq       uint64
}

func init() {

	privateIP := ip.PrivateIPToMachineID()
	defKey = uint64(privateIP & maxKey)          // 屏蔽高位
	machineId = uint64(privateIP & maxMachineId) // 屏蔽高位

	if defKey == 0 { // 机器IP没取到时用随机的
		rand.Seed(time.Now().UnixNano())
		defKey = uint64(rand.Uint32() & maxKey)
	}

	if xidDebug {
		fmt.Printf("xid: machineId=%06b(%d) defKey=%08b(%d)\n", machineId, machineId, defKey, defKey)
	}
}

// GetByKey 实在是没有 store id 才考虑该函数
func GetByKey(key uint64) uint64 {
	key = key & maxKey // 屏蔽高位
	t := generator()
	return (t.timestamp << timestampOffset) | (t.seq << sequenceOffset) | (machineId << machineIdOffset) | key
}

// Get 实在是没有 store id 才考虑该函数
func Get() uint64 {
	t := generator()
	return (t.timestamp << timestampOffset) | (t.seq << sequenceOffset) | (machineId << machineIdOffset) | defKey
}

func generator() idStruct {

	genMux.Lock()
	defer genMux.Unlock()

	ts := timestamp()

	if ts != last { // 不同毫秒重置 seq
		seq = 0
		last = ts
	} else if seq == maxSequence { // 相同毫秒但 seq 达到最大值 4096, 此时需要增加毫秒并重置 seq
		ts = nextMillisecond(ts)
		seq = 0
		last = ts
	} else {
		seq++ // 同一毫秒自增 seq
	}

	return idStruct{timestamp: ts, seq: seq}
}

func nextMillisecond(ts uint64) uint64 {
	i := timestamp()
	for ; i <= ts; i = timestamp() {
		<-time.After(time.Millisecond)
	}
	return i
}

func timestamp() uint64 {
	return (uint64(time.Now().UnixNano()-since) / uint64(time.Millisecond)) & maxTimestamp
}
