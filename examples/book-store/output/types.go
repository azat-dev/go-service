package executor

import (
	"fmt"
	validator "github.com/asaskevich/govalidator"
)

type Validatable interface {
	Validate() error
}

/////////////////////////////////////////////////////////////////////
//BookType

type BookType string

const (
	Magazine BookType = "magazineItem"
	BookItem BookType = "book"
)

/////////////////////////////////////////////////////////////////////
//Book

type Book struct {
	Type      BookType `json:"type"`
	Id        string   `json:"id"`
	AuthorId  string   `json:"authorId"`
	CreatedAt int64    `json:"createdAt"`
	Title     string   `json:"title"`
}

func (v *Book) Validate() error {

	{
		value := v.Type
		isValid := value.Validate() == nil

		if !isValid {
			return fmt.Errorf("Type is invalid")
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
		value := v.AuthorId
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("AuthorId is invalid")
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
//Author

type Author struct {
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	Patronymic *string `json:"patronymic"`
	Id         string  `json:"id"`
}

func (v *Author) Validate() error {

	{
		value := v.Name
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Name is invalid")
		}
	}

	{
		value := v.Surname
		isValid := (len(value) >= 0 && len(value) <= 255)

		if !isValid {
			return fmt.Errorf("Surname is invalid")
		}
	}

	{
		value := v.Patronymic
		isValid := value == nil || (len(*value) >= 0 && len(*value) <= 255)

		if !isValid {
			return fmt.Errorf("Patronymic is invalid")
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
//PARAMETERS

/////////////////////////////////////////////////////////////////////
//getAuthors
type GetAuthorsParams struct {
	Id string `json:"id"`
}

func (v *GetAuthorsParams) Validate() error {

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
//getBook
type GetBookParams struct {
	Id string `json:"id"`
}

func (v *GetBookParams) Validate() error {

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
//getBooks
type GetBooksParams struct {
	Id string `json:"id"`
}

func (v *GetBooksParams) Validate() error {

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
//getAuthor
type GetAuthorParams struct {
	Id string `json:"id"`
}

func (v *GetAuthorParams) Validate() error {

	{
		value := v.Id
		isValid := validator.IsUUID(value)

		if !isValid {
			return fmt.Errorf("Id is invalid")
		}
	}

	return nil
}
