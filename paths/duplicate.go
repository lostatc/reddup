package paths

import (
	"os"
	"io"
	"crypto/sha256"
)

type SHA256Sum [32]byte

const BlockSize int = 4096

// GetDuplicates determines which of the given files are identical and returns
// them. Each FilePaths object in the slice that is returned represents a group
// of identical files. Files are compared first by size and then by checksum.
func GetDuplicates(paths FilePaths) (duplicates []FilePaths) {
	// Get the sizes of each file.
	sizes := make(map[int64]FilePaths)
	sameSizeFiles := make(FilePaths, 0)
	for _, path := range paths {
		sizes[path.Stat.Size()] = append(sizes[path.Stat.Size()], path)
	}

	// Find files with the same size.
	for _, paths := range sizes {
		if len(paths) > 1 {
			sameSizeFiles = append(sameSizeFiles, paths...)
		}
	}

	// Get the hashes of files with the same size.
	hashes := make(map[SHA256Sum]FilePaths)
	for _, path := range sameSizeFiles {
		sum, err := checksum(path.Path)
		if err != nil {
			continue
		}
		hashes[sum] = append(hashes[sum], path)
	}

	// Find files with the same hash.
	for _, paths := range hashes {
		if len(paths) > 1 {
			duplicates = append(duplicates, paths)
		}
	}

	return duplicates
}

// GetNewestDuplicates returns the file with the most recent mtime for each
// group of duplicate files from the input.
func GetNewestDuplicates(paths FilePaths) (duplicates FilePaths) {
	allDuplicates := GetDuplicates(paths)
	for _, group := range allDuplicates {
		newestPath := group[0]

		for _, filePath := range group {
			if filePath.Stat.ModTime().After(newestPath.Stat.ModTime()) {
				newestPath = filePath
			}
		}

		duplicates = append(duplicates, newestPath)
	}

	return duplicates
}

// checksum returns the SHA256 sum of a given file.
func checksum(path string) (checksum SHA256Sum, err error) {
	file, err := os.Open(path)
	if err != nil {
		return checksum, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return checksum, err
	}

	// Convert the slice to a fixed-size array.
	copy(checksum[:], hash.Sum(nil))
	return checksum, nil
}
