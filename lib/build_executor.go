package lib

import (
	"fmt"
	"go/format"
	"strings"
)

func buildExecutorFile(service *Service) (string, error) {
	cases := ""
	for method, methodData := range service.Methods {
		cases += buildExecutorCase(method, methodData) + "\n"
	}

	text := fmt.Sprintf(`
		//!!!GENERATED BY "GO-SERVICE" DON'T CHANGE THIS FILE!!!
		package %v
		
		import (
			"encoding/json"
			"fmt"
			"github.com/pkg/errors"

			"github.com/akaumov/go-service/exchange"
		)

		type Executor struct {
			handler HandlerInterface
		}

		func NewExecutor(handler HandlerInterface) *Executor {
			return &Executor{
				handler: handler,
			}
		}
		
		func (e *Executor) Execute(packedMessage *[]byte) (*[]byte, error) {
			if packedMessage == nil {
				return nil, errors.New("message text is required")
			}

			response := e.execute(packedMessage)
			packed, _ := json.Marshal(response)
			return &packed, nil
		}

		func (e *Executor) execute(packedMessage *[]byte) exchange.ResponseMessage {
			
			var requestMessage exchange.RequestMessage
			err := json.Unmarshal(*packedMessage, &requestMessage)
			if err != nil {
				return exchange.NewErrorResponse("", "WrongRequest", "can't parse message")
			}

			requestId := requestMessage.Id

			switch(requestMessage.Method) {
				%v
			}

			return exchange.NewErrorResponse("", "WrongRequest", "no such method")
		}
	`, service.Name, cases)

	formattedText, err := format.Source([]byte(text))
	if err != nil {
		return "", err
	}

	return string(formattedText), nil
}

func buildExecutorCase(methodName MethodName, methodData MethodData) string {
	resultType := methodData.Result
	paramsName := strings.Title(string(methodName)) + "Params"

	returnTypeInfo := getTypeInfo(string(resultType))
	resultTypeGo := schemaTypeToGoType(string(resultType))
	if returnTypeInfo.IsCustomType || returnTypeInfo.IsArray {
		resultTypeGo = "*" + resultTypeGo
	}

	params := ""
	for paramName, paramType := range methodData.Params {
		if params != "" {
			params += ", "
		}

		param := "params." + strings.Title(string(paramName))

		paramTypeInfo := getTypeInfo(string(paramType))
		if paramTypeInfo.IsCustomType || paramTypeInfo.IsArray {
			param = "&" + param
		}

		params += fmt.Sprintf("%v", param)
	}

	handlerMethod := strings.Title(string(methodName))

	return fmt.Sprintf(`
		case "%v":
			var params %v
			err := json.Unmarshal(requestMessage.Params, &params)
			if err != nil {
				return exchange.NewErrorResponse(requestId, "WrongRequest", fmt.Sprintf("can't parse params: %%v", err))
			}

			err = params.Validate() 
			if err != nil {
				return exchange.NewErrorResponse(requestId, "WrongRequest", fmt.Sprintf("can't wrong params: %%v", err))
			}

			result, err := e.handler.%v(%v)
			if err != nil {
				return exchange.NewErrorResponse(requestId, "ServerError", err.Error())
			}
			
			return exchange.NewResultResponse(requestId, result)

	`, methodName, paramsName, handlerMethod, params)
}
