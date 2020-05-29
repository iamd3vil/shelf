package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-git/go-git/v5"
)

const shelvesDIR = ".shelves"

// GetOrCreateShelvesDir gets the directory in user's home folder
// where all shelves are stored. If a directory doesn't exist it creates one.
func GetOrCreateShelvesDir() (string, error) {
	// Get path to shelves directory under $HOME.
	shelfDir := getShelfDirPath()
	_, err := os.Stat(shelfDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Since the shelves directory doesn't exist, create it.
			err = os.Mkdir(shelfDir, 0755)
			if err != nil {
				return "", err
			}
			return shelfDir, nil
		}
		return "", err
	}
	return shelfDir, nil
}

func getShelfDirPath() string {
	var (
		home string
	)
	if runtime.GOOS == "linux" {
		home = os.Getenv("XDG_CONFIG_HOME")
		if home == "" {
			home = os.Getenv("HOME")
		}
	}
	return path.Join(home, shelvesDIR)
}

// SOURCE: https://gist.github.com/mimoo/25fc9716e0f1353791f5908f94d6e726
func compress(src string, buf io.Writer) error {
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)
	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}
		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)
		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	return nil
}

func createGitSnapshot(dir string) error {
	// Opens an already existing repository.
	r, err := git.PlainOpen(dir)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("Error while adding all files: %w", err)
	}
	commitMsg := fmt.Sprintf("snapshot: Automatic commit for snapshot taken at %s", time.Now().Format("Mon Jan _2 15:04:05 2006"))
	_, err = w.Commit(commitMsg, &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("Error while creating a commit: %s", err.Error())
	}
	err = r.Push(&git.PushOptions{})
	if err != nil {
		return fmt.Errorf("Error while pushing the commit: %w", err)
	}
	return nil
}

func createArchiveSnapshot(dir string, output string) error {
	var buf bytes.Buffer
	_ = compress(dir, &buf)
	// write the .tar.gz
	fileToWrite, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("Error while creating output file: %w", err)
	}
	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		return fmt.Errorf("Error while writing data to output: %w", err)
	}
	return nil
}
