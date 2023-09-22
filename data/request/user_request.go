package request

import "time"

type CreateUsersRequest struct {
	Name      string    `validate:"required,min=2,max=100" json:"name"`
	Email     string    `validate:"required,min=2,max=100" json:"email"`
	Password  string    `validate:"required,min=2,max=100" json:"password"`
	Rol       string    `json:"rol"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

type UpdateUsersRequest struct {
	Id    string `json:"_id"`
	Name  string `json:"name"`
	Rol   string `json:"rol"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `validate:"required" json:"email"`
	Password string `validate:"required,min=2,max=100" json:"password"`
}
