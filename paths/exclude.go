package paths

import (
	"path/filepath"
	"os"
	"bufio"
	"regexp"
	"log"
	"strings"

	"github.com/bmatcuk/doublestar"
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
		var absPattern string

		if strings.HasPrefix(relPattern, string(os.PathSeparator)) {
			absPattern = filepath.Join(startDir, relPattern, "**")
		} else {
			absPattern = filepath.Join("**", relPattern, "**")
		}

		matches, err := doublestar.PathMatch(absPattern, checkPath)
		if err != nil {
			continue
		}

		if matches {
			return true
		}
	}

	return false
}
