package workspace_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
)

type WorkSpacRepository interface {
	Save(workSpace request.CreateWorkSpaceRequest) error
	Update(workSpace request.UpdateWorkSpaceRequest) (response.WorkSpaceResponse, error)
	Delete(workSpaceId string) error
	FindById(workSpaceId string) (response.WorkSpaceResponse, error)
	FindAll() (response.WorkSpaceResponses, error)
}
