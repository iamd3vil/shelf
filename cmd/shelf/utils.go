package main

import (
	"os"
	"path"
	"runtime"
)

// GetHomeDirectory gets the directory in user's home folder
// where all shelves are stored. If a directory doesn't exist it creates one.
func GetHomeDirectory() (string, error) {
	// Get Home Directory path
	home := getHomeDirectoryPath()
	_, err := os.Stat(home)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(home, 0755)
			if err != nil {
				return "", err
			}
			return home, nil
		}
		return "", err
	}

	return home, nil
}

func getHomeDirectoryPath() string {
	var (
		home string
	)
	if runtime.GOOS == "linux" {
		home = os.Getenv("XDG_CONFIG_HOME")
		if home == "" {
			home = os.Getenv("HOME")
		}
	}
	return path.Join(home, ".shelves")
}
