package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/urfave/cli/v2"
)

// CreateShelf creates a new shelf
func CreateShelf(cliCtx *cli.Context) error {
	// Get Home Directory path
	home, err := GetHomeDirectory()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().First()

	if shelfName == "" {
		return errors.New("Shelf name has to be given")
	}

	shelfPath := path.Join(home, shelfName)

	err = os.Mkdir(shelfPath, 0755)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("Shelf named: \"%s\" already exist", shelfName)
		}
		return err
	}

	err = os.Chdir(shelfPath)
	if err != nil {
		return err
	}

	// Initialize git in the shelf
	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("[*] Created a shelf named: %s", shelfName)

	return nil
}
