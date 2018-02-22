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
