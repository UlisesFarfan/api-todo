package response

import (
	"time"
)

type UsersResponse struct {
	Id        string    `json:"_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Img       string    `json:"img"`
	Rol       string    `json:"rol"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

type UsersResponses []UsersResponse

type LoginResponse struct {
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
}
