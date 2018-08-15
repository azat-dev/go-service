package lib

import (
	"fmt"
	"go/format"
	"strings"
)

func buildTypesFile(service *Service) (string, error) {
	typesFileText := fmt.Sprintf(`
		package %v
		import (
			"encoding/json"
			"github.com/pkg/errors"
			validator "github.com/asaskevich/govalidator"
		)

		type Validatable interface {
			Validate() error
		}

	`, service.Package)

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

	for methodName, methodData := range service.Methods {
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
		return "", fmt.Errorf("can't format code: %v \n\n %v", err, typesFileText)
	}

	return string(formattedText), nil
}

func getGoType(typeInfo TypeInfo) string {
	var resultType string

	if typeInfo.IsVariable {
		return "interface{}"
	}

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

	for fieldName, fieldTypeInfo := range fields {
		fieldName := string(fieldName)
		goType := getGoType(fieldTypeInfo)
		fieldsText = fieldsText + fmt.Sprintf("%v %v `json:\"%v\"`\n", strings.Title(fieldName), goType, fieldName)
	}

	typeValidator, err := getTypeValidator(name, fields)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
		type %v struct {
			%v 
		}
		%v
		
		%v
	`, name, fieldsText, typeValidator, getUnmarshaller(name, fields)), nil
}

func getUnmarshaller(typeName TypeName, fields Fields) string {

	variableFields := map[string]TypeInfo{}
	plainFields := map[string]TypeInfo{}

	for fieldName, fieldTypeInfo := range fields {
		fieldName := strings.Title(string(fieldName))

		if fieldTypeInfo.IsVariable {
			variableFields[fieldName] = fieldTypeInfo
		} else {
			plainFields[fieldName] = fieldTypeInfo
		}
	}

	if len(variableFields) == 0 {
		return ""
	}

	commonFields := ""
	commonFieldsAssignment := ""
	for fieldName, fieldTypeInfo := range plainFields {
		commonFields += fmt.Sprintf("%v %v \n", fieldName, getGoType(fieldTypeInfo))
		commonFieldsAssignment += fmt.Sprintf("v.%v = commonFields.%v \n", fieldName, fieldName)
	}

	rawVariableFields := ""
	variableFieldsUnmarshal := ""

	for fieldName, fieldTypeInfo := range variableFields {
		rawVariableFields += fmt.Sprintf("%v json.RawMessage \n", fieldName)
		mapField := fieldTypeInfo.MapField

		cases := ""
		for value, mappingTypeInfo := range fieldTypeInfo.Mapping {
			cases += fmt.Sprintf(`
				case "%v":
					var parsedData %v
					err = json.Unmarshal(commonFields.%v, &parsedData)
					if err != nil {
						return err
					}

					v.%v = parsedData
					break
			`, value, getGoType(mappingTypeInfo), fieldName, fieldName)
		}

		variableFieldsUnmarshal += fmt.Sprintf(`
			switch(v.%v) {
				%v
				default:
					return errors.New("invalid %v value")
			}
		`, strings.Title(string(mapField)), cases, mapField)
	}

	return fmt.Sprintf(`
		func(v *%v) UnmarshalJSON(packed []byte) error {
			commonFields := struct{
				%v
				%v
			} {}

			err := json.Unmarshal(packed, &commonFields)
			if err != nil {
				return err
			}

			%v

			%v

			return nil
		}
	`, typeName, commonFields, rawVariableFields, commonFieldsAssignment, variableFieldsUnmarshal)
}

func getTypeValidator(name TypeName, fields Fields) (string, error) {
	conditions := ""

	for fieldName, fieldTypeInfo := range fields {
		fieldName := strings.Title(string(fieldName))

		condition := getValidateCondition(fieldName, fieldTypeInfo)
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

	if conditions == "" {
		return fmt.Sprintf(`
			func(v *%v) Validate() error {
				return nil
			}
		`, name), nil
	}

	return fmt.Sprintf(`
		func(v *%v) Validate() error {
			isValid := false
			%v
			return nil
		}
	`, name, conditions), nil
}

func getValidateCondition(fieldName string, typeInfo TypeInfo) string {
	if typeInfo.IsVariable {
		if typeInfo.IsOptional {
			return "true"
		}

		cases := ""
		for _, mapTypeInfo := range typeInfo.Mapping {

			goType := getGoType(mapTypeInfo)

			cases += fmt.Sprintf(`
				case %v:
					x := value.(%v)
					return (&x).Validate() == nil
			`, goType, goType)
		}

		return fmt.Sprintf(`
			func (value interface{}) bool {
				switch(value.(type)) {
					%v
					default:
						panic("not implemented")
				}
			}(v.%v)
		`, cases, fieldName)
	}

	if typeInfo.IsArray {
		result := "true"
		goType := getGoType(typeInfo)

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
			`, result, goType, fieldName)

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

		goType := getGoType(paramType)
		fieldsText = fieldsText + fmt.Sprintf("%v %v `json:\"%v\"`\n", strings.Title(paramName), goType, paramName)
		fields[FieldName(paramName)] = paramType
	}

	paramsValidator, err := getTypeValidator(TypeName(name), fields)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("type %v struct { %v } \n %v", name, fieldsText, paramsValidator), nil
}
