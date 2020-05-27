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
				Usage:   "creates a Shelf",
				Action:  CreateShelf,
			},
			{
				Name:        "track",
				Aliases:     []string{"t"},
				Usage:       "track a file",
				ArgsUsage:   "[shelfname] [filepath] [filename in shelf]",
				Action:      TrackFile,
				Description: "Tracks given file. The file is moved from the path and a symlink is created in it's place.",
			},
			{
				Name:        "clone",
				Aliases:     []string{"cl"},
				Usage:       "clones a shelf",
				ArgsUsage:   "[path to git repo for the shelf]",
				Description: "Clones a shelf from a git clone url",
				Action:      CloneShelf,
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