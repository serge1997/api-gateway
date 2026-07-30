package httpresponse

import "encoding/json"

type HttpResponse struct {
	Status  int
	Data    interface{}
	Message string
	Success bool
}

func FailResponse(err error, status int) HttpResponse {
	if status == 0 {
		status = 501
	}
	return HttpResponse{Status: status, Message: err.Error(), Data: nil, Success: false}
}

func SuccessResponse(data interface{}, status int, message string) HttpResponse {
	if status == 0 {
		status = 501
	}
	return HttpResponse{Status: status, Message: message, Data: data, Success: true}
}
func (h HttpResponse) Json() string {
	response := map[string]interface{}{
		"data":    h.Data,
		"message": h.Message,
		"status":  h.Status,
		"success": h.Success,
	}
	responseb, _ := json.Marshal(response)
	return string(responseb)
}

func ToJSON(message string, status int, success bool, data interface{}) string {
	response := map[string]interface{}{
		"data":    data,
		"message": message,
		"status":  status,
		"success": success,
	}
	responseb, _ := json.Marshal(response)
	return string(responseb)
}
