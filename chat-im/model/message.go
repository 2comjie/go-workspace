package model

import "time"

type ChatType int32

const (
	ChatTypePrivate ChatType = 0
	ChatTypeGroup   ChatType = 1
)

type MessageType int32

const (
	MessageTypeContent MessageType = 0
	MessageTypeImage   MessageType = 1
)

type Message struct {
	Seq     int64  `json:"seq"`
	ChatId  int64  `json:"chat_id"` // 群聊的时候chatid就是group id
	Content string `json:"content"`
	SlotId  int32  `json:"slot_id"`

	ChatType  ChatType  `json:"chat_type"`
	FromUid   uint32    `json:"from_uid"`
	ToUid     uint32    `json:"to_uid"`
	ToGroupId uint32    `json:"to_group_id"`
	SendTime  time.Time `json:"create_time"`
}
