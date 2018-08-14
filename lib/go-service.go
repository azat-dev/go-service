package lib

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"path/filepath"
)

func Build(serviceSchemaPath string, outputPath string) error {
	rawSchema, err := ioutil.ReadFile(serviceSchemaPath)
	if err != nil {
		return err
	}

	var service Service
	err = yaml.Unmarshal(rawSchema, &service)

	if err != nil {
		return fmt.Errorf("can't parse schema: %v/n", err)
	}

	typesFileText, err := buildTypesFile(&service)
	if err != nil {
		return err
	}

	fmt.Println(typesFileText)

	handlerInterfaceFileText, err := buildHandlerInterfaceFile(&service)
	if err != nil {
		return err
	}

	fmt.Println(handlerInterfaceFileText)

	executorFileText, err := buildExecutorFile(&service)
	if err != nil {
		return err
	}

	fmt.Println(executorFileText)

	err = ioutil.WriteFile(filepath.Join(outputPath, "types.go"), []byte(typesFileText), 0777)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(outputPath, "handler_interface.go"), []byte(handlerInterfaceFileText), 0777)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(outputPath, "executor.go"), []byte(executorFileText), 0777)
	if err != nil {
		return err
	}

	return nil
}
