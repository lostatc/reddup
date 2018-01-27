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

// MoveFiles moves the files srcPaths to the directory at destPath. File mtimes
// and permissions are preserved. If a file in destDir already exists, an error
// is returned.
func MoveFiles(srcPaths FilePaths, destDir string) (err error) {
	for _, srcPath := range srcPaths {
		basePath := filepath.Base(srcPath.Path)
		destPath := filepath.Join(destDir, basePath)
		err := moveFile(srcPath.Path, destPath)
		if err != nil {
			return err
		}
	}
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