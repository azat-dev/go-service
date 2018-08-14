package lib

type TypeName string
type FieldName string
type FieldType string
type Fields map[FieldName]TypeName

type MethodName string
type ParamName string

type MethodData struct {
	Params map[ParamName]TypeName `json:"params"`
	Result TypeName               `json:"result"`
}

type Service struct {
	Version     string                    `json:"version"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Types       map[TypeName]Fields       `json:"types"`
	Methods     map[MethodName]MethodData `json:"methods"`
}
