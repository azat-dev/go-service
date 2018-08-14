package exchange

import (
	"encoding/json"
)

type RequestMessage struct {
	Id     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type ResponseMessage struct {
	Id     string           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  *json.RawMessage `json:"error"`
}

type ErrorResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func NewErrorResponse(requestId string, name string, message string) ResponseMessage {
	packed, _ := json.Marshal(ErrorResponse{
		Name:    name,
		Message: message,
	})

	rawPacked := json.RawMessage(packed)

	return ResponseMessage{
		Id:     requestId,
		Result: nil,
		Error:  &rawPacked,
	}
}

func NewResultResponse(requestId string, result interface{}) ResponseMessage {

	packed, _ := json.Marshal(result)
	rawPacked := json.RawMessage(packed)

	return ResponseMessage{
		Id:     requestId,
		Result: &rawPacked,
		Error:  nil,
	}
}
