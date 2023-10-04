package response

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ColumnResponse struct {
	Id        primitive.ObjectID `json:"_id"`
	Name      string             `json:"name"`
	Notes     NoteResponses      `json:"notes"`
	CreatedAt time.Time          `json:"created_at"`
	UpdateAt  time.Time          `json:"update_at"`
}

type ColumnResponses []ColumnResponse
