package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Status string

const (
	Todo  Status = "todo"
	Doing Status = "doing"
	Done  Status = "done"
)

// Note data type
type Note struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Task      string             `json:"task"`
	Status    Status             `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdateAt  time.Time          `json:"update_at"`
}

type Notes []Note
