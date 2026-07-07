package shared

import "encoding/json"

type HttpResponse struct {
	Status  int
	Data    interface{}
	Message string
}

func FailResponse(err error, status int) HttpResponse {
	if status == 0 {
		status = 501
	}
	return HttpResponse{Status: status, Message: err.Error(), Data: nil}
}

func SuccessResponse(data interface{}, status int, message string) HttpResponse {
	if status == 0 {
		status = 501
	}
	return HttpResponse{Status: status, Message: message, Data: data}
}
func (h HttpResponse) Json() string {
	response := map[string]interface{}{
		"data":    h.Data,
		"message": h.Message,
		"status":  h.Status,
	}
	responseb, _ := json.Marshal(response)
	return string(responseb)
}
