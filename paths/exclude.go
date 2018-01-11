package paths

import (
	"path/filepath"
	"os"
	"bufio"
	"regexp"
	"log"
)

const CommentPattern string = `^\s*#`

type Exclude struct {
	Patterns []string
}

func NewExcludeFromFile(path string) (*Exclude, error) {
	file, err := os.Open(path)
	if err != nil {
		return &Exclude{}, err
	}
	defer file.Close()

	commentRegex, err := regexp.Compile(CommentPattern)
	if err != nil {
		log.Fatal(err)
	}

	patterns := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := scanner.Text()
		if commentRegex.FindString(pattern) == "" {
			patterns = append(patterns, pattern)
		}
	}

	return &Exclude{Patterns: patterns}, nil
}

// CheckMatch returns true if the given file path matches any pattern relative
// to startDir. Otherwise, it returns false.
func (e *Exclude) CheckMatch(checkPath string, startDir string) (matches bool) {
	for _, relPattern := range e.Patterns {
		absPattern := filepath.Join(startDir, relPattern)
		matches, err := filepath.Match(absPattern, checkPath)
		if err != nil {
			continue
		}
		if matches {
			return true
		}
	}

	return false
}

// Get the paths of files in the directory startPath that match the patterns.
//func (e *Exclude) Match(startPath string) (matches []FilePath) {
//	for _, relPattern := range e.Patterns {
//		absPattern := filepath.Join(startPath, relPattern)
//		paths, err := filepath.Glob(absPattern)
//		if err != nil {
//			continue
//		}
//		for _, path := range paths {
//			newFilePath, err := NewFilePath(path)
//			if err != nil {
//				continue
//			}
//			matches = append(matches, newFilePath)
//		}
//	}
//
//	return matches
//}
