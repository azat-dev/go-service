package db

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Column struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsNullable   bool   `json:"isNullable"`
	DefaultValue string `json:"defaultValue"`
}

type RemoteColumnName string

type Relation struct {
	Type           RelationType `json:"type"`
	Name           string       `json:"name"`
	RemoteTable    string       `json:"remoteTable"`
	ColumnsMapping []ColumnsMap `json:"columnsMap"`
}

type UniqueConstraint struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

type Table struct {
	Name              string             `json:"name"`
	Columns           []Column           `json:"columns"`
	PrimaryKeys       []ColumnName       `json:"primaryKeys"`
	Relations         []Relation         `json:"relations"`
	UniqueConstraints []UniqueConstraint `json:"uniqueConstraints"`
}

type Snapshot struct {
	Tables []Table `json:"tables"`
}

func getActions(migrationVersion string, actionIndex int) (*[]Action, error) {

	migrations, err := GetList()
	if err != nil {
		return nil, fmt.Errorf("can't read migrations: %v", err)
	}

	actions := []Action{}

	for _, migration := range *migrations {
		for index, action := range migration.Actions {
			actions = append(actions, action)

			if migrationVersion != "" &&
				migration.Id == migrationVersion &&
				actionIndex >= 0 &&
				index >= actionIndex {
				break
			}
		}

		if migrationVersion != "" && migration.Id == migrationVersion {
			break
		}
	}

	return &actions, nil
}

func GetSnapshot(actions []Action) (*Snapshot, error) {

	snapshot := Snapshot{
		Tables: []Table{},
	}

	err := applyActionsToSnapshot(&snapshot, actions)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func GetSnapshotWithAction(method string, params interface{}) (*Snapshot, error) {
	pActions, err := getActions("", -1)
	if err != nil {
		return nil, err
	}

	actions := *pActions
	packedParams, _ := json.MarshalIndent(params, "", "  ")

	actions = append(actions, Action{
		Method: method,
		Params: packedParams,
	})

	return GetSnapshot(actions)
}

func GetCurrentSnapshot() (*Snapshot, error) {
	return GetSnapshotForVersion("", -1)
}

func GetSnapshotForVersion(migrationId string, actionIndex int) (*Snapshot, error) {

	actions, err := getActions(migrationId, actionIndex)
	if err != nil {
		return nil, err
	}

	return GetSnapshot(*actions)
}

func GetStepBackSnapshot(migrationId string, actionIndex int) (*Snapshot, error) {

	pActions, err := getActions(migrationId, actionIndex)
	if err != nil {
		return nil, err
	}

	actions := *pActions

	if len(actions) > 0 {
		actions = actions[:len(actions)-1]
	}

	return GetSnapshot(actions)
}

func applyActionsToSnapshot(snapshot *Snapshot, actions []Action) error {

	for _, action := range actions {

		method, params, err := decodeAction(action.Method, action.Params)
		if err != nil {
			return fmt.Errorf("can't decode action %v/n", err)
		}

		switch method {
		case "addTable":
			err = applyAddTableToSnapshot(snapshot, params.(AddTableParams))
			break
		case "deleteTable":
			err = applyDeleteTableFromSnapshot(snapshot, params.(DeleteTableParams))
			break
		case "addColumn":
			err = applyAddColumnToSnapshot(snapshot, params.(AddColumnParams))
			break
		case "deleteColumn":
			err = applyDeleteColumnFromSnapshot(snapshot, params.(DeleteColumnParams))
			break
		case "addPrimaryKey":
			err = applyAddPrimaryKeyToSnapshot(snapshot, params.(AddPrimaryKeyParams))
			break
		case "deletePrimaryKey":
			err = applyDeletePrimaryKeyFromSnapshot(snapshot, params.(DeletePrimaryKeyParams))
			break
		case "addRelation":
			err = applyAddRelationToSnapshot(snapshot, params.(AddRelationParams))
			break
		case "deleteRelation":
			err = applyDeleteRelationFromSnapshot(snapshot, params.(DeleteRelationParams))
			break
		case "addUniqueConstraint":
			err = applyAddUniqueConstraintToSnapshot(snapshot, params.(AddUniqueConstraintParams))
			break
		case "deleteUniqueConstraint":
			err = applyDeleteUniqueConstraintFromSnapshot(snapshot, params.(DeleteUniqueConstraintParams))
			break
		}

		if err != nil {
			return fmt.Errorf("can't apply action '%v' %v: %v/n", method, params, err)
		}
	}

	return nil
}

func getTableFromSnapshot(snapshot *Snapshot, tableName string) *Table {

	tables := snapshot.Tables

	for index := 0; index < len(tables); index++ {
		table := &(tables[index])
		if table.Name == tableName {
			return table
		}
	}

	return nil
}

func applyAddTableToSnapshot(snapshot *Snapshot, params AddTableParams) error {

	existingTable := getTableFromSnapshot(snapshot, params.Name)
	if existingTable != nil {
		return fmt.Errorf("table '%v' already exist", params.Name)
	}

	snapshot.Tables = append(snapshot.Tables, Table{
		Name:        params.Name,
		Columns:     []Column{},
		PrimaryKeys: []ColumnName{},
		Relations:   []Relation{},
	})

	return nil
}

func applyDeleteTableFromSnapshot(snapshot *Snapshot, params DeleteTableParams) error {

	tableName := params.Name
	existingTable := getTableFromSnapshot(snapshot, tableName)

	if existingTable == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Name)
	}

	for index, table := range snapshot.Tables {
		if table.Name != tableName {
			continue
		}

		snapshot.Tables = append(snapshot.Tables[:index], snapshot.Tables[index+1:]...)
	}

	return nil
}

func getColumnFromTable(table *Table, columnName string) *Column {

	columns := table.Columns

	for index := 0; index < len(columns); index++ {
		column := columns[index]

		if column.Name == columnName {
			return &column
		}
	}

	return nil
}

func applyAddColumnToSnapshot(snapshot *Snapshot, params AddColumnParams) error {
	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	column := getColumnFromTable(table, params.Column)
	if column != nil {
		return fmt.Errorf("column '%v' doesn't exist", params.Column)
	}

	table.Columns = append(table.Columns, Column{
		Name:         params.Column,
		Type:         params.Type,
		IsNullable:   params.IsNullable,
		DefaultValue: params.DefaultValue,
	})

	return nil
}

func applyDeleteColumnFromSnapshot(snapshot *Snapshot, params DeleteColumnParams) error {

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	columnName := params.Column
	column := getColumnFromTable(table, columnName)
	if column == nil {
		return fmt.Errorf("column '%v' doesn't exist", params.Column)
	}

	for index, column := range table.Columns {
		if column.Name != columnName {
			continue
		}

		table.Columns = append(table.Columns[:index], table.Columns[index+1:]...)
	}
	return nil
}

func applyAddPrimaryKeyToSnapshot(snapshot *Snapshot, params AddPrimaryKeyParams) error {

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	column := getColumnFromTable(table, params.Column)
	if column == nil {
		return fmt.Errorf("column '%v' doesn't exist", params.Column)
	}

	for _, columnName := range table.PrimaryKeys {
		if columnName == ColumnName(params.Column) {
			return fmt.Errorf("primary key for column '%v' allready exist", params.Column)
		}
	}

	table.PrimaryKeys = append(table.PrimaryKeys, ColumnName(params.Column))
	return nil
}

func applyDeletePrimaryKeyFromSnapshot(snapshot *Snapshot, params DeletePrimaryKeyParams) error {

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	column := getColumnFromTable(table, params.Column)
	if column == nil {
		return fmt.Errorf("column '%v' doesn't exist", params.Column)
	}

	keyIndex := -1

	for index, columnName := range table.PrimaryKeys {
		if columnName == ColumnName(params.Column) {
			keyIndex = index
		}
	}

	if keyIndex == -1 {
		return fmt.Errorf("primary key for column '%v' doesn't exist", params.Column)
	}

	table.PrimaryKeys = append(table.PrimaryKeys[:keyIndex], table.PrimaryKeys[keyIndex+1:]...)
	return nil
}

func applyAddRelationToSnapshot(snapshot *Snapshot, params AddRelationParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("relation name is required")
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	remoteTable := getTableFromSnapshot(snapshot, params.RemoteTable)
	if remoteTable == nil {
		return fmt.Errorf("remote table '%v' doesn't exist", params.RemoteTable)
	}

	table.Relations = append(table.Relations, Relation{
		Name:           params.Name,
		Type:           params.Type,
		RemoteTable:    params.RemoteTable,
		ColumnsMapping: params.ColumnsMapping,
	})
	return nil
}

func applyDeleteRelationFromSnapshot(snapshot *Snapshot, params DeleteRelationParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("relation name is required")
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	for index, relation := range table.Relations {
		if relation.Name == params.Name {
			table.Relations = append(table.Relations[:index], table.Relations[index+1:]...)
			return nil
		}
	}

	return fmt.Errorf("relation \"%v\" doesn't exist", params.Name)
}

func applyAddUniqueConstraintToSnapshot(snapshot *Snapshot, params AddUniqueConstraintParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("constraint name is required")
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	if len(params.Name) == 0 {
		return fmt.Errorf("columns are required")
	}

	table.UniqueConstraints = append(table.UniqueConstraints, UniqueConstraint{
		Name:    params.Name,
		Columns: params.Columns,
	})
	return nil
}

func applyDeleteUniqueConstraintFromSnapshot(snapshot *Snapshot, params DeleteUniqueConstraintParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("constraint name is required")
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	for index, constraint := range table.UniqueConstraints {
		if constraint.Name == params.Name {
			table.UniqueConstraints = append(table.UniqueConstraints[:index], table.UniqueConstraints[index+1:]...)
			return nil
		}
	}

	return fmt.Errorf("constraint \"%v\" doesn't exist", params.Name)
}
