package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id             uuid.UUID `json:"id"`
	Login          string    `json:"login"`
	Email          string    `json:"email"`
	Avatar         string    `json:"avatar"`
	Password       []byte    `json:"-"`
	TimeOfRegister time.Time `json:"time_of_register"`
	NumOfTemplates uint16    `json:"num_of_templates"`
	NumOfReadmes   uint16    `json:"num_of_readmes"`
}
