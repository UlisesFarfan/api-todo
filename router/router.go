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
	columnController *controller.ColumnController,
) *gin.Engine {
	service := gin.Default()

	service.GET("/", func(context *gin.Context) {
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
	authenticationRouter.GET("/", middleware.DeserializeUser(userRepository), authenticationController.GetUserToken)

	// url/api/users/
	usersRouter := router.Group("/users")
	usersRouter.GET("/", middleware.DeserializeUser(userRepository), usersController.GetUsers)
	usersRouter.PATCH("/update", middleware.DeserializeUser(userRepository), usersController.UpdateUser)

	// url/api/work_space/
	workSpaceRouter := router.Group("/work_space")
	workSpaceRouter.POST("/", middleware.DeserializeUser(userRepository), workSpaceController.PostWorkSpace)
	workSpaceRouter.PATCH("/", middleware.DeserializeUser(userRepository), workSpaceController.UpdateWorkSpace)
	workSpaceRouter.PUT("/", middleware.DeserializeUser(userRepository), workSpaceController.UpdateColumnsOrder)
	workSpaceRouter.GET("/", middleware.DeserializeUser(userRepository), workSpaceController.FindAllByUserId)
	workSpaceRouter.GET("/:workspace_id", middleware.DeserializeUser(userRepository), workSpaceController.FindWorkSpaceById)
	workSpaceRouter.DELETE("/:workspace_id", middleware.DeserializeUser(userRepository), workSpaceController.DeleteWorkSpaceById)

	// url/api/column
	columnRouter := router.Group("/column")
	columnRouter.POST("/", middleware.DeserializeUser(userRepository), columnController.PostColumn)
	columnRouter.PATCH("/", middleware.DeserializeUser(userRepository), columnController.UpdateColumn)
	columnRouter.PUT("/", middleware.DeserializeUser(userRepository), columnController.UpdateNotesOrder)
	columnRouter.GET("/:column_id", middleware.DeserializeUser(userRepository), columnController.FindColumnById)
	columnRouter.DELETE("/:column_id", middleware.DeserializeUser(userRepository), columnController.DeleteColumnById)

	// url/api/note/
	noteRouter := router.Group("/note")
	noteRouter.POST("/", middleware.DeserializeUser(userRepository), noteController.PostNote)
	noteRouter.PATCH("/", middleware.DeserializeUser(userRepository), noteController.UpdateNote)
	noteRouter.GET("/:note_id", middleware.DeserializeUser(userRepository), noteController.FindNoteById)
	noteRouter.DELETE("/:note_id", middleware.DeserializeUser(userRepository), noteController.DeleteNoteById)

	return service
}
