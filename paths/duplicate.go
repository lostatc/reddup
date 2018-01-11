package paths

import (
	"os"
	"io"
	"crypto/sha256"
)

type SHA256Sum [32]byte

const BlockSize int = 4096

// GetDuplicates determines which of the given files are identical and returns
// them. It compares files first by size and then by checksum.
func GetDuplicates(paths []FilePath) (duplicates []FilePath) {
	// Get the sizes of each file.
	sizes := make(map[int64][]FilePath)
	sameSizeFiles := make([]FilePath, 0)
	for _, path := range paths {
		sizes[path.Size()] = append(sizes[path.Size()], path)
	}

	// Find files with the same size.
	for _, paths := range sizes {
		if len(paths) > 1 {
			sameSizeFiles = append(sameSizeFiles, paths...)
		}
	}

	// Get the hashes of files with the same size.
	hashes := make(map[SHA256Sum][]FilePath)
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
			duplicates = append(duplicates, paths...)
		}
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
