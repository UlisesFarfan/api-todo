package response

import (
	"api-todo/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NoteResponse struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" `
	Task      string             `json:"task"`
	Status    models.Status      `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdateAt  time.Time          `json:"update_at"`
}

type NoteResponses []NoteResponse
