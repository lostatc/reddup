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

	err = assertPathsEqual(duplicatePaths[0], expectedPaths, tempPath)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNewestDuplicates(t *testing.T) {
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
	os.Chtimes("letters/upper/A.txt", time.Now(), time.Now().Add(-time.Second))

	pathsToTest, err := NewFilePathsFromRel(testingFilePaths, tempPath)
	if err != nil {
		t.Fatal(err)
	}
	duplicatePaths := GetNewestDuplicates(*pathsToTest)
	expectedPaths := []string{"letters/a.txt"}

	err = assertPathsEqual(duplicatePaths, expectedPaths, tempPath)
	if err != nil {
		t.Error(err)
	}
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

	err = assertPathsEqual(filteredPaths, expectedPaths, tempPath)
	if err != nil {
		t.Error(err)
	}
}
