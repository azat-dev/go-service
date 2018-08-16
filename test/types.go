package executor

import (
	"encoding/json"
	validator "github.com/asaskevich/govalidator"
	"github.com/pkg/errors"
	"fmt"
)

type Validatable interface {
	Validate() error
}

/////////////////////////////////////////////////////////////////////
//Folder

type Folder struct {
	ChildrenTasksOrder   []string `json:"childrenTasksOrder"`
	Id                   string   `json:"id"`
	OwnerId              string   `json:"ownerId"`
	ParentId             *string  `json:"parentId"`
	CreatedAt            int64    `json:"createdAt"`
	UpdatedAt            int64    `json:"updatedAt"`
	Title                string   `json:"title"`
	ChildrenFoldersOrder []string `json:"childrenFoldersOrder"`
}

func (v *Folder) Validate() error {

	{
		value := v.Title
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Title is invalid")
		}
	}

	{
		value := v.ChildrenFoldersOrder
		isValid := (len(value) >= 0 && len(value) <= 100) &&
			func(value *[]string) bool {
				for _, item := range *value {
					isValid := validator.IsUUID(item)

					if !isValid {
						return false
					}
				}
				return true
			}(&value)

		if !isValid {
			return fmt.Errorf("ChildrenFoldersOrder is invalid")
		}
	}

	{
		value := v.ChildrenTasksOrder
		isValid := (len(value) >= 0 && len(value) <= 100) &&
			func(value *[]string) bool {
				for _, item := range *value {
					isValid := validator.IsUUID(item)

					if !isValid {
						return false
					}
				}
				return true
			}(&value)

		if !isValid {
			return fmt.Errorf("ChildrenTasksOrder is invalid")
		}
	}

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	{
		value := v.OwnerId
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("OwnerId is invalid")
		}
	}

	{
		value := v.ParentId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("ParentId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//OperationAddFolderParams

type OperationAddFolderParams struct {
	Position int     `json:"position"`
	Id       string  `json:"id"`
	Title    string  `json:"title"`
	ParentId *string `json:"parentId"`
}

func (v *OperationAddFolderParams) Validate() error {

	{
		value := v.Title
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Title is invalid")
		}
	}

	{
		value := v.ParentId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("ParentId is invalid")
		}
	}

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//SyncResult

type SyncResult struct {
	Status string `json:"status"`
}

func (v *SyncResult) Validate() error {
	return nil
}

/////////////////////////////////////////////////////////////////////
//Task

type Task struct {
	Id          string `json:"id"`
	OwnerId     string `json:"ownerId"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	IsCompleted bool   `json:"isCompleted"`
	Title       string `json:"title"`
}

func (v *Task) Validate() error {

	{
		value := v.Title
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Title is invalid")
		}
	}

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	{
		value := v.OwnerId
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("OwnerId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//GetChildrenItemsResult

type GetChildrenItemsResult struct {
	Tasks   []Task   `json:"tasks"`
	Folders []Folder `json:"folders"`
}

func (v *GetChildrenItemsResult) Validate() error {

	{
		value := v.Folders
		isValid := func(value *[]Folder) bool {
			for _, item := range *value {
				isValid := item.Validate() != nil

				if !isValid {
					return false
				}
			}
			return true
		}(&value)

		if !isValid {
			return fmt.Errorf("Folders is invalid")
		}
	}

	{
		value := v.Tasks
		isValid := func(value *[]Task) bool {
			for _, item := range *value {
				isValid := item.Validate() != nil

				if !isValid {
					return false
				}
			}
			return true
		}(&value)

		if !isValid {
			return fmt.Errorf("Tasks is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//OperationAddTaskParams

type OperationAddTaskParams struct {
	Position int     `json:"position"`
	FolderId *string `json:"folderId"`
	Id       string  `json:"id"`
	Title    string  `json:"title"`
}

func (v *OperationAddTaskParams) Validate() error {

	{
		value := v.FolderId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("FolderId is invalid")
		}
	}

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	{
		value := v.Title
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Title is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//OperationDeleteTaskParams

type OperationDeleteTaskParams struct {
	Id string `json:"id"`
}

func (v *OperationDeleteTaskParams) Validate() error {

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//OperationDeleteFolderParams

type OperationDeleteFolderParams struct {
	Id string `json:"id"`
}

func (v *OperationDeleteFolderParams) Validate() error {

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//OperationItem

type OperationItem struct {
	ServerVersion int         `json:"serverVersion"`
	Params        interface{} `json:"params"`
	Index         int         `json:"index"`
	Method        string      `json:"method"`
	Time          int64       `json:"time"`
}

func (v *OperationItem) Validate() error {

	{
		value := v.Params
		isValid := func(value interface{}) bool {
			switch value.(type) {

			case OperationAddTaskParams:
				x := value.(OperationAddTaskParams)
				return (&x).Validate() == nil

			case OperationDeleteTaskParams:
				x := value.(OperationDeleteTaskParams)
				return (&x).Validate() == nil

			case OperationAddFolderParams:
				x := value.(OperationAddFolderParams)
				return (&x).Validate() == nil

			case OperationDeleteFolderParams:
				x := value.(OperationDeleteFolderParams)
				return (&x).Validate() == nil

			default:
				panic("not implemented")
			}
		}(value)

		if !isValid {
			return fmt.Errorf("Params is invalid")
		}
	}

	return nil
}

func (v *OperationItem) UnmarshalJSON(packed []byte) error {
	commonFields := struct {
		Index         int
		Method        string
		Time          int64
		ServerVersion int

		Params json.RawMessage
	}{}

	err := json.Unmarshal(packed, &commonFields)
	if err != nil {
		return err
	}

	v.Index = commonFields.Index
	v.Method = commonFields.Method
	v.Time = commonFields.Time
	v.ServerVersion = commonFields.ServerVersion

	switch v.Method {

	case "deleteTask":
		var parsedData OperationDeleteTaskParams
		err = json.Unmarshal(commonFields.Params, &parsedData)
		if err != nil {
			return err
		}

		v.Params = parsedData
		break

	case "addFolder":
		var parsedData OperationAddFolderParams
		err = json.Unmarshal(commonFields.Params, &parsedData)
		if err != nil {
			return err
		}

		v.Params = parsedData
		break

	case "deleteFolder":
		var parsedData OperationDeleteFolderParams
		err = json.Unmarshal(commonFields.Params, &parsedData)
		if err != nil {
			return err
		}

		v.Params = parsedData
		break

	case "addTask":
		var parsedData OperationAddTaskParams
		err = json.Unmarshal(commonFields.Params, &parsedData)
		if err != nil {
			return err
		}

		v.Params = parsedData
		break

	default:
		return errors.New("invalid method value")
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//PARAMETERS

/////////////////////////////////////////////////////////////////////
//getTask
type GetTaskParams struct {
	Id string `json:"id"`
}

func (v *GetTaskParams) Validate() error {

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//getTasks
type GetTasksParams struct {
	FolderId *string `json:"folderId"`
}

func (v *GetTasksParams) Validate() error {

	{
		value := v.FolderId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("FolderId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//getFolder
type GetFolderParams struct {
	FolderId *string `json:"folderId"`
}

func (v *GetFolderParams) Validate() error {

	{
		value := v.FolderId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("FolderId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//getFolders
type GetFoldersParams struct {
	ParentId *string `json:"parentId"`
}

func (v *GetFoldersParams) Validate() error {

	{
		value := v.ParentId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("ParentId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//getChildrenItems
type GetChildrenItemsParams struct {
	FolderId *string `json:"folderId"`
}

func (v *GetChildrenItemsParams) Validate() error {

	{
		value := v.FolderId
		isValid := value == nil || validator.IsUUID(*value)

		if !isValid {
			return fmt.Errorf("FolderId is invalid")
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////
//sync
type SyncParams struct {
	Operations []OperationItem `json:"operations"`
}

func (v *SyncParams) Validate() error {

	{
		value := v.Operations
		isValid := func(value *[]OperationItem) bool {
			for _, item := range *value {
				isValid := item.Validate() != nil

				if !isValid {
					return false
				}
			}
			return true
		}(&value)

		if !isValid {
			return fmt.Errorf("Operations is invalid")
		}
	}

	return nil
}
