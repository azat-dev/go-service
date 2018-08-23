package lib

import (
	"fmt"
	"go/format"
	"strings"
	"github.com/pkg/errors"
)

func buildTypesFile(service *Service) (string, error) {
	typesFileText := fmt.Sprintf(`
		package %v
		import (
			"fmt"
			"encoding/json"
			"github.com/pkg/errors"
			validator "github.com/asaskevich/govalidator"
		)

		type Validatable interface {
			Validate() error
		}

	`, service.Package)

	for name, typeData := range service.Types {

		var err error
		var typeText string

		switch typeData.(type) {
		case StructTypeData:
			typeText, err = buildStructType(name, typeData.(StructTypeData))

		case EnumTypeData:
			typeText, err = buildEnumType(name, typeData.(EnumTypeData))
		}

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

func buildStructType(name TypeName, data StructTypeData) (string, error) {
	fieldsText := ""

	for fieldName, fieldTypeInfo := range data {
		fieldName := string(fieldName)
		goType := getGoType(fieldTypeInfo)
		fieldsText = fieldsText + fmt.Sprintf("%v %v `json:\"%v\"`\n", strings.Title(fieldName), goType, fieldName)
	}

	typeValidator, err := getTypeValidator(name, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
		type %v struct {
			%v 
		}
		%v
		
		%v
	`, name, fieldsText, typeValidator, getUnmarshaller(name, data)), nil
}

func buildEnumType(name TypeName, data EnumTypeData) (string, error) {

	typeName := strings.Title(string(name))
	valuesText := ""

	if data.Type == "int" {
		for valueName, value := range data.ValuesInteger {
			valuesText = valuesText + fmt.Sprintf("%v %v = %v\n", strings.Title(valueName), typeName, value)
		}
	} else if data.Type == "string" {
		for valueName, value := range data.ValuesString {
			valuesText = valuesText + fmt.Sprintf("%v %v = \"%v\"\n", strings.Title(strings.Replace(valueName, "\"", "\"", -1)), typeName, value)
		}
	} else {
		return "", errors.New("wrong enum type")
	}


	return fmt.Sprintf(`
		type %v %v 

		const (
			%v 
		)
	`, typeName, data.Type, valuesText), nil
}

func getUnmarshaller(typeName TypeName, fields StructTypeData) string {

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

func getTypeValidator(name TypeName, fields StructTypeData) (string, error) {
	conditions := ""

	for fieldName, fieldTypeInfo := range fields {
		fieldName := strings.Title(string(fieldName))

		condition := getValidateCondition("value", fieldTypeInfo)
		if condition == "true" {
			continue
		}

		conditions += fmt.Sprintf(`
			 {
				value := v.%v
				isValid := %v

				if !isValid {
					return fmt.Errorf("%v is invalid")
				}
			}
		`, fieldName, condition, fieldName)
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
			%v
			return nil
		}
	`, name, conditions), nil
}

func getValidateCondition(valueName string, typeInfo TypeInfo) string {
	if typeInfo.IsVariable {
		return getValidateConditionForVariableValue(valueName, typeInfo)
	}

	if typeInfo.IsArray {
		return getValidateConditionForArrayValue(valueName, typeInfo)
	}

	if typeInfo.IsCustomType {
		return getValidateConditionForCustomType(valueName, typeInfo)
	}

	return getValidateConditionForSimpleValue(valueName, typeInfo)
}

func getValidateConditionForCustomType(valueName string, typeInfo TypeInfo) string {
	if typeInfo.IsOptional {
		return fmt.Sprintf("%v == nil || (%v.Validate() == nil)", valueName, valueName)
	}

	return fmt.Sprintf("%v.Validate() == nil", valueName)
}

func getValidateConditionForVariableValue(valueName string, typeInfo TypeInfo) string {
	if typeInfo.IsOptional {
		//TODO
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

	return fmt.Sprintf(
		`func (value interface{}) bool {
				switch(value.(type)) {
					%v
					default:
						panic("not implemented")
				}
			}(%v)
		`, cases, valueName)
}

func getValidateConditionForSimpleValue(valueName string, typeInfo TypeInfo) string {

	nilCheck := fmt.Sprintf("%v == nil || ", valueName)

	switch typeInfo.DataType {
	case "uuid":
		if typeInfo.IsOptional {
			return fmt.Sprintf("%v validator.IsUUID(*%v)", nilCheck, valueName)
		}
		return fmt.Sprintf("validator.IsUUID(%v)", valueName)

	case "email":
		if typeInfo.IsOptional {
			return fmt.Sprintf("%v validator.IsEmail(*%v)", nilCheck, valueName)
		}
		return fmt.Sprintf("validator.IsEmail(%v)", valueName)

	case "string":
		if typeInfo.IsOptional {
			return fmt.Sprintf(
				"%v (%v)",
				nilCheck, getLengthCondition("*"+valueName, typeInfo.Min, typeInfo.Max))
		}

		return getLengthCondition(valueName, typeInfo.Min, typeInfo.Max)
	}

	return "true"
}

func getValidateConditionForArrayValue(valueName string, typeInfo TypeInfo) string {

	goType := getGoType(typeInfo)

	lengthCondition := getLengthCondition(valueName, typeInfo.Min, typeInfo.Max)
	if lengthCondition == "true" {
		lengthCondition = ""
	} else {
		lengthCondition += " &&\n"
	}

	itemCondition := "item.Validate() == nil"
	if !typeInfo.IsCustomType {
		itemCondition = getValidateConditionForSimpleValue("item", typeInfo)
	}

	value := "&" + valueName
	if typeInfo.IsOptional {
		value = valueName
	}

	itemsCondition := fmt.Sprintf(
		`func (value *%v) bool {
				for _, item := range *value {
					isValid := %v

					if !isValid {
						return false
					}
				}
				return true
			} (%v)
	`, goType, itemCondition, value)

	if (typeInfo.IsOptional) {
		return fmt.Sprintf("%v == nil || (%v %v)", valueName, lengthCondition, itemsCondition)
	}

	return lengthCondition + itemsCondition
}

func getLengthCondition(valueName string, min int, max int) string {
	if max > 0 {
		return fmt.Sprintf(
			"(len(%v) >= %v && len(%v) <= %v)",
			valueName, min, valueName, max,
		)

	} else if min > 0 {

		return fmt.Sprintf(
			"(len(%v) >= %v)",
			valueName, min,
		)
	}

	return "true"
}

func buildParamsForMethod(methodName MethodName, methodData MethodData) (string, error) {
	name := strings.Title(string(methodName)) + "Params"
	fieldsText := ""

	fields := StructTypeData{}
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
