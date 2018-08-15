package lib

import (
	"regexp"
	"strconv"
	"strings"
)

type TypeName string
type FieldName string
type Fields map[FieldName]TypeInfo
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

func (f *Fields) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var parsedData map[string]interface{}

	err := unmarshal(&parsedData)
	if err != nil {
		return err
	}

	*f = Fields{}

	for fieldName, value := range parsedData {
		if !strings.HasSuffix(fieldName, "?") {

			value := value.(string)
			(*f)[FieldName(fieldName)] = getTypeInfo(value)

		} else {

			fieldName := strings.TrimSuffix(fieldName, "?")

			value := value.(map[interface{}]interface{})
			mapField := value["mapField"].(string)

			mapping := map[string]TypeInfo{}

			for key, value := range value["mapping"].(map[interface{}]interface{}) {
				mapping[key.(string)] = getTypeInfo(value.(string))
			}

			(*f)[FieldName(fieldName)] = TypeInfo{
				MapField:   FieldName(mapField),
				Mapping:    mapping,
				IsVariable: true,
			}
		}
	}

	return err
}

type MethodName string
type ParamName string

type MethodData struct {
	Params map[ParamName]TypeInfo `json:"params"`
	Result TypeInfo               `json:"result"`
}

func (f *MethodData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	parsedData := struct {
		Params map[ParamName]TypeName `json:"params"`
		Result TypeName               `json:"result"`
	}{}

	err := unmarshal(&parsedData)
	if err != nil {
		return err
	}

	*f = MethodData{
		Params: map[ParamName]TypeInfo{},
	}
	f.Result = getTypeInfo(string(parsedData.Result))

	for key, value := range parsedData.Params {
		f.Params[key] = getTypeInfo(string(value))
	}

	return err
}

type TypeMapping struct {
	MapField string              `json:"mapField"`
	Mapping  map[string]TypeName `json:"mapping"`
}

type Service struct {
	Version     string                    `json:"version"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Types       map[TypeName]Fields       `json:"types"`
	Methods     map[MethodName]MethodData `json:"methods"`
	Package     string                    `json:"package"`
}
