package model

type Chat struct {
	ChatId int64  `json:"chat_id"`
	UserA  uint32 `json:"user_a"`
	USerB  uint32 `json:"user_b"`
}
