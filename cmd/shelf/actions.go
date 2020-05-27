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

	// Create a shelf db
	_, err = NewDB(shelfPath, shelfName)
	if err != nil {
		return err
	}

	fmt.Printf("[*] Created a shelf named: %s\n", shelfName)

	return nil
}

// TrackFile tracks the given file.
// The file is moved to the shelf and a symlink is created in its place.
// It stores the filename and symlink path in the shelf's db.
func TrackFile(cliCtx *cli.Context) error {
	// Get Home Directory path
	home, err := GetHomeDirectory()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().Get(0)
	if shelfName == "" {
		return errors.New("Shelf name has to be given")
	}

	filePath := cliCtx.Args().Get(1)
	if filePath == "" {
		return errors.New("File path to track can't be blank")
	}

	// Check if the given shelf exists
	_, err = os.Stat(path.Join(home, shelfName))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Shelf named: %s doesn't exist", shelfName)
		}
		return err
	}

	// Check if the file exists and is not a symlink
	stat, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s doesn't exist", filePath)
		}
		return err
	}
	if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
		return fmt.Errorf("%s shouldn't be a symlink", filePath)
	}

	fileName := path.Base(filePath)
	// If the filename is given as an argument, save the file with that name
	if cliCtx.Args().Get(2) != "" {
		fileName = cliCtx.Args().Get(2)
	}

	// Check if there's already a file in the shelf with this fileName
	_, err = os.Stat(path.Join(home, shelfName, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			goto Moving
		}
		return err
	}
	return fmt.Errorf("file with name %s already exists in the shelf. Please mention the filename to used for this file in the shelf", fileName)

Moving:
	// Move file to the shelf
	err = os.Rename(filePath, path.Join(home, shelfName, fileName))
	if err != nil {
		return err
	}

	fmt.Printf("[*] Moved file at %s to %s\n", filePath, path.Join(home, shelfName, fileName))

	// Create symlink
	err = os.Symlink(path.Join(home, shelfName, path.Base(filePath)), filePath)
	if err != nil {
		// Since we can't create a symlink, we should put back the file which is moved
		err = os.Rename(path.Join(home, shelfName, path.Base(filePath)), filePath)
		if err != nil {
			return err
		}
		return err
	}

	// Put it in the db
	db, dbPath, err := GetDB(path.Join(home, shelfName))
	if err != nil {
		return err
	}
	db.AddLink(fileName, filePath)
	f, err := os.Create(dbPath)
	if err != nil {
		return err
	}
	err = db.Marshal(f)
	if err != nil {
		return err
	}

	return nil
}

// CloneShelf clones the shelf from the given git repo url
func CloneShelf(cliCtx *cli.Context) error {
	home, err := GetHomeDirectory()
	if err != nil {
		return err
	}
	err = os.Chdir(home)
	if err != nil {
		return err
	}

	url := cliCtx.Args().First()
	if url == "" {
		return errors.New("Git repo url for the shelf has to be provided")
	}

	fmt.Printf("[*] Cloning from %s\n", url)

	cmd := exec.Command("git", "clone", url)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// RestoreShelf restores all the symlinks from the given shelf
func RestoreShelf(cliCtx *cli.Context) error {
	home, err := GetHomeDirectory()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().First()
	if shelfName == "" {
		return errors.New("Shelf name can't be empty")
	}

	shelfPath := path.Join(home, shelfName)

	// Check if the given shelf exists
	_, err = os.Stat(shelfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Shelf named: %s doesn't exist", shelfName)
		}
		return err
	}

	// Read the db
	db, _, err := GetDB(shelfPath)
	if err != nil {
		return err
	}

	// Loop over each link and put a symlink
	for fName, lPath := range db.Links {
		// Check if there is a file
		// If there is no file with the file name in the shelf, skip over it
		_, err := os.Stat(path.Join(shelfPath, fName))
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("[*] Warning: File missing in the shelf: %s. Skipping...\n", fName)
				continue
			}
			return err
		}

		err = os.Symlink(path.Join(shelfPath, fName), lPath)
		if err != nil {
			if os.IsExist(err) {
				fmt.Printf("[*] Warning: There is a already a file at %s. Skipping restoring: %s\n", lPath, fName)
				continue
			}
			return err
		}
	}

	fmt.Printf("[*] Restored %s shelf\n", shelfName)

	return nil
}

// WhereShelf changes the directory to given shelf's directory
func WhereShelf(cliCtx *cli.Context) error {
	home, err := GetHomeDirectory()
	if err != nil {
		return err
	}
	shelfName := cliCtx.Args().First()
	if shelfName == "" {
		return errors.New("shelf name can't be empty")
	}

	shelfPath := path.Join(home, shelfName)

	// Check if shelf exists
	_, err = os.Stat(shelfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("shelf named: %s doesn't exists", shelfName)
		}
		return err
	}

	fmt.Println(shelfPath)
	return nil
}
