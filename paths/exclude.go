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

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && commentRegex.FindString(pattern) == "" {
			patterns = append(patterns, pattern)
		}
	}

	return &Exclude{Patterns: patterns}, nil
}

// CheckMatch returns true if the given file path matches any pattern relative
// to startDir. Otherwise, it returns false.
func (e *Exclude) CheckMatch(checkPath string, startDir string) (matched bool) {
	for _, relPattern := range e.Patterns {
		var absPatterns []string

		// Create one pattern to match the path itself and one pattern to match any paths under it.
		if strings.HasPrefix(relPattern, string(os.PathSeparator)) {
			absPatterns = append(absPatterns, filepath.Join(startDir, relPattern))
			absPatterns = append(absPatterns, filepath.Join(startDir, relPattern, "**"))
		} else {
			absPatterns = append(absPatterns, filepath.Join(startDir, "**", relPattern))
			absPatterns = append(absPatterns, filepath.Join(startDir, "**", relPattern, "**"))
		}

		for _, absPattern := range absPatterns {
			matched, err := doublestar.PathMatch(absPattern, checkPath)
			if err != nil {
				continue
			}

			if matched {
				return true
			}
		}
	}

	return false
}
