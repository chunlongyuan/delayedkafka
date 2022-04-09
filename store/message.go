package store

import (
	"time"
)

const TableMessage = `delay_message`

const (
	StatusDelay  = int8(0)
	StatusDelete = int8(1)
	StatusSpent  = int8(2)
)

// 消息

type Message struct {
	ID          uint64    `json:"id" gorm:"primaryKey;auto_increment:false;column:id;not null;comment:'数据库的唯一id'"`   // id
	Topic       string    `json:"topic" gorm:"type:varchar(100);column:topic;not null;default:'';comment:'消息主题'"`    // kafka 接收用的 topic
	DelayMs     int64     `json:"delay_ms" gorm:"column:delay_ms;not null;default:0;comment:'延迟毫秒数'"`                // 延迟多少毫秒
	TTRms       int64     `json:"-" gorm:"column:ttr_ms;not null;default:0;comment:'触发时刻的毫秒时间戳'"`                    // 消费时刻的毫秒时间戳
	State       int8      `json:"-" gorm:"column:state;not null;default:0;comment:'消息状态: 0:待消费 1:已删除 2:已消费'"`        // 消费时刻的毫秒时间戳
	Body        string    `json:"body" gorm:"type:varchar(10240);column:body;not null;default:'';comment:'消息内容'"`    // 元数据
	CreatedAtMs int64     `json:"created_at_ms" gorm:"column:created_at_ms;not null;default:0;comment:'消息创建的毫秒时间戳'"` // 创建毫秒时间戳
	CreatedAt   time.Time `json:"-" gorm:"column:created_at;type:datetime;not null;default:current_timestamp;comment:'创建时间'"`
	UpdatedAt   time.Time `json:"-" gorm:"column:updated_at;type:datetime;not null;default:current_timestamp on update current_timestamp;comment:'更新时间'"`
}

func (*Message) TableName() string {
	return TableMessage
}
