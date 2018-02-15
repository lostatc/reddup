package paths

import (
	"testing"
)

func TestScanTree(t *testing.T) {
	tempPath, teardownFunc := setupFiles(t)
	defer teardownFunc()

	testCases := []struct {
		testName string
		mode FileMode
		expectedPaths []string
	}{
		{"Any", ModeAny, testingPaths},
		{"Files and dirs", ModeFile | ModeDir, testingPaths},
		{"Files only", ModeFile, testingFilePaths},
		{"Dirs only", ModeDir, testingDirPaths},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			scannedPaths, err := ScanTree(tempPath, tc.mode)
			if err != nil {
				t.Fatal(err)
			}

			err = assertPathsEqual(scannedPaths, tc.expectedPaths, tempPath)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
