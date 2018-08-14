package lib

import (
	"fmt"
	"strings"
	"go/format"
	"regexp"
	"strconv"
)

func buildTypesFile(service *Service) (string, error) {
	typesFileText := fmt.Sprintf(`
		package %v
		import (
			"github.com/pkg/errors"
			validator "github.com/asaskevich/govalidator"
		)
	`, service.Name)

	for name, fields := range service.Types {
		typeText, err := buildType(name, fields)
		if err != nil {
			return "", err
		}

		typesFileText += fmt.Sprintf(`
			/////////////////////////////////////////////////////////////////////
			//%v
			%v
		`, name, typeText)
	}

	typesFileText += `
		/////////////////////////////////////////////////////////////////////
		//PARAMETERS
	`

	for methodName, methodData  := range service.Methods {
		paramsText, err := buildParamsForMethod(methodName, methodData)
		if err != nil {
			return "", err
		}

		typesFileText += fmt.Sprintf(`
			/////////////////////////////////////////////////////////////////////
			//%v
			%v
		`, methodName, paramsText)
	}

	formattedText, err := format.Source([]byte(typesFileText))
	if err != nil {
		return "", err
	}

	return string(formattedText), nil
}

type TypeInfo struct {
	IsCustomType bool
	DataType     string
	IsArray      bool
	IsOptional   bool
	Min          int
	Max          int
}

func getTypeInfo(schemaType string) TypeInfo {
	result := TypeInfo{
		Max: -1,
		Min: 0,
	}

	var r = regexp.MustCompile(`^(?P<array>\[\])?(?P<type>[\w]+)(\((?P<min>[0-9]+)\s*(,\s*(?P<max>[0-9]+))?\))?(?P<optional>[?])?$`)
	matches := r.FindAllStringSubmatch(schemaType, -1)
	groups := r.SubexpNames()

	var value string
	for index, group := range groups {

		value = matches[0][index]

		switch group {
		case "array":
			result.IsArray = (value != "")
		case "optional":
			result.IsOptional = (value != "")
		case "type":
			result.DataType = value
		case "min":
			if value != "" {
				result.Min, _ = strconv.Atoi(value)
				result.Max, _ = strconv.Atoi(value)
			}
		case "max":
			if value != "" {
				result.Max, _ = strconv.Atoi(value)
			}
		}
	}

	result.IsCustomType = strings.Title(result.DataType) == result.DataType

	return result
}

func schemaTypeToGoType(schemaType string) string {
	typeInfo := getTypeInfo(schemaType)

	var resultType string

	switch typeInfo.DataType {
	case "uuid":
		resultType = "string"
	case "int":
		resultType = "int"
	case "int64":
		resultType = "int64"
	case "boolean":
		resultType = "bool"
	case "time":
		resultType = "int64"
	case "string":
		resultType = "string"
	case "email":
		resultType = "string"

	default:
		resultType = strings.Title(typeInfo.DataType)
	}

	if typeInfo.IsArray {
		resultType = "[]" + resultType
	}

	if typeInfo.IsOptional {
		resultType = "*" + resultType
	}

	return resultType
}

func buildType(name TypeName, fields Fields) (string, error) {
	fieldsText := ""

	for fieldName, fieldType := range fields {
		fieldName := string(fieldName)
		goType := schemaTypeToGoType(string(fieldType))
		fieldsText = fieldsText + fmt.Sprintf("%v %v `json:\"%v\"`\n", strings.Title(fieldName), goType, fieldName)
	}

	typeValidator, err := getTypeValidator(name, fields)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("type %v struct { %v }\n %v", name, fieldsText, typeValidator), nil
}

func getTypeValidator(name TypeName, fields Fields) (string, error) {
	conditions := ""

	for fieldName, fieldType := range fields {
		fieldName := strings.Title(string(fieldName))

		condition := getValidateCondition(fieldName, fieldType)
		if condition == "true" {
			continue
		}

		conditions += fmt.Sprintf(`
			isValid = %v

			if !isValid {
				return errors.New("%v is invalid")
			}
		`, condition, fieldName)
	}

	return fmt.Sprintf(`
		func(v *%v) Validate() error {
			isValid := false
			%v
			return nil
		}
	`, name, conditions), nil
}

func getValidateCondition(fieldName string, fieldType TypeName) string {
	typeInfo := getTypeInfo(string(fieldType))

	if typeInfo.IsArray {
		result := "true"
		goType := schemaTypeToGoType(string(fieldType))

		if typeInfo.Max != -1 {
			result = fmt.Sprintf(`
				func (value %v) bool {
					length := len(value)
					return length >= %v && length <= %v
				}(v.%v)
			`, goType, typeInfo.Min, typeInfo.Max, fieldName)

		} else if typeInfo.Min > 0 {

			result = fmt.Sprintf(`
				%v &&
				func (value %v) bool {
					length := len(value)
					return length >= %v
				}(v.%v)
			`, result, goType, fieldName, typeInfo.Min)
		}

		if typeInfo.IsCustomType {
			result = fmt.Sprintf(`
				%v &&
				func (value %v) bool {
					for _, item := range value {
						if item.Validate() != nil {
							return false
						}
					}
					return true
				}(v.%v)
			`, result, goType, fieldName, )

		}

		if typeInfo.IsOptional {
			return fmt.Sprintf(`
				v.%v == nil || (%v)
			`, fieldName, strings.TrimSpace(result))
		}

		return result
	}

	nilCheck := fmt.Sprintf("v.%v == nil || ", fieldName)
	if typeInfo.IsCustomType {
		return fmt.Sprintf("%v v.%v.Validate() == nil", nilCheck, fieldName)
	}

	if typeInfo.IsArray {
		panic("NOT IMPLEMENTED")
	} else {
		switch typeInfo.DataType {
		case "uuid":
			if typeInfo.IsOptional {
				return fmt.Sprintf("%v validator.IsUUID(*v.%v)", nilCheck, fieldName)
			}
			return fmt.Sprintf("validator.IsUUID(v.%v)", fieldName)

		case "email":
			if typeInfo.IsOptional {
				return fmt.Sprintf("%v validator.IsEmail(*v.%v)", nilCheck, fieldName)
			}
			return fmt.Sprintf("validator.IsEmail(v.%v)", fieldName)
		}
	}

	return "true"
}

func buildParamsForMethod(methodName MethodName, methodData MethodData) (string, error) {

	name := strings.Title(string(methodName)) + "Params"
	fieldsText := ""

	fields := Fields{}
	for paramName, paramType := range methodData.Params {
		paramName := string(paramName)

		goType := schemaTypeToGoType(string(paramType))
		fieldsText = fieldsText + fmt.Sprintf("%v %v `json:\"%v\"`\n", strings.Title(paramName), goType, paramName)
		fields[FieldName(paramName)] = paramType
	}

	paramsValidator, err := getTypeValidator(TypeName(name), fields)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("type %v struct { %v } \n %v", name, fieldsText, paramsValidator), nil
}