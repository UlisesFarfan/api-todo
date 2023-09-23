package service

import (
	"api-todo/data/request"
	"api-todo/data/response"
)

type AuthenticationService interface {
	Login(users request.LoginRequest) (response.UsersResponse, string, error)
	Register(users request.CreateUsersRequest) error
	GetUserToken(user_id string) (response.UsersResponse, error)
}
