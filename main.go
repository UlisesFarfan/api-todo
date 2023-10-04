package main

import (
	"api-todo/config"
	"api-todo/controller"
	"api-todo/database"
	"api-todo/repositories/column_repository"
	"api-todo/repositories/note_repository"
	"api-todo/repositories/user_repository"
	"api-todo/repositories/workspace_repository"
	"api-todo/router"
	"api-todo/service"
	"context"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/cors"
)

func main() {
	config, _ := config.LoadConfig(".")
	db := database.GetDatabase()
	ctx := context.Background()
	validate := validator.New()

	// init repository
	userRepository := user_repository.NewUsersRepositoryImpl(db, ctx)
	workSpaceRepository := workspace_repository.NewWorkSpaceRepositoryImpl(db, ctx)
	noteRepository := note_repository.NewUsersRepositoryImpl(db, ctx)
	columnRepository := column_repository.NewColumnRepositoryImpl(db, ctx)

	// init service
	authenticationService := service.NewAuthenticationServiceImpl(userRepository, validate)

	// init controller
	authenticationController := controller.NewAuthenticationController(authenticationService)
	userController := controller.NewUsersController(userRepository)
	workSpaceController := controller.NewWorkSpaceController(workSpaceRepository)
	noteController := controller.NewNoteController(noteRepository)
	columnController := controller.NewColumnController(columnRepository)

	// init routes
	routes := router.NewRouter(
		userRepository,
		authenticationController,
		userController,
		workSpaceController,
		noteController,
		columnController,
	)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173/", "http://localhost:5173", "https://todoappbyulises.vercel.app/", "https://todoappbyulises.vercel.app"},
		AllowCredentials: true,
		AllowedMethods:   []string{"POST", "PUT", "PATCH", "GET", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "X-Api-Key", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(routes)

	// server_err := server.ListenAndServe()
	// log.Panic(server_err)
	log.Fatal((http.ListenAndServe(":"+config.ServerPort, handler)))
}
