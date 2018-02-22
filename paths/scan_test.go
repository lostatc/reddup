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
