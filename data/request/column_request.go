package request

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateColumnRequest struct {
	WorkSpaceId string               `json:"workspace_id"`
	Name        string               `json:"name"`
	Notes       []primitive.ObjectID `json:"notes"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdateAt    time.Time            `json:"update_at"`
}

type UpdateColumnRequest struct {
	Id   string `json:"_id"`
	Name string `json:"name"`
}

type UpdateNotesOrderRequest struct {
	Id    string   `json:"_id"`
	Notes []string `json:"notes"`
}
