package column_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
)

type ColumnRepository interface {
	Save(column request.CreateColumnRequest) (response.ColumnResponse, error)
	Update(column request.UpdateColumnRequest) (response.ColumnResponse, error)
	Delete(columnId string) error
	FindById(columnId string) (response.ColumnResponse, error)
	FindAll(workspaceId string) (response.ColumnResponses, error)
	UpdateNotesOrder(column request.UpdateNotesOrderRequest) (response.ColumnResponse, error)
}
