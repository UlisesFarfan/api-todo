package service

import (
	"api-todo/config"
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/repositories/user_repository"
	"api-todo/security"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type AuthenticationServiceImpl struct {
	UsersRepository user_repository.UsersRepository
	Validate        *validator.Validate
}

func NewAuthenticationServiceImpl(usersRepository user_repository.UsersRepository, validate *validator.Validate) AuthenticationService {
	return &AuthenticationServiceImpl{
		UsersRepository: usersRepository,
		Validate:        validate,
	}
}

// Login implements AuthenticationService
func (a *AuthenticationServiceImpl) Login(users request.LoginRequest) (response.UsersResponse, string, error) {
	// Find username in database
	new_users, users_err := a.UsersRepository.FindByEmail(users.Email)
	if users_err != nil {
		fmt.Println("ERROOOOR", users_err)
		return response.UsersResponse{}, "", errors.New("invalid username or Password")
	}

	config, _ := config.LoadConfig(".")

	verify_error := security.VerifyPassword(new_users.Password, users.Password)
	if verify_error != nil {
		fmt.Println(verify_error)
		return response.UsersResponse{}, "", errors.New("invalid username or Password")
	}

	// Generate Token
	token, err_token := security.GenerateToken(config.TokenExpiresIn, new_users.Id, config.TokenSecret)
	if err_token != nil {
		return response.UsersResponse{}, "", err_token
	}
	user_response := response.UsersResponse{
		Id:         new_users.Id.Hex(),
		Name:       new_users.Name,
		Email:      new_users.Email,
		Rol:        string(new_users.Rol),
		WorkSpaces: new_users.WorkSpaces,
		CreatedAt:  new_users.CreatedAt,
		UpdateAt:   new_users.UpdateAt,
	}
	return user_response, token, nil

}

// Register implements AuthenticationService
func (a *AuthenticationServiceImpl) Register(users request.CreateUsersRequest) error {

	hashedPassword, err := security.HashPassword(users.Password)

	newUser := request.CreateUsersRequest{
		Name:     users.Name,
		Email:    users.Email,
		Password: hashedPassword,
	}
	err = a.UsersRepository.Save(newUser)
	if err != nil {
		return err
	}
	return nil
}

// GetUserToken implements AuthenticationService.
func (a *AuthenticationServiceImpl) GetUserToken(user_id string) (response.UsersResponse, error) {
	user_response, err := a.UsersRepository.FindById(user_id)
	return user_response, err
}
