package paths

import (
	"os"

	"github.com/djherbis/times"
)

type FilePath struct {
	Path string
	Time times.Timespec
	Stat os.FileInfo
}

// NewFilePath creates a new FilePath struct from a path.
func NewFilePath(path string) (*FilePath, error) {
	info, err := os.Stat(path)
	if err != nil {
		return new(FilePath), err
	}
	timeInfo := times.Get(info)
	return &FilePath{Path: path, Time: timeInfo, Stat: info}, nil
}

type FilePaths []FilePath

// NewFilePaths creates a new FilePaths struct from file paths.
func NewFilePaths(paths []string) (*FilePaths, error) {
	var output FilePaths
	for _, filePath := range paths {
		newFilePath, err := NewFilePath(filePath)
		if err != nil {
			return nil, err
		}
		output = append(output, *newFilePath)
	}
	return &output, nil
}

// Difference returns all FilePath objects found in this slice but not in
// other.
func (f FilePaths) Difference(other FilePaths) FilePaths {
	output := make(FilePaths, 0)
	otherMap := make(map[FilePath]struct{})

	for _, filePath := range other {
		otherMap[filePath] = struct{}{}
	}

	for _, filePath := range f {
		if _, ok := otherMap[filePath]; !ok {
			output = append(output, filePath)
		}
	}

	return output
}

// TotalSize returns the total size of all the FilePath objects in bytes.
func (f FilePaths) TotalSize() (size int64) {
	for _, filePath := range f {
		size += filePath.Stat.Size()
	}

	return size
}
