package note_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
)

type NoteRepository interface {
	Save(note request.CreateNoteRequest, workspaceId string) error
	Update(note request.UpdateNoteRequest) (response.NoteResponse, error)
	Delete(noteId string) error
	FindById(noteId string) (response.NoteResponse, error)
	FindAll() (response.NoteResponses, error)
}
