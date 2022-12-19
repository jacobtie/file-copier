package copier

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jacobtie/file-copier/internal/concurrency"
	"github.com/jacobtie/file-copier/internal/ds"
	"github.com/jacobtie/file-copier/internal/functional"
)

type explorer struct {
	sourceDir  string
	outputDir  string
	fileTypes  []string
	numWorkers int
	queue      *ds.ThreadSafeQueue[string]
}

func newExplorer(input *CopyFilesInput) *explorer {
	return &explorer{
		sourceDir:  input.SourceDir,
		outputDir:  input.OutputDir,
		fileTypes:  input.FileTypes,
		numWorkers: input.NumWorkers,
		queue:      ds.NewThreadSafeQueue[string](),
	}
}

func (e *explorer) explore() error {
	outputDirPath, err := e.makeOutputDir()
	if err != nil {
		return fmt.Errorf("could not make output dir: %w", err)
	}
	e.outputDir = outputDirPath // overwrite with created subdir
	e.queue.Enqueue(e.sourceDir)
	semaphore := concurrency.NewSemaphore(e.numWorkers)
	var firstError error
	var firstErrorMutex sync.Mutex
	var wg sync.WaitGroup
	numRunning := 0
	var numRunningMutex sync.Mutex
	for !e.queue.IsEmpty() || numRunning > 0 {
		semaphore.Lock()
		if e.queue.IsEmpty() {
			semaphore.Unlock()
			continue
		}
		nextSubDir, err := e.queue.Dequeue()
		if err != nil {
			firstErrorMutex.Lock()
			firstError = fmt.Errorf("an unexpected error occured: %w", err)
			firstErrorMutex.Unlock()
			break
		}
		wg.Add(1)
		numRunningMutex.Lock()
		numRunning++
		numRunningMutex.Unlock()
		go func(subDir string) {
			defer func() {
				wg.Done()
				numRunningMutex.Lock()
				numRunning--
				numRunningMutex.Unlock()
			}()
			if err := e.exploreSubDir(nextSubDir); err != nil {
				firstErrorMutex.Lock()
				firstError = fmt.Errorf("failed to explore subdir %s: %w", nextSubDir, err)
				firstErrorMutex.Unlock()
			}
		}(nextSubDir)
		semaphore.Unlock()
		if firstError != nil {
			break
		}
	}
	wg.Wait()
	if firstError != nil {
		return firstError
	}
	return nil
}

func (e *explorer) makeOutputDir() (string, error) {
	timestamp := time.Now().Unix()
	dirName := fmt.Sprintf("copied_files_%d", timestamp)
	fullOutputDirPath := path.Join(e.outputDir, dirName)
	if err := os.Mkdir(fullOutputDirPath, 0755); err != nil {
		return "", err
	}
	return fullOutputDirPath, nil
}

func (e *explorer) exploreSubDir(subDir string) error {
	subDirs, matchedFiles, err := e.getDirsAndMatchedFiles(subDir)
	if err != nil {
		return fmt.Errorf("error reading initial files in source dir: %w", err)
	}
	if len(matchedFiles) > 0 {
		if err := e.saveFiles(matchedFiles, subDir); err != nil {
			return fmt.Errorf("error saving files: %w", err)
		}
	}
	fullDirNames := functional.Map(subDirs, e.getDirEntryNameFn(subDir))
	e.queue.Enqueue(fullDirNames...)
	return nil
}

func (e *explorer) getDirsAndMatchedFiles(dirName string) ([]os.DirEntry, []os.DirEntry, error) {
	files, err := os.ReadDir(dirName)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read all items in %s directory: %w", dirName, err)
	}
	subDirs := functional.Filter(files, e.isDir)
	matchedFiles := functional.Filter(files, e.isMatchedFile)
	return subDirs, matchedFiles, nil
}

func (*explorer) isDir(dirEntry os.DirEntry) bool {
	return dirEntry.IsDir()
}

func (e *explorer) isMatchedFile(dirEntry os.DirEntry) bool {
	for _, fileType := range e.fileTypes {
		if strings.HasSuffix(dirEntry.Name(), fmt.Sprintf(".%s", fileType)) {
			return true
		}
	}
	return false
}

func (e *explorer) getDirEntryNameFn(currentPath string) func(os.DirEntry) string {
	return func(file os.DirEntry) string {
		return path.Join(currentPath, file.Name())
	}
}

func (e *explorer) saveFiles(files []fs.DirEntry, sourceDirName string) error {
	relativePath := strings.Split(sourceDirName, e.sourceDir)[1]
	outputDir := path.Join(e.outputDir, relativePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not make output dir %s: %w", outputDir, err)
	}
	for _, file := range files {
		sourceFileName := path.Join(sourceDirName, file.Name())
		data, err := os.ReadFile(sourceFileName)
		if err != nil {
			return fmt.Errorf("could not read file %s: %w", sourceFileName, err)
		}
		outputFileName := path.Join(outputDir, file.Name())
		if err := os.WriteFile(outputFileName, data, 0755); err != nil {
			return fmt.Errorf("could not write file %s to %s: %w", sourceFileName, outputFileName, err)
		}
	}
	return nil
}
