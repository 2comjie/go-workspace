package model

import "time"

type GroupRole int32

const (
	GroupRoleMember GroupRole = 0
	GroupRoleOwner  GroupRole = 1
	GroupRoleAdmin  GroupRole = 2
)

type Group struct {
	GroupId    int64     `json:"group_id"`
	Owner      uint32    `json:"owner"`
	Name       string    `json:"name"`
	CreateTime time.Time `json:"create_time"`
	MaxMembers int32     `json:"max_members"`
	SharedKey  string    `json:"shared_key"` // 分片键
}

type GroupMember struct {
	Uid      uint32    `json:"uid"`
	GroupId  int64     `json:"group_id"`
	NickName string    `json:"nick_name"`
	Role     GroupRole `json:"role"`
	JoinTime time.Time `json:"join_time"`
}

type GroupWithMembers struct {
	Group
	Members []*GroupMember `json:"members"`
}
