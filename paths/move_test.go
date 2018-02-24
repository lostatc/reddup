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
	"os"
	"fmt"
	"path/filepath"
)

func TestMoveStructuredFiles(t *testing.T) {
	t.Run("Empty dest", func(t *testing.T) {
		expectedPaths := []string{"letters/upper/A.txt", "numbers/1.txt"}
		srcPath, teardownFunc := setupFiles(t)
		defer teardownFunc()

		destPath, teardownFunc := setupTempDir(t)
		defer teardownFunc()

		pathsToTest, err := NewFilePathsFromRel(expectedPaths, srcPath)
		if err != nil {
			t.Fatal(err)
		}

		err = MoveStructuredFiles(srcPath, *pathsToTest, destPath)
		assertError(t, err, false)

		for _, filePath := range expectedPaths {
			// Check that files don't exist in the source directory.
			if _, err := os.Stat(filepath.Join(srcPath, filePath)); !os.IsNotExist(err) {
				t.Error(fmt.Sprintf("File exists in source directory: %v", filePath))
			}

			// Check that files exist in destination directory.
			if _, err := os.Stat(filepath.Join(destPath, filePath)); os.IsNotExist(err) {
				t.Error(fmt.Sprintf("File missing from destination directory: %v", filePath))
			}
		}
	})

	t.Run("Files in dest", func(t *testing.T) {
		expectedPaths := []string{"letters/upper/A.txt", "numbers/1.txt"}
		srcPath, teardownFunc := setupFiles(t)
		defer teardownFunc()

		destPath, teardownFunc := setupTempDir(t)
		defer teardownFunc()

		// Create a file in the destination directory.
		os.MkdirAll(filepath.Join(destPath, "letters/upper"), newDirPerm)
		file, err := os.Create(filepath.Join(destPath, "letters/upper/A.txt"))
		if err != nil {
			t.Fatal(err)
		}
		file.Close()

		pathsToTest, err := NewFilePathsFromRel(expectedPaths, srcPath)
		if err != nil {
			t.Fatal(err)
		}

		// Trying to move a file into the destination directory which already
		// exists should return an error.
		err = MoveStructuredFiles(srcPath, *pathsToTest, destPath)
		assertError(t, err, true)
	})
}
