/*
Copyright © 2018 Garrett Powell <garrett@gpowell.net>

This file is part of reddup.

reddup is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

reddup is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with reddup.  If not, see <http://www.gnu.org/licenses/>.
*/

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

// setupTempDir creates a temporary directory.
func setupTempDir(t *testing.T) (tempPath string, teardownFunc func()) {
	tempPath, err := ioutil.TempDir("", "reddup-")
	if err != nil {
		t.Error(err)
	}

	return tempPath, func() {
		os.RemoveAll(tempPath)
	}
}

// setupFiles creates a set of test files in a temporary directory.
func setupFiles(t *testing.T) (tempPath string, teardownFunc func()) {
	tempPath, teardownFunc = setupTempDir(t)
	os.Chdir(tempPath)

	for _, dirPath := range testingDirPaths {
		absPath := filepath.Join(tempPath, dirPath)
		os.Mkdir(absPath, 0700)
	}

	for _, filePath := range testingFilePaths {
		absPath := filepath.Join(tempPath, filePath)
		file, err := os.Create(absPath)
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}

	return tempPath, teardownFunc
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
// expected file paths and fails the test if they are not. The expected file
// paths are paths relative to startPath.
func assertPathsEqual(t *testing.T, returned FilePaths, expected []string, startPath string) {
	expectedFilePaths, err := NewFilePathsFromRel(expected, startPath)
	if err != nil {
		t.Fatal(err)
	}

	if !returned.Equals(*expectedFilePaths) {
		t.Fatal(fmt.Sprintf(
			"\nReturned but not expected: %v\nExpected but not returned: %v",
			returned.Difference(*expectedFilePaths),
			expectedFilePaths.Difference(returned)))
	}
}

// assertError fails the test if an error was expected and not returned or if
// an error was returned but not expected.
func assertError(t *testing.T, returnedError error, errorExpected bool) {
	if returnedError == nil && errorExpected {
		t.Fatal("error expected but not returned")
	} else if returnedError != nil && !errorExpected {
		t.Fatal(fmt.Sprintf("error returned but not expected\nError: %v", returnedError))
	}
}
