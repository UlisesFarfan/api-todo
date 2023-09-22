package user_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/models"
)

type UsersRepository interface {
	Save(users request.CreateUsersRequest) error
	Update(users request.UpdateUsersRequest) (response.UsersResponse, error)
	Delete(usersId string) error
	FindById(userId string) (response.UsersResponse, error)
	FindAll() (response.UsersResponses, error)
	FindByEmail(username string) (models.User, error)
}
