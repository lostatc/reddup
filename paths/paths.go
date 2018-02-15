package paths

import (
	"os"
	"path/filepath"

	"github.com/djherbis/times"
	"fmt"
	"strings"
	"sort"
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

// String returns the default string representation of the type.
func (f FilePath) String() string {
	return fmt.Sprintf("\"%v\"", f.Path)
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

// NewFilePathsFromRel creates a new FilePaths struct from a set of relative
// file paths and a base path.
func NewFilePathsFromRel(paths []string, base string) (*FilePaths, error) {
	var absPaths []string
	for _, relPath := range paths {
		absPaths = append(absPaths, filepath.Join(base, relPath))
	}
	return NewFilePaths(absPaths)
}

// String returns the default string representation of the type. This satisfies
// the fmt.Stringer interface.
func (f FilePaths) String() string {
	var pathStrings []string
	for _, filePath := range f {
		pathStrings = append(pathStrings, filePath.String())
	}
	return "[" + strings.Join(pathStrings, ", ") + "]"
}

// Len satisfies the sort.Interface interface.
func (f FilePaths) Len() int {
	return len(f)
}

// Swap satisfies the sort.Interface interface.
func (f FilePaths) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// Less satisfies the sort.Interface interface.
func (f FilePaths) Less(i, j int) bool {
	return f[i].Path < f[j].Path
}

// Equals returns true if this slice and other contain all the same paths.
func (f FilePaths) Equals(other FilePaths) bool {
	if len(f) != len(other) {
		return false
	}

	sort.Sort(f)
	sort.Sort(other)

	for i, filePath := range f {
		if filePath.Path != other[i].Path {
			return false
		}
	}

	return true
}

// Difference returns all FilePath objects found in this slice but not in
// other.
func (f FilePaths) Difference(other FilePaths) FilePaths {
	output := make(FilePaths, 0)
	otherMap := make(map[string]struct{})

	for _, filePath := range other {
		otherMap[filePath.Path] = struct{}{}
	}

	for _, filePath := range f {
		if _, ok := otherMap[filePath.Path]; !ok {
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
