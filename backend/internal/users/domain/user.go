package domain

import (
	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID `db:"id"`
	Name      *string   `db:"name"`
	Bio       *string   `db:"bio"`
	AvatarURL *string   `db:"avatar_url"`
}

// UserInfo является вспомогательным типом для передачи базовой информации о пользователе в другие сервисы через gRPC
type UserInfo struct {
	Id   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

func (u User) GetInfo() UserInfo {
	name := ""
	if u.Name != nil {
		name = *u.Name
	}
	return UserInfo{
		Id:   u.Id,
		Name: name,
	}
}
