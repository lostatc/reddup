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
	"testing"
	"io/ioutil"
)

const excludeTestFileContents = `
# Comment
*.odt
/documents/reports		
  /documents/**/receipt.pdf
books/republic.pdf

	  # Comment
/photos/portrait.???
#/photos/landscape.???
/documents/code.py[cod]
`
var excludeTestPatterns = []string {
"*.odt", "/documents/reports", "/documents/**/receipt.pdf", "books/republic.pdf",
"/photos/portrait.???", "/documents/code.py[cod]",
}

func TestNewExcludeFromFile(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "reddup-")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	tempFile.WriteString(excludeTestFileContents)

	exclude, err := NewExcludeFromFile(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(exclude.Patterns)
	for i, pattern := range exclude.Patterns {
		if pattern != excludeTestPatterns[i] {
			t.Errorf("%v != %v", pattern, excludeTestPatterns[i])
		}
	}
}

func TestCheckMatch(t *testing.T) {
	exclude := Exclude{Patterns: excludeTestPatterns}

	testCases := []struct {
		CheckPath string
		Matches bool
	}{
		{"/dir/foo/thesis.odt", true},
		{"/dir/documents/reports/essay.pdf", true},
		{"/dir/documents/reports/foo/essay.pdf", true},
		{"/dir/foo/documents/reports/essay.pdf", false},
		{"/foo/documents/reports/essay.pdf", false},
		{"/dir/documents/receipt.pdf", true},
		{"/dir/documents/foo/receipt.pdf", true},
		{"/dir/documents/foo.pdf", false},
		{"/dir/books/republic.pdf", true},
		{"/dir/foo/books/republic.pdf", true},
		{"/dir/books/foo/republic.pdf", false},
		{"/dir/photos/portrait.png", true},
		{"/dir/photos/portrait.jpeg", false},
		{"/dir/documents/code.pyc", true},
		{"/dir/documents/code.py", false},
	}

	for _, tc := range testCases {
		result := exclude.CheckMatch(tc.CheckPath, "/dir")

		if result != tc.Matches {
			t.Errorf("Path: %v", tc.CheckPath)
		}
	}
}
