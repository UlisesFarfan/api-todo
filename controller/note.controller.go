package controller

import (
	"api-todo/data/request"
	"api-todo/data/response"
	"api-todo/repositories/note_repository"

	"net/http"

	"github.com/gin-gonic/gin"
)

type NoteController struct {
	noteRepository note_repository.NoteRepository
}

func NewNoteController(repository note_repository.NoteRepository) *NoteController {
	return &NoteController{noteRepository: repository}
}

func (controller *NoteController) PostNote(c *gin.Context) {
	createNoteRequest := request.CreateNoteRequest{}
	workspaceId := c.Param("workspace_id")
	err := c.BindJSON(&createNoteRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	err = controller.noteRepository.Save(createNoteRequest, workspaceId)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Create note successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *NoteController) UpdateNote(c *gin.Context) {
	updateNoteRequest := request.UpdateNoteRequest{}
	err := c.BindJSON(&updateNoteRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	update_note, err := controller.noteRepository.Update(updateNoteRequest)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}

	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Create note successfully", update_note)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *NoteController) FindNoteById(c *gin.Context) {
	note_id := c.Param("note_id")
	note_response, err := controller.noteRepository.FindById(note_id)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Successfully fetch by id", note_response)
	c.IndentedJSON(http.StatusOK, webResponse)
}

func (controller *NoteController) DeleteNoteById(c *gin.Context) {
	note_id := c.Param("note_id")
	err := controller.noteRepository.Delete(note_id)
	if err != nil {
		webResponse := response.CreateResponse(http.StatusBadRequest, "Bad Request", err.Error(), nil)
		c.IndentedJSON(http.StatusBadRequest, webResponse)
		return
	}
	webResponse := response.CreateResponse(http.StatusOK, "Ok", "Deleted note by id successfully", nil)
	c.IndentedJSON(http.StatusOK, webResponse)
}
