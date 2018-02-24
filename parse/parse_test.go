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

package parse

import (
	"testing"
	"fmt"
	"time"
)

// assertValue checks if the returned value is equal to the expected value and
// fails the test if they are not.
func assertValue(t *testing.T, returned interface{}, expected interface{}) {
	if returned != expected {
		t.Fatal(fmt.Sprintf("\nExpected: %v\nReturned: %v", expected, returned))
	}
}

// assertIntSlice checks if the returned value is equal to the expected value and
// fails the test if they are not.
func assertIntSlice(t *testing.T, returned []int, expected []int) {
	for i, value := range returned {
		if value != expected[i]	 {
			t.Fatal(fmt.Sprintf("\nExpected: %v\nReturned: %v", expected, returned))
		}
	}
}

// assertError fails the test if an error was expected and not returned or if
// an error was returned but not expected.
func assertError(t *testing.T, returnedError error, errorExpected bool) {
	if returnedError == nil && errorExpected {
		t.Fatal("error expected but not returned")
	} else if returnedError != nil && !errorExpected {
		t.Fatal(fmt.Sprintf("error returned but not expected\nError: %v", returnedError))
	}
}

func TestReadFileSize(t *testing.T) {
	testCases := []struct {
		TestName string
		Input string
		ExpectedOutput int64
		ErrorExpected bool
	}{
		{"Metric", "10KB", 10000, false},
		{"Binary", "1120KiB", 1146880, false},
		{"Lower-case", "15mb", 15000000, false},
		{"Mixed-case", "2GIb", 2147483648, false},
		{"Zero", "0TB", 0, false},
		{"No unit", "100", 0, true},
		{"No number", "TiB", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			returned, err := ReadFileSize(tc.Input)
			assertValue(t, returned, tc.ExpectedOutput)
			assertError(t, err, tc.ErrorExpected)
		})
	}
}

func TestReadDuration(t *testing.T) {
	testCases := []struct {
		TestName string
		Input string
		ExpectedOutput time.Duration
		ErrorExpected bool
	}{
		{"Single unit", "6m", time.Hour * 4320, false},
		{"Multiple units", "1y2m26d3h", time.Hour * 10707, false},
		{"Upper-case", "3D", time.Hour * 72, false},
		{"No unit", "100", 0, true},
		{"No number", "y", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			returned, err := ReadDuration(tc.Input)
			assertValue(t, returned, tc.ExpectedOutput)
			assertError(t, err, tc.ErrorExpected)
		})
	}
}

func TestReadNumberRanges(t *testing.T) {
	testCases := []struct {
		TestName string
		Input string
		ExpectedOutput []int
		ErrorExpected bool
	}{
		{"Individual numbers", "2,4,8", []int{2, 4, 8}, false},
		{"Ranges", "2-4,16-20", []int{2, 3, 4, 16, 17, 18, 19, 20}, false},
		{"Mixed", "2-4,16", []int{2, 3, 4, 16}, false},
		{"With spaces", " 2 - 4 , 16 ", []int{2, 3, 4, 16}, false},
		{"Empty", "", []int{}, false},
		{"Letters", "a-b,c", []int{}, true},
		{"Negative numbers", "-1-2", []int{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.TestName, func(t *testing.T) {
			returned, err := ReadNumberRanges(tc.Input)
			assertIntSlice(t, returned, tc.ExpectedOutput)
			assertError(t, err, tc.ErrorExpected)
		})
	}
}