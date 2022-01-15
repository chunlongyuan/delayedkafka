package messenger

import (
	"time"
)

// 消息

type Message struct {
	DelayMs     int64  `json:"delay_ms"`      // 延迟多少毫秒
	Body        string `json:"body"`          // 元数据
	CreatedAtMs int64  `json:"created_at_ms"` // 毫秒时间戳
}

func (m Message) NoDelay() bool {
	return !time.Unix(0, m.CreatedAtMs*1e6).Add(time.Duration(m.DelayMs) * time.Millisecond).After(time.Now().Add(time.Second))
}
