package controller

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/repositories/user_repository"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userRepository user_repository.UsersRepository
}

func NewUsersController(repository user_repository.UsersRepository) *UserController {
	return &UserController{userRepository: repository}
}

// Get all users
func (controller *UserController) GetUsers(c *gin.Context) {
	users, err := controller.userRepository.FindAll()
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusAccepted, "Ok", "Successfully fetch all user data", users)
	c.IndentedJSON(http.StatusAccepted, webResponse)
}

// Update user
func (controller *UserController) UpdateUser(c *gin.Context) {
	var user request.UpdateUsersRequest
	err := c.BindJSON(&user)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	update_user, err := controller.userRepository.Update(user)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusAccepted, "Ok", "User update successfully", update_user)
	c.IndentedJSON(http.StatusAccepted, webResponse)
}
