package response

type response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func CreateResponse(
	code int,
	status string,
	message string,
	data interface{},
) response {
	new_response := response{
		Code:    code,
		Status:  status,
		Message: message,
		Data:    data,
	}
	return new_response
}
