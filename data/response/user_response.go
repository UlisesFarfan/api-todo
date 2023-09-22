package response

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsersResponse struct {
	Id         string               `json:"_id"`
	Name       string               `json:"name"`
	Email      string               `json:"email"`
	Rol        string               `json:"rol"`
	WorkSpaces []primitive.ObjectID `json:"work_spaces"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdateAt   time.Time            `json:"update_at"`
}

type UsersResponses []UsersResponse

type LoginResponse struct {
	TokenType string        `json:"token_type"`
	Token     string        `json:"token"`
	User      UsersResponse `json:"user"`
}
