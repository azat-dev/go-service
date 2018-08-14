package lib

import (
	"go/format"
	"fmt"
	"strings"
)

func buildHandlerInterfaceFile(service *Service) (string, error) {
	methods := ""
	for method, methodData := range service.Methods {
		methods += buildHandlerMethod(method, methodData) + "\n"
	}

	text := fmt.Sprintf(`
		//!!!GENERATED BY "GO-SERVICE" DON'T CHANGE THIS FILE!!!
		package %v

		type HandlerInterface interface {
			%v
		}
	`, service.Name, methods)

	formattedText, err := format.Source([]byte(text))
	if err != nil {
		return "", err
	}

	return string(formattedText), nil
}

func buildHandlerMethod(methodName MethodName, methodData MethodData) string {
	resultType := methodData.Result

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

		paramTypeInfo := getTypeInfo(string(paramType))
		paramTypeGo := schemaTypeToGoType(string(paramType))
		if paramTypeInfo.IsCustomType || paramTypeInfo.IsArray {
			paramTypeGo = "*" + paramTypeGo
		}

		params += fmt.Sprintf("%v %v", paramName, paramTypeGo)
	}

	name := strings.Title(string(methodName))
	return fmt.Sprintf("%v(%v) (%v, error)", name, params, resultTypeGo)
}