package input

import (
	"regexp"
	"log"
	"math"
	"strconv"
	"strings"
	"fmt"
	"time"
)

const sizePattern string = `^(?i)([0-9]+)\s*([KMG])(B|iB)?$`
const durationPattern string = `(?i)([0-9]+)\s*([hdmy])`
const (
	Hour time.Duration = time.Hour
	Day time.Duration = Hour * 24
	Month time.Duration = Day * 30
	Year time.Duration = Month * 12
)

func FileSize(size string) (numBytes int, err error) {
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
	var exponent float64

	switch unit {
	case "", "ib":
		base = 1024
	case "b":
		base = 1000
	}

	switch prefix {
	case "k":
		exponent = 1
	case "m":
		exponent = 2
	case "g":
		exponent = 3
	}

	coefficient, err := strconv.Atoi(num)
	if err != nil {
		log.Fatal(err)
	}

	return coefficient * int(math.Pow(base, exponent)), nil
}

func Duration(duration string) (time.Duration, error) {
	durationRegex, err := regexp.Compile(durationPattern)
	if err != nil {
		log.Fatal(err)
	}
	matches := durationRegex.FindAllStringSubmatch(duration, -1)
	output := time.Duration(0)

	for _, match := range matches {
		value, unit := match[1], match[2]
		num, err := strconv.Atoi(value)
		if err != nil {
			log.Fatal(err)
		}
		unit = strings.ToLower(unit)

		switch unit {
		case "h":
			output += time.Duration(num) * Hour
		case "d":
			output += time.Duration(num) * Day
		case "m":
			output += time.Duration(num) * Month
		case "y":
			output += time.Duration(num) * Year
		}
	}

	return output, nil
}
