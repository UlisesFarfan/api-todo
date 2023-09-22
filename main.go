package main

import (
	"api-todo/config"
	"api-todo/controller"
	"api-todo/database"
	"api-todo/repositories/note_repository"
	"api-todo/repositories/user_repository"
	"api-todo/repositories/workspace_repository"
	"api-todo/router"
	"api-todo/service"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	// lead env
	loadConfig, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	db := database.GetDatabase()
	ctx := context.Background()
	validate := validator.New()

	// init repository
	userRepository := user_repository.NewUsersRepositoryImpl(db, ctx)
	workSpaceRepository := workspace_repository.NewWorkSpaceRepositoryImpl(db, ctx)
	noteRepository := note_repository.NewUsersRepositoryImpl(db, ctx)

	// init service
	authenticationService := service.NewAuthenticationServiceImpl(userRepository, validate)

	// init controller
	authenticationController := controller.NewAuthenticationController(authenticationService)
	userController := controller.NewUsersController(userRepository)
	workSpaceController := controller.NewWorkSpaceController(workSpaceRepository)
	noteController := controller.NewNoteController(noteRepository)

	// init routes
	routes := router.NewRouter(
		userRepository,
		authenticationController,
		userController,
		workSpaceController,
		noteController,
	)

	// add cors
	routes.Use(CORSMiddleware())
	server := &http.Server{
		Addr:           ":" + loadConfig.ServerPort,
		Handler:        routes,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server_err := server.ListenAndServe()
	log.Panic(server_err)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
