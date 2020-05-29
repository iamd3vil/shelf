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
			{
				Name:        "snapshot",
				Aliases:     []string{"s"},
				Usage:       "creates a snapshot of existing shelves",
				Description: "Provides a way to snapshot your existing shelves to git or an archive file.",
				Subcommands: []*cli.Command{
					{
						Name:      "git",
						Aliases:   []string{"g"},
						Usage:     "Creates an automated commit to check in shelves directory in an existing git repo.",
						ArgsUsage: "[shelfname]",
						Action:    SnapshotGitShelf,
					},
					{
						Name:      "archive",
						Aliases:   []string{"a"},
						Usage:     "creates a snapshot of existing shelves",
						ArgsUsage: "[shelfname]",
						Action:    SnapshotArchiveShelf,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "output, o",
								Usage: "Path to store the archive file.",
							},
						},
					},
				},
			},
			{
				Name:        "restore",
				Aliases:     []string{"r"},
				Usage:       "restores all the links from a shelf",
				ArgsUsage:   "[shelfname]",
				Description: "Restores all the symlinks from the given shelf",
				Action:      RestoreShelf,
			},
			{
				Name:    "where",
				Aliases: []string{"w"},
				Usage:   "prints where the given shelf is",
				Action:  WhereShelf,
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
