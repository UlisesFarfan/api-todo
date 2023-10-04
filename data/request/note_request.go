package request

import (
	"time"
)

type CreateNoteRequest struct {
	ColumnId  string    `json:"column_id"`
	Task      string    `json:"task"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

type UpdateNoteRequest struct {
	Id   string `json:"_id"`
	Task string `json:"task"`
}
