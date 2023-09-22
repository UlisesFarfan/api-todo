package request

import (
	"api-todo/models"
	"time"
)

type CreateNoteRequest struct {
	Task      string        `json:"task"`
	Status    models.Status `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdateAt  time.Time     `json:"update_at"`
}

type UpdateNoteRequest struct {
	Id     string        `json:"_id"`
	Task   string        `json:"task"`
	Status models.Status `json:"status"`
}
