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
	// Get path to shelves directory.
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().First()

	if shelfName == "" {
		return errors.New("shelf name has to be given")
	}

	shelfPath := path.Join(shelfDir, shelfName)

	err = os.Mkdir(shelfPath, 0755)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("shelf named: \"%s\" already exist", shelfName)
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
	// Get path to shelves directory.
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().Get(0)
	if shelfName == "" {
		return errors.New("shelf name has to be given")
	}

	filePath := cliCtx.Args().Get(1)
	if filePath == "" {
		return errors.New("file path to track can't be blank")
	}

	// Check if the given shelf exists
	_, err = os.Stat(path.Join(shelfDir, shelfName))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("shelf named: %s doesn't exist", shelfName)
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
	_, err = os.Stat(path.Join(shelfDir, shelfName, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			goto Moving
		}
		return err
	}
	return fmt.Errorf("file with name %s already exists in the shelf. Please mention the filename to used for this file in the shelf", fileName)

Moving:
	// Move file to the shelf
	err = os.Rename(filePath, path.Join(shelfDir, shelfName, fileName))
	if err != nil {
		return err
	}

	fmt.Printf("[*] Moved file at %s to %s\n", filePath, path.Join(shelfDir, shelfName, fileName))

	// Create symlink
	err = os.Symlink(path.Join(shelfDir, shelfName, fileName), filePath)
	if err != nil {
		// Since we can't create a symlink, we should put back the file which is moved
		err = os.Rename(path.Join(shelfDir, shelfName, fileName), filePath)
		if err != nil {
			return err
		}
		return err
	}

	// Put it in the db
	db, dbPath, err := GetDB(path.Join(shelfDir, shelfName))
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
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}
	err = os.Chdir(shelfDir)
	if err != nil {
		return err
	}

	url := cliCtx.Args().First()
	if url == "" {
		return errors.New("git repo url for the shelf has to be provided")
	}

	fmt.Printf("[*] Cloning from %s\n", url)

	cmd := exec.Command("git", "clone", url)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// SnapshotGitShelf prepares a backup of existing shelf using `.git`
func SnapshotGitShelf(cliCtx *cli.Context) error {
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}
	shelfName := cliCtx.Args().First()
	if shelfName == "" {
		return errors.New("shelf name can't be empty")
	}
	directory := path.Join(shelfDir, shelfName)
	err = createGitSnapshot(directory)
	if err != nil {
		return fmt.Errorf("error while creating a snapshot with git: %w", err)
	}
	return nil
}

// SnapshotArchiveShelf prepares a backup of existing shelf using a compressed archive file.
func SnapshotArchiveShelf(cliCtx *cli.Context) error {
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}
	shelfName := cliCtx.Args().First()
	outputDir := cliCtx.String("output")
	if shelfName == "" {
		return errors.New("shelf name can't be empty")
	}
	if outputDir == "" {
		return errors.New("output path can't be empty")
	}
	directory := path.Join(shelfDir, shelfName)
	outputPath := path.Join(outputDir, fmt.Sprintf("%s.tar.gz", shelfName))
	err = createArchiveSnapshot(directory, outputPath)
	if err != nil {
		return fmt.Errorf("error while creating a snapshot with archive: %w", err)
	}
	return nil
}

// RestoreShelf restores all the symlinks from the given shelf
func RestoreShelf(cliCtx *cli.Context) error {
	shelfDir, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}

	shelfName := cliCtx.Args().First()
	if shelfName == "" {
		return errors.New("shelf name can't be empty")
	}

	shelfPath := path.Join(shelfDir, shelfName)

	// Check if the given shelf exists
	_, err = os.Stat(shelfPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("shelf named: %s doesn't exist", shelfName)
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

		// If the symlink path in db is absolute, it will be put in home directory
		if !path.IsAbs(lPath) {
			fmt.Printf("lPath is abs: ")
			home := getHomeDir()
			lPath = path.Join(home, lPath)
		}

		err = os.MkdirAll(path.Dir(lPath), 0755)
		if err != nil {
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
	home, err := GetOrCreateShelvesDir()
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

func GetListOfFilesInShelf(cliCtx *cli.Context) error {
	home, err := GetOrCreateShelvesDir()
	if err != nil {
		return err
	}
	shelfName := cliCtx.Args().First()
	if shelfName == "" {
		return errors.New("shelf name can't be empty")
	}

	shelfPath := path.Join(home, shelfName)

	db, _, err := GetDB(shelfPath)
	if err != nil {
		return err
	}
	fmt.Printf("List of files tracked in shelf %s are:\n", shelfName)
	links := db.GetLinks()
	for _, v := range links {
		fmt.Println(v)
	}
	return nil
}
