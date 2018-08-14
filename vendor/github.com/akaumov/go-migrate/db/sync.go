package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func applyAddTable(transaction *sql.Tx, params AddTableParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("table is required")
	}

	query := fmt.Sprintf("CREATE TABLE \"%v\" ();", params.Name)
	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't create table %v: %v\n", params.Name, err)
	}

	return nil
}

func applyDeleteTable(transaction *sql.Tx, params DeleteTableParams) error {

	if strings.TrimSpace(params.Name) == "" {
		return fmt.Errorf("table is required")
	}

	query := fmt.Sprintf("DROP TABLE \"%v\"", params.Name)
	_, err := transaction.Exec(query)

	if err != nil {
		return fmt.Errorf("can't delete table %v: %v\n", params.Name, err)
	}

	return nil
}

func applyAddColumn(transaction *sql.Tx, params AddColumnParams) error {

	if strings.TrimSpace(params.Table) == "" {
		return fmt.Errorf("table is required")
	}

	if strings.TrimSpace(params.Column) == "" {
		return fmt.Errorf("column is required")
	}

	columnType := params.Type
	notNullParam := ""
	if !params.IsNullable {
		notNullParam = "NOT NULL"
	}

	defaultValueParam := ""
	if params.DefaultValue != "" {
		defaultValueParam = fmt.Sprintf("DEFAULT %v;", params.DefaultValue)
	}

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			ADD COLUMN "%v" %v %v %v
	`, params.Table, params.Column, columnType, notNullParam, defaultValueParam)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't add column '%v' to table '%v': %v \n", params.Column, params.Table, err, query)
	}

	return nil
}

func applyDeleteColumn(transaction *sql.Tx, params DeleteColumnParams) error {

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			DROP COLUMN "%v"
	`, params.Table, params.Column)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't delete column '%v' at table '%v': %v\n", params.Column, params.Table, err)
	}

	return nil
}

func applyAddPrimaryKey(transaction *sql.Tx, migrationId string, actionIndex int, params AddPrimaryKeyParams) error {

	snapshot, err := GetSnapshotForVersion(migrationId, actionIndex)
	if err != nil {
		return err
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	column := getColumnFromTable(table, params.Column)
	if column == nil {
		return fmt.Errorf("column '%v' doesn't exist", params.Column)
	}

	constraintName := params.Table + "_pkey"

	if len(table.PrimaryKeys) > 1 {
		query := fmt.Sprintf(`
			ALTER TABLE "%v"
				DROP CONSTRAINT "%v"
		`, params.Table, constraintName)

		_, err := transaction.Exec(query)
		if err != nil {
			return err
		}
	}

	keys := ""
	for index, key := range table.PrimaryKeys {
		if index == 0 {
			keys = fmt.Sprintf(`"%v"`, key)
		} else {
			keys += fmt.Sprintf(`, "%v"`, key)
		}

	}

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			ADD CONSTRAINT "%v" PRIMARY KEY (%v);
	`, params.Table, constraintName, keys)

	_, err = transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't add primary key '%v' to table '%v': %v\n", params.Column, params.Table, err)
	}

	return nil
}

func applyDeletePrimaryKey(transaction *sql.Tx, migrationId string, actionIndex int, params DeletePrimaryKeyParams) error {

	constraintName := params.Table + "_pkey"

	snapshot, err := GetSnapshotForVersion(migrationId, actionIndex)
	if err != nil {
		return err
	}

	table := getTableFromSnapshot(snapshot, params.Table)
	if table == nil {
		return fmt.Errorf("table '%v' doesn't exist", params.Table)
	}

	query := fmt.Sprintf(`
			ALTER TABLE "%v"
				DROP CONSTRAINT "%v"
		`, params.Table, constraintName)

	_, err = transaction.Exec(query)
	if err != nil {
		return err
	}

	keys := ""
	for _, key := range table.PrimaryKeys {
		if key == ColumnName(params.Column) {
			continue
		}

		if keys == "" {
			keys = fmt.Sprintf(`"%v"`, key)
		} else {
			keys += fmt.Sprintf(`, "%v"`, key)
		}

	}

	query = fmt.Sprintf(`
		ALTER TABLE "%v"
			ADD CONSTRAINT pkey PRIMARY KEY (%v);
	`, params.Table, keys)

	_, err = transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't add primary key '%v' to table '%v': %v\n", params.Column, params.Table, err)
	}

	return nil
}

func applyAddRelation(transaction *sql.Tx, params AddRelationParams) error {

	columns := ""
	remoteColumns := ""

	for _, mapping := range params.ColumnsMapping {
		if columns == "" {
			columns = fmt.Sprintf(`"%v"`, mapping.Column)
			remoteColumns = fmt.Sprintf(`"%v"`, mapping.RemoteColumn)
		} else {
			columns += fmt.Sprintf(`, "%v"`, mapping.Column)
			remoteColumns += fmt.Sprintf(`, "%v"`, mapping.RemoteColumn)
		}
	}

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			ADD CONSTRAINT "%v" FOREIGN KEY (%v)
			REFERENCES "%v" (%v) MATCH SIMPLE
			ON UPDATE NO ACTION
			ON DELETE NO ACTION;
	`, params.Table, params.Name, columns, params.RemoteTable, remoteColumns)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't add relation '%v' to table '%v': %v\n", params.Name, params.Table, err)
	}

	return nil
}

func applyAddUniqueConstraint(transaction *sql.Tx, params AddUniqueConstraintParams) error {

	columns := ""

	for _, column := range params.Columns {
		if columns == "" {
			columns = fmt.Sprintf(`"%v"`, column)
		} else {
			columns += fmt.Sprintf(`, "%v"`, column)
		}
	}

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			ADD CONSTRAINT "%v" UNIQUE (%v)
	`, params.Table, params.Name, columns)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't add unique constraint '%v' to table '%v': %v\n", params.Name, params.Table, err)
	}

	return nil
}

func applyDeleteRelation(transaction *sql.Tx, params DeleteRelationParams) error {

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			DROP CONSTRAINT "%v"
	`, params.Table, params.Name)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't delete relation '%v' to table '%v': %v\n", params.Name, params.Table, err)
	}

	return nil
}

func applyDeleteUniqueConstraint(transaction *sql.Tx, params DeleteUniqueConstraintParams) error {

	query := fmt.Sprintf(`
		ALTER TABLE "%v"
			DROP CONSTRAINT "%v"
	`, params.Table, params.Name)

	_, err := transaction.Exec(query)
	if err != nil {
		return fmt.Errorf("can't delete unique constraint '%v' to table '%v': %v\n", params.Name, params.Table, err)
	}

	return nil
}

func PingDb(userName string, password string, dbName string, host string, port int) error {

	dbConnectionString := fmt.Sprintf("user=%v password=%v dbname=%v host=%v port=%v sslmode=disable",
		userName,
		password,
		dbName,
		host,
		port)

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		return fmt.Errorf("can't connect to db: %v", err)
	}
	defer func() { db.Close() }()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("can't connect to db: %v", err)
	}

	log.Println("Connected to db")
	return nil
}

func Sync(migrationsDir string, userName string, password string, dbName string, host string, port int) (string, error) {

	migrations, err := GetList()
	if err != nil {
		return "", fmt.Errorf("can't read migrations: %v\n", err)
	}

	dbConnectionString := fmt.Sprintf("user=%v password=%v dbname=%v host=%v port=%v sslmode=disable",
		userName,
		password,
		dbName,
		host,
		port)

	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		return "", fmt.Errorf("can't connect to db: %v", err)
	}
	defer func() { db.Close() }()

	err = db.Ping()
	if err != nil {
		return "", fmt.Errorf("can't connect to db: %v", err)
	}

	log.Println("Connected to db")
	transaction, err := db.Begin()
	if err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("can't start transaction: %v", err)
	}

	err = addMigrationsTableIfNotExist(transaction)
	if err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("can't add migration table: %v", err)
	}

	currentMigrationId, err := getCurrentSyncedMigrationId(transaction)
	if err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("can't read current migration state: %v", err)
	}

	_, err = GetCurrentSnapshot()
	if err != nil {
		return "", err
	}

	var lastMigration *Migration
	isCurrentMigrationPassed := currentMigrationId == ""

	for _, migration := range *migrations {

		if migration.Id == currentMigrationId {
			isCurrentMigrationPassed = true
			continue
		}

		if !isCurrentMigrationPassed {
			continue
		}

		err = applyMigrationActions(transaction, migration)
		if err != nil {
			transaction.Rollback()
			return "", fmt.Errorf("can't apply migration %v: %v\n", migration.Id, err)
		}

		addMigrationToMigrationsTable(transaction, migration)
		if err != nil {
			transaction.Rollback()
			return "", fmt.Errorf("can't add migration to migrations table %v: %v\n", migration.Id, err)
		}

		lastMigration = &migration
	}

	err = transaction.Commit()
	if err != nil {
		return "", err
	}

	if lastMigration != nil {
		return lastMigration.Id, nil
	}

	return "", fmt.Errorf("no migrations")
}

func getCurrentSyncedMigrationId(transaction *sql.Tx) (string, error) {

	row := transaction.QueryRow("SELECT id FROM _migrations  ORDER BY id DESC  LIMIT 1")

	var migrationId string
	err := row.Scan(&migrationId)
	if err == sql.ErrNoRows {
		return "", nil
	}

	return migrationId, err
}

func applyMigrationActions(transaction *sql.Tx, migration Migration) error {

	fmt.Println(migration.Id)

	for index, action := range migration.Actions {

		var err error

		method, params, err := decodeAction(action.Method, action.Params)
		if err != nil {
			return fmt.Errorf("can't decode action %v\n", err)
		}

		switch method {
		case "addTable":
			err = applyAddTable(transaction, params.(AddTableParams))
			break
		case "deleteTable":
			err = applyDeleteTable(transaction, params.(DeleteTableParams))
			break
		case "addColumn":
			err = applyAddColumn(transaction, params.(AddColumnParams))
			break
		case "deleteColumn":
			err = applyDeleteColumn(transaction, params.(DeleteColumnParams))
			break
		case "addPrimaryKey":
			err = applyAddPrimaryKey(transaction, migration.Id, index, params.(AddPrimaryKeyParams))
			break
		case "deletePrimaryKey":
			err = applyDeletePrimaryKey(transaction, migration.Id, index, params.(DeletePrimaryKeyParams))
			break
		case "addRelation":
			err = applyAddRelation(transaction, params.(AddRelationParams))
			break
		case "deleteRelation":
			err = applyDeleteRelation(transaction, params.(DeleteRelationParams))
			break
		case "addUniqueConstraint":
			err = applyAddUniqueConstraint(transaction, params.(AddUniqueConstraintParams))
			break
		case "deleteUniqueConstraint":
			err = applyDeleteUniqueConstraint(transaction, params.(DeleteUniqueConstraintParams))
			break
		}

		if err != nil {
			fmt.Println("#"+strconv.Itoa(index), method, "error")
			return fmt.Errorf("can't apply action #%v=\"%v\": %v\n", index, method, err)
		} else {
			fmt.Println("#"+strconv.Itoa(index), method, "success", "")
		}
	}

	fmt.Println()

	return nil
}

func decodeAction(method string, params json.RawMessage) (string, interface{}, error) {

	var err error
	switch method {
	case "addTable":
		var addTableParams AddTableParams
		err = json.Unmarshal(params, &addTableParams)
		if err != nil {
			return "", nil, err
		}

		return method, addTableParams, nil

	case "deleteTable":
		var deleteTableParams DeleteTableParams
		err = json.Unmarshal(params, &deleteTableParams)
		if err != nil {
			return "", nil, err
		}

		return method, deleteTableParams, nil

	case "addColumn":
		var addColumnParams AddColumnParams
		err = json.Unmarshal(params, &addColumnParams)
		if err != nil {
			return "", nil, err
		}

		return method, addColumnParams, nil

	case "deleteColumn":
		var deleteColumnParams DeleteColumnParams
		err = json.Unmarshal(params, &deleteColumnParams)
		if err != nil {
			return "", nil, err
		}

		return method, deleteColumnParams, nil

	case "addPrimaryKey":
		var addPrimaryKeyParams AddPrimaryKeyParams
		err = json.Unmarshal(params, &addPrimaryKeyParams)
		if err != nil {
			return "", nil, err
		}

		return method, addPrimaryKeyParams, nil

	case "deletePrimaryKey":
		var deletePrimaryKeyParams DeletePrimaryKeyParams
		err = json.Unmarshal(params, &deletePrimaryKeyParams)
		if err != nil {
			return "", nil, err
		}

		return method, deletePrimaryKeyParams, nil

	case "addRelation":
		var addRelationParams AddRelationParams
		err = json.Unmarshal(params, &addRelationParams)
		if err != nil {
			return "", nil, err
		}

		return method, addRelationParams, nil

	case "deleteRelation":
		var deleteRelationParams DeleteRelationParams
		err = json.Unmarshal(params, &deleteRelationParams)
		if err != nil {
			return "", nil, err
		}

		return method, deleteRelationParams, nil

	case "addUniqueConstraint":
		var addUniqueConstraintParams AddUniqueConstraintParams
		err = json.Unmarshal(params, &addUniqueConstraintParams)
		if err != nil {
			return "", nil, err
		}

		return method, addUniqueConstraintParams, nil

	case "deleteUniqueConstraint":
		var deleteUniqueConstraintParams DeleteUniqueConstraintParams
		err = json.Unmarshal(params, &deleteUniqueConstraintParams)
		if err != nil {
			return "", nil, err
		}

		return method, deleteUniqueConstraintParams, nil
	}

	return "", nil, nil
}

func addMigrationsTableIfNotExist(transaction *sql.Tx) error {
	_, err := transaction.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
        	id varchar(255) NOT NULL,
        	data text NOT NULL,
        	PRIMARY KEY (id)
    )`)

	return err
}

func addMigrationToMigrationsTable(transaction *sql.Tx, migration Migration) error {
	packedMigration, _ := json.Marshal(migration)
	_, err := transaction.Exec("INSERT INTO _migrations (id, data) VALUES ($1, $2)", migration.Id, packedMigration)
	return err
}
