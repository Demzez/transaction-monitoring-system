package dto

import (
	"time"
	"transaction-monitoring-system/protoStruct"
)

type UserDTO struct {
	UserID    int64
	Login     string
	RoleId    int64
	CreatedAt time.Time
}

func (u *UserDTO) DTOToProto() *protoStruct.RespUser {
	if u == nil {
		return nil
	}

	return &protoStruct.RespUser{
		UserId:    u.UserID,
		Login:     u.Login,
		RoleId:    u.RoleId,
		CreatedAt: u.CreatedAt.Unix(),
	}
}
