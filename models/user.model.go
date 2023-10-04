package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rol string

const (
	AdminRol Rol = "admin"
	UserRol  Rol = "user"
)

// User data type
type User struct {
	Id         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" `
	Name       string               `json:"name"`
	Email      string               `json:"email"`
	Img        string               `json:"img"`
	WorkSpaces []primitive.ObjectID `json:"work_spaces"`
	Rol        Rol                  `json:"rol"`
	Password   string               `json:"password"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdateAt   time.Time            `json:"update_at"`
}

type Users []User
