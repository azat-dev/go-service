# ![Logo](https://s3-eu-west-1.amazonaws.com/timeio.xyz/logo.svg) go-service

> Boilerplate code generator for json-rpc services.
  Features: parsing json to structures, input data validation

## Installation

    go get github.com/akaumov/go-service
    
## Quick usage guide

### 1.Create a YAML file with your service description. Example:

```yaml
---
version: 0.0.1
name: booksApi
package: executor
description: api service for books store
types:
  Book:
    id: uuid
    authorId: uuid
    createdAt: time
    title: string(0,255)
  Author:
    id: uuid
    name: string(0,255)
    surname: string (0,255)
    patronymic: string(0,255)?
methods:
  getBook:
    params:
      id: uuid
    result: Book
  getBooks:
    params:
      id: uuid
    result: "[]Book"
  getAuthor:
    params:
      id: uuid
    result: Author
  getAuthors:
    params:
      id: uuid
    result: "[]Author"
 ```
 

### 2.Run command
 

    go-service /path-to-your-schema-file /output-directory


### 3. Look in your output directory 3 files:
- **executor.go** - contains object that will run your code
- **handler_interface.go** - contains interface for handler;
- **types.go** - contains generated types

### 4. Implement your handler interface

```go
type Handler struct {
}

fun (h * Handler) GetAuthor(session SessionInterface, id string) (*Author, error) {
    //..some code here
}

fun (h * Handler) GetAuthors(session SessionInterface, id string) (*[]Author, error) {
    //..some code here
}

fun (h * Handler) GetBook(session SessionInterface, id string) (*Book, error) {
 //..some code here
}

fun (h * Handler) GetBooks(session SessionInterface, id string) (*[]Book, error) {
    //..some code here
}

var _ executor.HandlerInterface = (*Handler)(nil)
```

### 5. Run your handler with executor

```go
response, error := h.executor.Execute(session, inputJsonText)
```


## Copyright and licensing
 
Unless otherwise noted, the source files are distributed under the *MIT License*
found in the LICENSE.txt file.