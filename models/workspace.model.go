package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RolWorkSpace string

const (
	AdminWorkSpace  RolWorkSpace = "admin"
	NormalWorkSpace RolWorkSpace = "normal"
)

type UserRef struct {
	UserId primitive.ObjectID `json:"user_id"`
	Rol    RolWorkSpace       `json:"role"`
}

// WorkSpace data type
type WorkSpace struct {
	Id        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" `
	Name      string               `json:"name"`
	Users     []UserRef            `json:"users"`
	Columns   []primitive.ObjectID `json:"columns"`
	CreatedAt time.Time            `json:"created_at"`
	UpdateAt  time.Time            `json:"update_at"`
}

type WorkSpaces []WorkSpace
