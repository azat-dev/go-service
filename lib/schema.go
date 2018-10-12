package lib

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"strconv"
	"strings"
)

type TypeName string
type FieldName string
type StructTypeData map[FieldName]TypeInfo
type TypeInfo struct {
	IsCustomType bool
	DataType     string
	IsArray      bool
	IsOptional   bool
	Min          int
	Max          int
	IsVariable   bool
	MapField     FieldName
	Mapping      map[string]TypeInfo
}

type EnumTypeData struct {
	Type          string
	ValuesString  map[string]string
	ValuesInteger map[string]int
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

type MethodName string
type ParamName string

type Parameter struct {
	Name     ParamName
	TypeInfo TypeInfo
}

type MethodData struct {
	Params []Parameter `json:"params"`
	Result TypeInfo    `json:"result"`
}

func (f *MethodData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parsedData := struct {
		Params yaml.MapSlice `json:"params"`
		Result TypeName      `json:"result"`
	}{}

	err := unmarshal(&parsedData)
	if err != nil {
		return err
	}

	*f = MethodData{
		Params: []Parameter{},
	}
	f.Result = getTypeInfo(string(parsedData.Result))

	for _, data := range parsedData.Params {
		f.Params = append(f.Params, Parameter{
			Name:     ParamName(data.Key.(string)),
			TypeInfo: getTypeInfo(string(data.Value.(string))),
		})
	}

	return err
}

type TypeMapping struct {
	MapField string              `json:"mapField"`
	Mapping  map[string]TypeName `json:"mapping"`
}

type TypesData map[TypeName]interface{}

func (t *TypesData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parsedTypes := map[TypeName]interface{}{}

	err := unmarshal(&parsedTypes)
	if err != nil {
		return err
	}

	*t = TypesData{}

	for typeName, value := range parsedTypes {

		cleanedTypeName := strings.Title(strings.TrimSuffix(string(typeName), "(enum)"))

		var err error
		var resultData interface{}
		parsedDataType := value.(map[interface{}]interface{})

		if strings.HasSuffix(string(typeName), "(enum)") {
			var enumData *EnumTypeData
			enumData, err = unmarshalEnum(parsedDataType)
			resultData = *enumData

		} else {
			var structData *StructTypeData
			structData, err = unmarshalStructData(parsedDataType)
			resultData = *structData
		}

		if err != nil {
			return err
		}

		(*t)[TypeName(cleanedTypeName)] = resultData
	}

	return err
}

func unmarshalEnum(value map[interface{}]interface{}) (*EnumTypeData, error) {

	enumType := value["type"].(string)

	valuesInteger := map[string]int{}
	valuesString := map[string]string{}

	switch enumType {
	case "string":
		for key, value := range value["values"].(map[interface{}]interface{}) {
			key := key.(string)
			valuesString[key] = value.(string)
		}
	case "int":
		for key, value := range value["values"].(map[interface{}]interface{}) {
			key := key.(string)
			valuesInteger[key] = value.(int)
		}
	}

	return &EnumTypeData{
		Type:          enumType,
		ValuesString:  valuesString,
		ValuesInteger: valuesInteger,
	}, nil
}

func unmarshalStructData(parsedData map[interface{}]interface{}) (*StructTypeData, error) {

	result := StructTypeData{}
	for fieldName, value := range parsedData {

		fieldName := fieldName.(string)
		if strings.HasSuffix(fieldName, "?") {

			fieldName := strings.TrimSuffix(fieldName, "?")

			value := value.(map[interface{}]interface{})
			mapField := value["mapField"].(string)

			mapping := map[string]TypeInfo{}

			for key, value := range value["mapping"].(map[interface{}]interface{}) {
				mapping[key.(string)] = getTypeInfo(value.(string))
			}

			result[FieldName(fieldName)] = TypeInfo{
				MapField:   FieldName(mapField),
				Mapping:    mapping,
				IsVariable: true,
			}

		} else {

			value := value.(string)
			result[FieldName(fieldName)] = getTypeInfo(value)
		}
	}

	return &result, nil
}

type Service struct {
	Version     string                    `json:"version"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Types       TypesData                 `json:"types"`
	Methods     map[MethodName]MethodData `json:"methods"`
	Package     string                    `json:"package"`
}
