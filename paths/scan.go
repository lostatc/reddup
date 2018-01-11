package paths

import (
	"os"
	"sync"
	"runtime"
	"path/filepath"

	"github.com/djherbis/times"
)

type FileMode uint

type FilePath struct {
	Path string
	times.Timespec
	os.FileInfo
}

const (
	ModeFile FileMode = 1 << iota
	ModeDir
	ModeLink
	ModeAny = ModeFile | ModeDir | ModeLink
)

// NewFilePath creates a new FilePath struct from a path.
func NewFilePath(path string) (FilePath, error) {
	info, err := os.Stat(path)
	timeInfo := times.Get(info)
	return FilePath{Path: path, Timespec: timeInfo, FileInfo: info}, err
}

// scanDir accepts directory paths on a channel and returns their contents on a
// channel.
func scanDir(jobs chan FilePath, results chan FilePath, wg *sync.WaitGroup) {
	for path := range jobs {
		if !path.IsDir() {
			continue
		}

		file, err := os.Open(path.Path)
		if err != nil {
			continue
		}

		fileInfo, err := file.Readdir(0)
		if err != nil {
			continue
		}

		file.Close()

		for _, info := range fileInfo {
			absolutePath := filepath.Join(path.Path, info.Name())
			timeInfo := times.Get(info)
			filePath := FilePath{Path: absolutePath, Timespec: timeInfo, FileInfo: info}
			results <- filePath
		}
	}
	wg.Done()
}

// scanTrees concurrently scans all given file paths and returns directory
// contents.
func scanTrees(paths []FilePath, workers int) []FilePath {
	jobs := make(chan FilePath, 100)
	results := make(chan FilePath, 100)
	output := make([]FilePath, 0)
	var wg sync.WaitGroup

	// Start workers.
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go scanDir(jobs, results, &wg)
	}

	// Close channel when last worker exits.
	go func(results chan FilePath) {
		wg.Wait()
		close(results)
	}(results)

	// Pass file paths to workers.
	go func(paths []FilePath) {
		for _, path := range paths {
			jobs <- path
		}
		close(jobs)
	}(paths)

	// Retrieve file paths from workers.
	for path := range results {
		output = append(output, path)
	}

	// Scan directories returned from workers.
	if len(output) > 0 {
		output = append(output, scanTrees(output, workers)...)
	}

	return output
}

// ScanTree returns all the file paths in the tree rooted at rootPath and with
// the type specified by mask.
func ScanTree(rootPath string, mode FileMode) ([]FilePath, error) {
	root, err := NewFilePath(rootPath)
	if err != nil {
		return nil, err
	}

	paths := []FilePath{root}
	allPaths := scanTrees(paths, runtime.NumCPU())
	outputPaths := make([]FilePath, 0)

	for _, path := range allPaths {
		if ((mode & ModeFile) == ModeFile) && ((path.Mode() & os.ModeType) == 0) {
			outputPaths = append(outputPaths, path)
		} else if ((mode & ModeDir) == ModeDir) && ((path.Mode() & os.ModeType) == os.ModeDir) {
			outputPaths = append(outputPaths, path)
		} else if ((mode & ModeLink) == ModeLink) && ((path.Mode() & os.ModeType) == os.ModeSymlink) {
			outputPaths = append(outputPaths, path)
		}
	}

	return outputPaths, nil
}
