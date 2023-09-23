package router

import (
	"api-todo/controller"
	"api-todo/middleware"
	"api-todo/repositories/user_repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	userRepository user_repository.UsersRepository,
	authenticationController *controller.AuthenticationController,
	usersController *controller.UserController,
	workSpaceController *controller.WorkSpaceController,
	noteController *controller.NoteController,
) *gin.Engine {
	service := gin.Default()

	service.GET("", func(context *gin.Context) {
		context.JSON(http.StatusOK, "welcome home")
	})

	service.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	// url/api
	router := service.Group("/api")

	// url/api/authentication/
	authenticationRouter := router.Group("/authentication")
	authenticationRouter.POST("/register", authenticationController.Register)
	authenticationRouter.POST("/login", authenticationController.Login)
	authenticationRouter.GET("", middleware.DeserializeUser(userRepository), authenticationController.GetUserToken)

	// url/api/users/
	usersRouter := router.Group("/users")
	usersRouter.GET("", middleware.DeserializeUser(userRepository), usersController.GetUsers)
	usersRouter.PATCH("/update", usersController.UpdateUser)

	// url/api/work_space/
	workSpaceRouter := router.Group("/work_space")
	workSpaceRouter.POST("", workSpaceController.PostWorkSpace)
	workSpaceRouter.PATCH("", workSpaceController.UpdateWorkSpace)
	workSpaceRouter.GET("/:workspace_id", workSpaceController.FindWorkSpaceById)
	workSpaceRouter.DELETE("/:workspace_id", workSpaceController.DeleteWorkSpaceById)

	// url/api/work_space/
	noteRouter := router.Group("/note")
	noteRouter.POST("/:workspace_id", noteController.PostNote)
	noteRouter.PATCH("", noteController.UpdateNote)
	noteRouter.GET("/:note_id", noteController.FindNoteById)
	noteRouter.DELETE("/:note_id", noteController.DeleteNoteById)

	return service
}
