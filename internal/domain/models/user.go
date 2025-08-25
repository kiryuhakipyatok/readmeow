package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id uuid.UUID `json:"id"`
	Credentials
	Avatar         string    `json:"avatar"`
	TimeOfRegister time.Time `json:"time_of_register"`
	NumOfTemplates uint32    `json:"num_of_templates"`
	NumOfReadmes   uint32    `json:"num_of_readmes"`
}

type Credentials struct {
	Nickname string `json:"nickname"`
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password []byte `json:"-"`
}
