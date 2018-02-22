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
	"regexp"
	"log"
	"math"
	"strconv"
	"strings"
	"fmt"
	"time"
)

const sizePattern string = `^(?i)([0-9]+)\s*([KMGTPEZY])(B|iB)?$`
const durationPattern string = `(?i)([0-9]+)\s*([hdmy])`

const (
	durationHour time.Duration = time.Hour
	durationDay = durationHour * 24
	durationMonth = durationDay * 30
	durationYear = durationMonth * 12
)

// This is used by ReadFileSize. The keys in this map must be lower-case.
var sizeReadPrefixes = map[string]int{"k": 1, "m": 2, "g": 3, "t": 4, "p": 5, "e": 6, "z": 7, "y": 8}

// This is used by ReadDuration. The keys in this map must be lower-case.
var durationReadUnits = map[string]time.Duration{"h": durationHour, "d": durationDay, "m": durationMonth, "y": durationYear}

// These are used by FormatFileSize.
var sizeFormatUnits = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
const sizeFormatBase = 1024

// These are used by ReadNumberRanges.
const rangeSeparator = ","
const rangeSpecifier = "-"

// ReadFileSize parses a human-readable file size string and returns a number
// of bytes. It recognizes both metric and binary units.
func ReadFileSize(size string) (numBytes int64, err error) {
	sizeRegex, err := regexp.Compile(sizePattern)
	if err != nil {
		log.Fatal(err)
	}

	match := sizeRegex.FindStringSubmatch(size)
	if len(match) == 0 {
		return 0, fmt.Errorf("the string '%s' is not a valid file size", size)
	}
	num, prefix, unit := match[1], match[2], match[3]
	prefix = strings.ToLower(prefix)
	unit = strings.ToLower(unit)

	var base float64
	switch unit {
	case "", "ib":
		base = 1024
	case "b":
		base = 1000
	}

	exponent := float64(sizeReadPrefixes[prefix])

	coefficient, err := strconv.Atoi(num)
	if err != nil {
		log.Fatal(err)
	}

	return int64(coefficient * int(math.Pow(base, exponent))), nil
}

// FormatFileSize formats a number of bytes
func FormatFileSize(numBytes int64) string {
	if numBytes < sizeFormatBase {
		return fmt.Sprintf("%vB", numBytes)
	}

	output := float32(numBytes)

	prefixCounter := -1
	for output >= sizeFormatBase && prefixCounter < len(sizeFormatUnits) {
		output /= sizeFormatBase
		prefixCounter++
	}

	return fmt.Sprintf("%.1f%s", output, sizeFormatUnits[prefixCounter])
}

// ReadDuration parses a human-readable duration string. It recognizes the
// units "h", "d", "m", and "y".
func ReadDuration(duration string) (time.Duration, error) {
	durationRegex, err := regexp.Compile(durationPattern)
	if err != nil {
		log.Fatal(err)
	}
	matches := durationRegex.FindAllStringSubmatch(duration, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("the string '%s' is not a valid time duration", duration)
	}
	output := time.Duration(0)

	for _, match := range matches {
		value, unit := match[1], match[2]
		num, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
		unit = strings.ToLower(unit)

		output += time.Duration(num) * durationReadUnits[unit]
	}

	return output, nil
}

// ReadNumberRanges parses a comma separated list of number ranges (e.g. "1,7-12,15,47-50").
func ReadNumberRanges(input string) (numbers []int, err error) {
	if strings.TrimSpace(input) == "" {
		return numbers, nil
	}

	ranges := strings.Split(input, rangeSeparator)

	for _, numberRange := range ranges {
		rangeNumbers := strings.Split(numberRange, rangeSpecifier)

		if len(rangeNumbers) > 2 {
			return numbers, fmt.Errorf("number ranges must only contain positive integers")
		}

		start, err := strconv.Atoi(strings.TrimSpace(rangeNumbers[0]))
		if err != nil {
			return numbers, fmt.Errorf("number ranges must only contain positive integers")
		}

		if len(rangeNumbers) == 1 {
			numbers = append(numbers, start)
		} else if len(rangeNumbers) == 2 {
			end, err := strconv.Atoi(strings.TrimSpace(rangeNumbers[1]))
			if err != nil {
				return numbers, fmt.Errorf("number ranges must only contain positive integers")
			}

			for i := start; i <= end; i++ {
				numbers = append(numbers, i)
			}
		}
	}

	return numbers, nil
}
