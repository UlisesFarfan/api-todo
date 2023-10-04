package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Note data type
type Column struct {
	Id        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string               `json:"name"`
	Notes     []primitive.ObjectID `json:"notes"`
	CreatedAt time.Time            `json:"created_at"`
	UpdateAt  time.Time            `json:"update_at"`
}

type Columns []Column
