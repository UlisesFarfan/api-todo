package request

import "time"

type CreateWorkSpaceRequest struct {
	UserId    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}

type UpdateWorkSpaceRequest struct {
	Id     string `json:"_id"`
	Name   string `json:"name"`
	UserId string `json:"user_id"`
}
