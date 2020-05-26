package main

import (
	cli "github.com/urfave/cli/v2" // imports as package "cli"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Action:  CreateShelf,
			},
		},
		Name:        "shelf",
		Description: "A Good Symlinks Manager",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
