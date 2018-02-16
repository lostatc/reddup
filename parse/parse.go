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

// The keys in these maps must be lower-case.
var sizeReadPrefixes = map[string]int{"k": 1, "m": 2, "g": 3, "t": 4, "p": 5, "e": 6, "z": 7, "y": 8}
var durationReadUnits = map[string]time.Duration{"h": durationHour, "d": durationDay, "m": durationMonth, "y": durationYear}

var sizeFormatUnits = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
const sizeFormatBase = 1024

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
