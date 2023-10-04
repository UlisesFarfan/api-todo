package response

import (
	"api-todo/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkSpaceResponse struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" `
	Name      string             `json:"name"`
	Users     []models.UserRef   `json:"users"`
	Columns   ColumnResponses    `json:"columns"`
	CreatedAt time.Time          `json:"created_at"`
	UpdateAt  time.Time          `json:"update_at"`
}

type WorkSpaceResponses []WorkSpaceResponse
