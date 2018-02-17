package paths

import (
	"sort"
	"math"
	"time"
	"os"
	"crypto/sha256"
	"io"
)

type SHA256Sum [32]byte

type FilePriority struct {
	File FilePath
	Priority float64
}

const BlockSize int = 4096

// prioritizePaths sorts paths based on their size and atime. Paths with a
// larger size and less recent atime are sorted first.
func prioritize(paths FilePaths) (sorted FilePaths) {
	// Get a priority for each file path based on the size and atime.
	priorities := make([]FilePriority, 0)
	for _, path := range paths {
		size := path.Stat.Size()
		var priority float64
		if size == 0 {
			priority = math.Inf(1)
		} else {
			priority = float64(path.Time.AccessTime().Unix() / size)
		}
		priorities = append(priorities, FilePriority{File: path, Priority: priority})
	}

	// Sort by path and then by priority so that the output for a given input
	// is always the same.
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].File.Path < priorities[j].File.Path
	})
	sort.SliceStable(priorities, func(i, j int) bool {
		return priorities[i].Priority < priorities[j].Priority
	})

	for _, path := range priorities {
		sorted = append(sorted, path.File)
	}

	return sorted
}

// Filter returns the files with the largest size and most recent atime that
// fit within totalSize and were last accessed at least minDuration in the past.
func Filter(paths FilePaths, totalSize int64, minDuration time.Duration) FilePaths {
	sortedPaths :=  prioritize(paths)
	remainingSpace := int64(totalSize)
	output := make(FilePaths, 0)
	maxAtime := time.Now().Add(-minDuration)

	for _, path := range sortedPaths {
		if path.Stat.Size() == 0 || path.Time.AccessTime().After(maxAtime) {
			continue
		}

		newRemainingSpace := remainingSpace - path.Stat.Size()
		if newRemainingSpace >= 0 {
			output = append(output, path)
			remainingSpace = newRemainingSpace
		}
	}

	return output
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

// GetDuplicates determines which of the given files are identical and returns
// them. Each FilePaths slice in the slice that is returned represents a group
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

// DuplicateFilter is equivalent to Filter except duplicate files are
// prioritized first. Duplicate files are sorted in descending order of their
// size.
func DuplicateFilter(paths FilePaths, totalSize int64, minDuration time.Duration) FilePaths {
	var selectedPaths FilePaths

	duplicatePaths := GetNewestDuplicates(paths)
	for _, filePath := range duplicatePaths {
		filePath.Flags |= FlagDuplicate
		selectedPaths = append(selectedPaths, filePath)
	}

	maxSize := totalSize - duplicatePaths.TotalSize()
	selectedPaths = append(selectedPaths, Filter(paths.Difference(duplicatePaths), maxSize, minDuration)...)

	return selectedPaths
}
