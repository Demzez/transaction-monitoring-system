package dto

import "time"

type UserDTO struct {
	Login     string
	Password  string
	CreatedAt time.Time
}

func (t *UserDTO) DtoToProto() string {
	return "" // TODO: убрать затычку
}
