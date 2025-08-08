package models

import (
	"github.com/google/uuid"
)

type User struct {
	Id             uuid.UUID
	Login          string
	Email          string
	Avatar         string
	Password       []byte `json:"-"`
	TimeOfRegister int64
	NumOfTemplates uint16
	NumOfReadmes   uint16
}
