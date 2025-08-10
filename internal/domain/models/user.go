package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id             uuid.UUID
	Login          string
	Email          string
	Avatar         string
	Password       []byte `json:"-"`
	TimeOfRegister time.Time
	NumOfTemplates uint16
	NumOfReadmes   uint16
}
