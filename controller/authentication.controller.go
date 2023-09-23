package controller

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthenticationController struct {
	authenticationService service.AuthenticationService
}

func NewAuthenticationController(service service.AuthenticationService) *AuthenticationController {
	return &AuthenticationController{authenticationService: service}
}

// Login func
func (controller *AuthenticationController) Login(ctx *gin.Context) {
	loginRequest := request.LoginRequest{}
	err := ctx.ShouldBindJSON(&loginRequest)
	user, token, err_token := controller.authenticationService.Login(loginRequest)
	fmt.Println(err_token)
	if err_token != nil {
		fmt.Println(err, err_token)
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", "Invalid username or password", nil)
		ctx.JSON(http.StatusBadRequest, webResponse)
		return
	}
	resp := response.LoginResponse{
		TokenType: "Bearer",
		Token:     token,
		User:      user,
	}
	webResponse := response.CreateResponse(200, "Ok", "Successfully logged in", resp)
	ctx.JSON(http.StatusOK, webResponse)
}

// Register func
func (controller *AuthenticationController) Register(ctx *gin.Context) {
	createUsersRequest := request.CreateUsersRequest{}
	err := ctx.BindJSON(&createUsersRequest)
	if err != nil {
		webResponse := response.CreateResponse(400, "Bad Request", err.Error(), nil)
		ctx.JSON(http.StatusOK, webResponse)
		return
	}
	err = controller.authenticationService.Register(createUsersRequest)
	if err != nil {
		webResponse := response.CreateResponse(400, "Bad Request", err.Error(), nil)
		ctx.JSON(http.StatusOK, webResponse)
		return
	}
	webResponse := response.CreateResponse(200, "Ok", "Successfully registered", nil)
	ctx.JSON(http.StatusOK, webResponse)
}

// Get user by token
func (controller *AuthenticationController) GetUserToken(ctx *gin.Context) {
	user_id := ctx.GetString("currentUserId")
	user_response, err := controller.authenticationService.GetUserToken(user_id)
	if err != nil {
		webResponse := response.CreateResponse(400, "Bad Request", err.Error(), nil)
		ctx.JSON(http.StatusOK, webResponse)
		return
	}
	webResponse := response.CreateResponse(200, "Ok", "Successfully registered", user_response)
	ctx.JSON(http.StatusOK, webResponse)
}
