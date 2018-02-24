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
	"time"
)

func TestGetDuplicates(t *testing.T) {
	tempPath, teardownFunc := setupFiles(t)
	defer teardownFunc()

	contents := fileContents {
		{"letters/a.txt", "aaa"},
		{"letters/upper/A.txt", "aaa"},
		{"numbers/1.txt", "111"},
	}
	err := writeFiles(contents)
	if err != nil {
		t.Fatal(err)
	}

	pathsToTest, err := NewFilePathsFromRel(testingFilePaths, tempPath)
	if err != nil {
		t.Fatal(err)
	}
	duplicatePaths := GetDuplicates(*pathsToTest)
	expectedPaths := []string{"letters/a.txt", "letters/upper/A.txt"}

	assertPathsEqual(t, duplicatePaths[0], expectedPaths, tempPath)
}

func TestGetOldestDuplicates(t *testing.T) {
	tempPath, teardownFunc := setupFiles(t)
	defer teardownFunc()

	contents := fileContents {
		{"letters/a.txt", "aaa"},
		{"letters/upper/A.txt", "aaa"},
		{"numbers/1.txt", "111"},
	}
	err := writeFiles(contents)
	if err != nil {
		t.Fatal(err)
	}

	// Change the mtime to a time in the past.
	os.Chtimes("letters/upper/A.txt", time.Now(), time.Now().Add(-time.Second))

	pathsToTest, err := NewFilePathsFromRel(testingFilePaths, tempPath)
	if err != nil {
		t.Fatal(err)
	}
	duplicatePaths := GetOldestDuplicates(*pathsToTest)
	expectedPaths := []string{"letters/upper/A.txt"}

	assertPathsEqual(t, duplicatePaths, expectedPaths, tempPath)
}

func TestFilter(t *testing.T) {
	tempPath, teardownFunc := setupFiles(t)
	defer teardownFunc()

	contents := fileContents {
		{"letters/a.txt", "a"},
		{"letters/upper/A.txt", "A"},
		{"numbers/1.txt", "11"},
	}
	err := writeFiles(contents)
	if err != nil {
		t.Fatal(err)
	}
	os.Chtimes("letters/a.txt", time.Now().Add(-time.Second * 3), time.Now())
	os.Chtimes("letters/upper/A.txt", time.Now().Add(-time.Second), time.Now())

	// Because it is limited to 2 bytes, Filter must choose a.txt and A.txt or
	// it must choose 1.txt. Because 1.txt has a more recent atime, it will
	// not choose 1.txt. Because the atime of A.txt is beyond the limit, it
	// will not choose A.txt. This leaves a.txt.
	pathsToTest, err := NewFilePathsFromRel(testingFilePaths, tempPath)
	if err != nil {
		t.Fatal(err)
	}
	filteredPaths := Filter(*pathsToTest, 2, time.Second * 2)
	expectedPaths := []string{"letters/a.txt"}

	assertPathsEqual(t, filteredPaths, expectedPaths, tempPath)
}
