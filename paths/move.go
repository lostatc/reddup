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
	"os"
	"path/filepath"
	"io"
)

const newDirPerm os.FileMode = 0700

// moveFile moves the file at srcPath to destPath. All necessary directories
// are created. The file mtime and permissions are preserved. If destPath
// already exists, an error is returned.
func moveFile(srcPath, destPath string) (err error) {

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(destPath), newDirPerm)
	destFile, err := os.OpenFile(destPath, os.O_WRONLY | os.O_CREATE | os.O_EXCL, srcInfo.Mode())
	if err != nil {
		return err
	}

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	srcFile.Close()
	destFile.Close()
	os.Chtimes(destPath, srcInfo.ModTime(), srcInfo.ModTime())
	os.Remove(srcPath)

	return nil
}

// MoveStructuredFiles moves the files srcPaths to the directory at destPath
// and preserves their original file structure relative to srcDir. File mtimes
// and permissions are preserved. If a file in destDir already exists, an error
// is returned.
func MoveStructuredFiles(srcDir string, srcPaths FilePaths, destDir string) (err error) {
	for _, srcPath := range srcPaths {
		relPath, err := filepath.Rel(srcDir, srcPath.Path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)
		err = moveFile(srcPath.Path, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}