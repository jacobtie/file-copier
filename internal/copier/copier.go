package copier

import (
	"fmt"
	"os"
)

type CopyFilesInput struct {
	SourceDir  string
	OutputDir  string
	FileTypes  []string
	NumWorkers int
}

func CopyFiles(input *CopyFilesInput) error {
	if err := ensureDirExists(input.SourceDir); err != nil {
		return fmt.Errorf("error ensuring source dir: %w", err)
	}
	if err := ensureDirExists(input.OutputDir); err != nil {
		return fmt.Errorf("error ensuring output dir: %w", err)
	}
	if err := newExplorer(input).explore(); err != nil {
		return err
	}
	return nil
}

func ensureDirExists(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("dir does not exist: %w", err)
		}
		return fmt.Errorf("error occurred reading dir: %w", err)
	}
	return nil
}
