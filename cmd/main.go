package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jacobtie/file-copier/internal/copier"
)

func main() {
	sourceDir, outputDir, fileTypesRaw, numWorkers, err := parseFlags()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := copier.CopyFiles(&copier.CopyFilesInput{
		SourceDir:  sourceDir,
		OutputDir:  outputDir,
		FileTypes:  strings.Split(fileTypesRaw, ","),
		NumWorkers: numWorkers,
	}); err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Completed successfully")
}

func parseFlags() (string, string, string, int, error) {
	sourceDir := flag.String("sourcedir", "", "source directory to crawl")
	outputDir := flag.String("outputdir", "", "output directory to crawl")
	fileTypes := flag.String("filetypes", "", "filetypes to copy")
	numWorkers := flag.Int("workers", 1, "number of workers (1-20, default 1)")
	flag.Parse()
	if sourceDir == nil || *sourceDir == "" {
		return "", "", "", 0, fmt.Errorf("missing sourcedir")
	}
	if outputDir == nil || *outputDir == "" {
		return "", "", "", 0, fmt.Errorf("missing outputdir")
	}
	if fileTypes == nil || *fileTypes == "" {
		return "", "", "", 0, fmt.Errorf("missing filetypes")
	}
	if numWorkers == nil || *numWorkers < 1 || *numWorkers > 20 {
		return "", "", "", 0, fmt.Errorf("invalid workers")
	}
	return *sourceDir, *outputDir, *fileTypes, *numWorkers, nil
}
