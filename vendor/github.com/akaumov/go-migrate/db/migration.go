package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ghodss/yaml"

	_ "github.com/lib/pq"
)

const migrationsDirectoryName = "migrations"

type Format string

const (
	YAML Format = "yaml"
	JSON Format = "json"
)

type ColumnName string

type AddTableParams struct {
	Name string `json:"name"`
}

type DeleteTableParams struct {
	Name string `json:"name"`
}

type AddColumnParams struct {
	Table        string `json:"table"`
	Column       string `json:"column"`
	Type         string `json:"type"`
	IsNullable   bool   `json:"isNullable"`
	DefaultValue string `json:"defaultValue"`
}

type DeleteColumnParams struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type AddPrimaryKeyParams struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type DeletePrimaryKeyParams struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type AddUniqueConstraintParams struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type DeleteUniqueConstraintParams struct {
	Table string `json:"table"`
	Name  string `json:"name"`
}

type RelationType string

const (
	Object = RelationType("object")
	Array  = RelationType("array")
)

type ColumnsMap struct {
	Column       string `json:"column"`
	RemoteColumn string `json:"remoteColumn"`
}

type AddRelationParams struct {
	Type           RelationType `json:"type"`
	Name           string       `json:"name"`
	Table          string       `json:"table"`
	RemoteTable    string       `json:"remoteTable"`
	ColumnsMapping []ColumnsMap `json:"columnsMapping"`
}

type DeleteRelationParams struct {
	Table string `json:"table"`
	Name  string `json:"name"`
}

type Action struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Migration struct {
	SchemaVersion string   `json:"schemaVersion"`
	Id            string   `json:"id"`
	Description   string   `json:"description"`
	Actions       []Action `json:"actions"`
}

func GetMigrationsDirectoryPath() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	directory := filepath.Join(pwd, migrationsDirectoryName)
	return directory, nil
}

func AddMigration(description string, outputFormat Format) (string, error) {

	dateId := time.Now().UTC().Format("20060102150405")

	descriptionId := strings.ToLower(description)
	descriptionId = strings.Replace(descriptionId, " ", "_", -1)

	descriptionIdLength := len(descriptionId)
	if descriptionIdLength > 50 {
		descriptionIdLength = 50
	}
	descriptionId = descriptionId[:descriptionIdLength]

	var fileName string
	if descriptionId != "" {
		fileName = fmt.Sprintf("%v_%v.%v", dateId, descriptionId, outputFormat)
	} else {
		fileName = fmt.Sprintf("%v.%v", dateId, outputFormat)
	}

	migration := Migration{
		SchemaVersion: "1",
		Id:            dateId,
		Description:   description,
		Actions:       []Action{},
	}

	migrationsDir, err := GetMigrationsDirectoryPath()
	if err != nil {
		return "", err
	}

	//TODO: add checking usage of instance name
	if _, err := os.Stat(migrationsDir); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		err = os.Mkdir(migrationsDir, 0777)
		if err != nil {
			return "", err
		}
	}

	var packedMigration []byte

	switch outputFormat {
	case JSON:
		packedMigration, _ = json.MarshalIndent(migration, "", "  ")
	case YAML:
		packedMigration, _ = yaml.Marshal(migration)
	}

	if err != nil {
		return "", err
	}

	return fileName, ioutil.WriteFile(filepath.Join(migrationsDir, fileName), packedMigration, 0777)
}

func getMigrationPath(id string) (string, error) {

	migrationsDirectoryPath, err := GetMigrationsDirectoryPath()
	if err != nil {
		return "", err
	}

	configsPathPattern := filepath.Join(migrationsDirectoryPath, id+"*.json")
	files, err := filepath.Glob(configsPathPattern)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no such migration")
	}

	_, fileName := filepath.Split(files[0])
	migrationPath := filepath.Join(migrationsDirectoryPath, fileName)
	return migrationPath, nil
}

func GetText(id string) (string, error) {

	migrationPath, err := getMigrationPath(id)
	if err != nil {
		return "", nil
	}

	migration, err := ioutil.ReadFile(migrationPath)
	return string(migration), nil
}

func Get(id string) (*Migration, error) {
	rawMigration, err := GetText(id)
	if err != nil {
		return nil, err
	}

	var migration Migration
	err = json.Unmarshal(([]byte)(rawMigration), &migration)

	if err != nil {
		return nil, fmt.Errorf("can't parse migration: %v/n", err)
	}

	return &migration, nil
}

func GetList() (*[]Migration, error) {

	migrationsDirectoryPath, err := GetMigrationsDirectoryPath()
	if err != nil {
		return nil, err
	}

	configsPathPattern := filepath.Join(migrationsDirectoryPath, "*.json")
	files, err := filepath.Glob(configsPathPattern)

	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	result := []Migration{}

	for _, migrationPath := range files {
		_, fileName := filepath.Split(migrationPath)
		migrationId := strings.TrimSuffix(fileName, ".json")

		migration, err := Get(migrationId)
		if err != nil {
			return nil, fmt.Errorf("can't read migration %v/n", err)
		}

		result = append(result, *migration)
	}

	return &result, err
}

func addActionToMigrationFile(method string, params interface{}) (string, error) {

	migrations, err := GetList()
	if err != nil {
		return "", fmt.Errorf("can't get migration %v/n", err)
	}

	migrationsSize := len(*migrations)
	if migrationsSize == 0 {
		return "", fmt.Errorf("migration doesn't exist, please add migration/n")
	}

	_, err = GetSnapshotWithAction(method, params)
	if err != nil {
		return "", err
	}

	packedParams, _ := json.MarshalIndent(params, "", "  ")

	lastMigration := (*migrations)[migrationsSize-1]
	action := Action{
		Method: method,
		Params: (json.RawMessage)(packedParams),
	}

	lastMigration.Actions = append(lastMigration.Actions, action)

	packedMigration, _ := json.MarshalIndent(lastMigration, "", "  ")
	migrationPath, _ := getMigrationPath(lastMigration.Id)
	err = ioutil.WriteFile(migrationPath, packedMigration, 0777)
	if err != nil {
		return "", fmt.Errorf("can't write migration/n")
	}

	return lastMigration.Id, nil
}

func AddTable(tableName string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	params := AddTableParams{
		Name: tableName,
	}

	return addActionToMigrationFile("addTable", params)
}

func DeleteTable(tableName string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	params := DeleteTableParams{
		Name: tableName,
	}

	return addActionToMigrationFile("deleteTable", params)
}

func AddColumn(tableName string, columnName string, columnType string, isNullable bool, defaultValue string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(columnName) == "" {
		return "", fmt.Errorf("column name is required /n")
	}

	if strings.TrimSpace(columnType) == "" {
		return "", fmt.Errorf("column type is required /n")
	}

	params := AddColumnParams{
		Table:        tableName,
		Column:       columnName,
		IsNullable:   isNullable,
		Type:         columnType,
		DefaultValue: defaultValue,
	}

	return addActionToMigrationFile("addColumn", params)
}

func DeleteColumn(tableName string, columnName string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(columnName) == "" {
		return "", fmt.Errorf("column name is required /n")
	}

	params := DeleteColumnParams{
		Table:  tableName,
		Column: columnName,
	}

	return addActionToMigrationFile("deleteColumn", params)
}

func AddPrimaryKey(tableName string, columnName string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(columnName) == "" {
		return "", fmt.Errorf("column name is required /n")
	}

	params := AddPrimaryKeyParams{
		Table:  tableName,
		Column: columnName,
	}

	return addActionToMigrationFile("addPrimaryKey", params)
}

func DeletePrimaryKey(tableName string, columnName string) (string, error) {

	if strings.TrimSpace(tableName) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(columnName) == "" {
		return "", fmt.Errorf("column name is required /n")
	}

	params := DeletePrimaryKeyParams{
		Table:  tableName,
		Column: columnName,
	}

	return addActionToMigrationFile("deletePrimaryKey", params)
}

func AddRelation(relationName string, relationType RelationType, table string, remoteTable string, columnsMapping []ColumnsMap) (string, error) {

	if strings.TrimSpace(table) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(relationName) == "" {
		return "", fmt.Errorf("relation name is required /n")
	}

	params := AddRelationParams{
		Name:           relationName,
		Table:          table,
		Type:           relationType,
		RemoteTable:    remoteTable,
		ColumnsMapping: columnsMapping,
	}

	return addActionToMigrationFile("addRelation", params)
}

func DeleteRelation(table string, relationName string) (string, error) {

	if strings.TrimSpace(table) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(relationName) == "" {
		return "", fmt.Errorf("relation name is required /n")
	}

	params := DeleteRelationParams{
		Name:  relationName,
		Table: table,
	}

	return addActionToMigrationFile("deleteRelation", params)
}

func AddUniqueConstraint(constrtaintName string, table string, columns []string) (string, error) {

	if strings.TrimSpace(table) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(constrtaintName) == "" {
		return "", fmt.Errorf("constraint name is required /n")
	}

	if len(columns) == 0 {
		return "", fmt.Errorf("columns are required /n")
	}

	params := AddUniqueConstraintParams{
		Name:    constrtaintName,
		Table:   table,
		Columns: columns,
	}

	return addActionToMigrationFile("addUniqueConstraint", params)
}

func DeleteUniqueConstraint(table string, constrtaintName string) (string, error) {

	if strings.TrimSpace(table) == "" {
		return "", fmt.Errorf("table name is required /n")
	}

	if strings.TrimSpace(constrtaintName) == "" {
		return "", fmt.Errorf("constraint name is required /n")
	}

	params := DeleteUniqueConstraintParams{
		Name:  constrtaintName,
		Table: table,
	}

	return addActionToMigrationFile("deleteUniqueConstraint", params)
}
