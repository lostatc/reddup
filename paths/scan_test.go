package paths

import (
	"testing"
)

func TestScanTree(t *testing.T) {
	tempPath, teardownFunc := setupFiles(t)
	defer teardownFunc()

	testCases := []struct {
		TestName string
		Mode FileMode
		ExpectedPaths []string
	}{
		{"Any", ModeAny, testingPaths},
		{"Files and dirs", ModeFile | ModeDir, testingPaths},
		{"Files only", ModeFile, testingFilePaths},
		{"Dirs only", ModeDir, testingDirPaths},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			scannedPaths, err := ScanTree(tempPath, tc.Mode)
			if err != nil {
				t.Fatal(err)
			}

			assertPathsEqual(t, scannedPaths, tc.ExpectedPaths, tempPath)
		})
	}
}
