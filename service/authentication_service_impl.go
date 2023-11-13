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
func (a *AuthenticationServiceImpl) Login(users request.LoginRequest) (string, error) {
	// Find username in database
	new_users, users_err := a.UsersRepository.FindByEmail(users.Email)
	if users_err != nil {
		return "", errors.New("invalid username or Password")
	}

	config, _ := config.LoadConfig(".")

	verify_error := security.VerifyPassword(new_users.Password, users.Password)
	if verify_error != nil {
		fmt.Println(verify_error)
		return "", errors.New("invalid username or Password")
	}

	// Generate Token
	token, err_token := security.GenerateToken(config.TokenExpiresIn, new_users.Id, config.TokenSecret)
	if err_token != nil {
		return "", err_token
	}

	return token, nil

}

// Register implements AuthenticationService
func (a *AuthenticationServiceImpl) Register(users request.CreateUsersRequest) error {
	hashedPassword, _ := security.HashPassword(users.Password)
	newUser := request.CreateUsersRequest{
		Name:     users.Name,
		Email:    users.Email,
		Img:      users.Img,
		Password: hashedPassword,
	}
	err := a.UsersRepository.Save(newUser)
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
