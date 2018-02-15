package paths

import (
	"testing"
	"io/ioutil"
	"path/filepath"
	"os"
	"fmt"
)

type fileContents []struct {
	Path string
	Contents string
}

var testingDirPaths = []string {
	"empty", "letters", "letters/upper", "numbers",
}

var testingFilePaths = []string {
	"letters/a.txt", "letters/upper/A.txt", "numbers/1.txt",
}

var testingPaths = append(testingDirPaths, testingFilePaths...)

// setupFiles creates a set of test files in a temporary directory.
func setupFiles(t *testing.T) (tempPath string, teardownFunc func()) {
	tempPath, err := ioutil.TempDir("", "reddup-")
	if err != nil {
		t.Error(err)
	}
	os.Chdir(tempPath)

	for _, dirPath := range testingDirPaths {
		absPath := filepath.Join(tempPath, dirPath)
		os.Mkdir(absPath, 0700)
	}

	for _, filePath := range testingFilePaths {
		absPath := filepath.Join(tempPath, filePath)
		file, err := os.Create(absPath)
		if err != nil {
			t.Error(err)
		}
		file.Close()
	}

	return tempPath, func() {
		os.RemoveAll(tempPath)
	}
}

// writeFiles writes the given strings to their corresponding files.
func writeFiles(contents fileContents) error {
	for _, fc := range contents {
		file, err := os.Create(fc.Path)
		if err != nil {
			return err
		}

		_, err = file.WriteString(fc.Contents)
		if err != nil {
			return err
		}

		file.Close()
	}

	return nil
}

// assertPathsEqual checks that the returned file paths are the same as the
// expected file paths and returns an error if they are not. The expected file
// paths are paths relative to startPath.
func assertPathsEqual(returned FilePaths, expected []string, startPath string) (err error) {
	expectedFilePaths, err := NewFilePathsFromRel(expected, startPath)
	if err != nil {
		return err
	}

	if !returned.Equals(*expectedFilePaths) {
		return fmt.Errorf(
			"\nReturned but not expected: %v\nExpected but not returned: %v",
			returned.Difference(*expectedFilePaths),
			expectedFilePaths.Difference(returned))
	}

	return nil
}
