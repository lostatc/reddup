package paths

import (
	"sort"
	"math"
	"time"
)

type FilePriority struct {
	File FilePath
	Priority float64
}

// prioritizePaths sorts paths based on their size and atime.
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

// Filter returns the lowest-priority file paths that fit within totalSize and
// were last accessed at least minDuration in the past.
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
