package main

import (
	"bitbucket.org/timeio/go-service/lib"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.2"

	app.Commands = []cli.Command{
		{
			Name:   "build",
			Usage:  "build migrations",
			Action: build,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func build(c *cli.Context) error {
	args := c.Args()

	filePath := args.Get(0)
	if filePath == "" {
		return errors.New("file path is required")
	}

	outputPath := args.Get(1)
	if outputPath == "" {
		return errors.New("output path is required")
	}

	return lib.Build(filePath, outputPath)
}
