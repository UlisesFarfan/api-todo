package workspace_repository

import (
	"api-todo/data/request"
	"api-todo/data/response"
)

type WorkSpacRepository interface {
	Save(workSpace request.CreateWorkSpaceRequest) error
	Update(workSpace request.UpdateWorkSpaceRequest) (response.WorkSpaceResponse, error)
	Delete(workSpaceId string) error
	FindById(workSpaceId string, user_id string) (response.WorkSpaceResponseDetail, error)
	FindAllByUserId(user_id string) (response.WorkSpaceResponses, error)
	UpdateColumnsOrder(workSpace request.UpdateColumnOrderRequest) (response.WorkSpaceResponse, error)
}
