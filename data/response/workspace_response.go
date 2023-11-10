package response

import (
	"api-todo/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkSpaceResponse struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" `
	Name string             `json:"name"`
	Rol  models.ROL         `json:"rol"`
}

type WorkSpaceResponseDetail struct {
	Id      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" `
	Name    string             `json:"name"`
	Rol     models.ROL         `json:"rol"`
	Columns ColumnResponses    `json:"columns"`
}

type WorkSpaceResponses []WorkSpaceResponse
