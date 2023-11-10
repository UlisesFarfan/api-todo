package controller

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/repositories/workspace_repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WorkSpaceController struct {
	workSpaceRepository workspace_repository.WorkSpacRepository
}

func NewWorkSpaceController(repository workspace_repository.WorkSpacRepository) *WorkSpaceController {
	return &WorkSpaceController{workSpaceRepository: repository}
}

// Create work_space
func (controller *WorkSpaceController) PostWorkSpace(c *gin.Context) {
	CreateWorkSpaceRequest := request.CreateWorkSpaceRequest{}
	err := c.BindJSON(&CreateWorkSpaceRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	err = controller.workSpaceRepository.Save(CreateWorkSpaceRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusCreated, "Ok", "WorkSpace created successfully", nil)
	c.IndentedJSON(http.StatusCreated, webResponse)
}

// Update wor_space
func (controller *WorkSpaceController) UpdateWorkSpace(c *gin.Context) {
	UpdateWorkSpaceRequest := request.UpdateWorkSpaceRequest{}
	err := c.BindJSON(&UpdateWorkSpaceRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	update_workspace, err := controller.workSpaceRepository.Update(UpdateWorkSpaceRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "WorkSpace update successfully", update_workspace)
	c.IndentedJSON(http.StatusOK, webResponse)
}

// Find by _id
func (controller *WorkSpaceController) FindWorkSpaceById(c *gin.Context) {
	workspace_id := c.Param("workspace_id")
	user_id := c.GetString("currentUserId")
	workspace_response, err := controller.workSpaceRepository.FindById(workspace_id, user_id)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Successfully fetch by id data", workspace_response)
	c.IndentedJSON(http.StatusOK, webResponse)
}

// Delete WorkSpace
func (controller *WorkSpaceController) DeleteWorkSpaceById(c *gin.Context) {
	workspace_id := c.Param("workspace_id")
	err := controller.workSpaceRepository.Delete(workspace_id)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "WorkSpace deleted successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}

// FindAllByUserId WorkSpace
func (controller *WorkSpaceController) FindAllByUserId(c *gin.Context) {
	user_id := c.GetString("currentUserId")
	workspace_response, err := controller.workSpaceRepository.FindAllByUserId(user_id)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Successfully fetch all workspaces data", workspace_response)
	c.IndentedJSON(http.StatusOK, webResponse)
}

// Update columns order
func (controller *WorkSpaceController) UpdateColumnsOrder(c *gin.Context) {
	update_columns_order := request.UpdateColumnOrderRequest{}
	bindErr := c.BindJSON(&update_columns_order)
	if bindErr != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", bindErr.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	_, err := controller.workSpaceRepository.UpdateColumnsOrder(update_columns_order)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Update columns order successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}
