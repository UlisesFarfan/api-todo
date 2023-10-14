package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Note data type
type Note struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Task      string             `json:"task"`
	CreatedAt time.Time          `json:"created_at"`
	UpdateAt  time.Time          `json:"update_at"`
}

type Notes []Note
