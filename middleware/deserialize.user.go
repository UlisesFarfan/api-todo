package middleware

import (
	"api-todo/config"
	"api-todo/repositories/user_repository"
	"api-todo/security"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DeserializeUser(userRepository user_repository.UsersRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var token string
		authorizationHeader := ctx.Request.Header.Get("Authorization")
		fields := strings.Fields(authorizationHeader)

		if len(fields) != 0 && fields[0] == "Bearer" {
			token = fields[1]
		}

		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "You are not logged in"})
			return
		}

		config, _ := config.LoadConfig()
		sub, err := security.ValidateToken(token, config.TokenSecret)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		id := fmt.Sprint(sub)
		result, err := userRepository.FindById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "the user belonging to this token no logger exists"})
			return
		}

		ctx.Set("currentUser", result.Name)
		ctx.Set("currentUserEmail", result.Email)
		ctx.Set("currentUserRol", result.Rol)
		ctx.Set("currentUserId", result.Id)
		ctx.Next()

	}
}
