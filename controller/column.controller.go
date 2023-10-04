package controller

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/repositories/column_repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ColumnController struct {
	columnRepository column_repository.ColumnRepository
}

func NewColumnController(repository column_repository.ColumnRepository) *ColumnController {
	return &ColumnController{columnRepository: repository}
}

func (controller *ColumnController) PostColumn(c *gin.Context) {
	createColumnRequest := request.CreateColumnRequest{}
	err := c.BindJSON(&createColumnRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	res, err := controller.columnRepository.Save(createColumnRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Create column successfully", res)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *ColumnController) UpdateColumn(c *gin.Context) {
	createColumnRequest := request.UpdateColumnRequest{}
	err := c.BindJSON(&createColumnRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	result, err := controller.columnRepository.Update(createColumnRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Create column successfully", result)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *ColumnController) FindColumnById(c *gin.Context) {
	columnId := c.Param("column_id")
	result, err := controller.columnRepository.FindById(columnId)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Successfully fetch by id", result)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *ColumnController) DeleteColumnById(c *gin.Context) {
	columnId := c.Param("column_id")
	err := controller.columnRepository.Delete(columnId)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Delete by id successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *ColumnController) UpdateNotesOrder(c *gin.Context) {
	update_notes_order := request.UpdateNotesOrderRequest{}
	bindErr := c.BindJSON(&update_notes_order)
	if bindErr != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", bindErr.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	_, err := controller.columnRepository.UpdateNotesOrder(update_notes_order)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Update notes order successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}
