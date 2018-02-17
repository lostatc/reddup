package parse

import (
	"testing"
	"fmt"
	"time"
)

func assertValue(t *testing.T, returned interface{}, expected interface{}) {
	if returned != expected {
		t.Error(fmt.Sprintf("\nExpected: %v\nReturned: %v", expected, returned))
	}
}

func assertIntSlice(t *testing.T, returned []int, expected []int) {
	for i, value := range returned {
		if value != expected[i]	 {
			t.Error(fmt.Sprintf("\nExpected: %v\nReturned: %v", expected, returned))
		}
	}
}

func assertError(t *testing.T, returnedError error, errorExpected bool) {
	if returnedError == nil && errorExpected {
		t.Error("error expected but not returned")
	} else if returnedError != nil && !errorExpected {
		t.Error(fmt.Sprintf("error returned but not expected\nError: %v", returnedError))
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