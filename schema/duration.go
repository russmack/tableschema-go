package schema

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRegexp = regexp.MustCompile(
	`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+\.?\d*S)?`)

const (
	hoursInYear   = time.Duration(24*365) * time.Hour
	hoursInMonth  = time.Duration(24*30) * time.Hour
	hoursInWeek   = time.Duration(24*7) * time.Hour
	hoursInDay    = time.Duration(24) * time.Hour
	minutesInHour = time.Duration(60) * time.Minute

	yearNanos   = 365.25 * 24 * float64(time.Hour)    // 31557600000000000
	monthNanos  = 30.4375 * 24.0 * float64(time.Hour) // 2629800000000000
	weekNanos   = 7 * 24.0 * float64(time.Hour)       // 604800000000000
	dayNanos    = 24.0 * float64(time.Hour)           // 86400000000000
	hourNanos   = 1 * float64(time.Hour)              // 3600000000000
	minuteNanos = 1 * float64(time.Minute)            // 60000000000
	secondNanos = 1 * float64(time.Second)            // 1000000000

	seps          = "PYMWDTHMS"
	indexOfMonths = 2
	indexOfT      = 5
	delimTime     = 'T'
	unitM         = 'M' // Ambiguous Months and Minutes
	prefix        = 'P'
)

func parseISODateDuration(dur string) ([7]float64, error) {
	// Example: d := "P3Y6M4DT12H30M5S"

	// Matches store.
	m := [7]float64{}
	if dur == "" {
		return m, errors.New("error: empty")
	}
	if len(dur) <= 2 {
		return m, errors.New("error: duration is too short to be valid")
	}
	// Ensure duration starts with 'P'.
	if dur[0] != seps[0] {
		return m, errors.New("error: missing 'P' prefix")
	}
	// Index pointers for duration and separator strings, and matches slice.
	di := 1
	si := 1
	mi := 0
	b := bytes.NewBuffer(make([]byte, 0, 64))
	// Loop over duration, collecting number then character, repeatedly.
	for si < len(seps) {
		if di >= len(dur) {
			break
		}
		// Consume numeric.
		for dur[di] >= '0' && dur[di] <= '9' {
			b.WriteByte(dur[di])
			di++
		}
		if dur[di] == '.' {
			b.WriteByte(dur[di])
			di++
			if dur[di] >= '0' && dur[di] <= '9' {
				for dur[di] >= '0' && dur[di] <= '9' {
					b.WriteByte(dur[di])
					di++
				}
			} else {
				return m, errors.New("error: missing digit after decimal")
			}
		}
		// Consume letter.
		// Iterate over separators (si),
		// looking for a match for the current duration character (di).
		for si < len(seps) {
			if dur[di] == seps[si] {
				// If unit is T skip.
				if dur[di] != delimTime {
					// Distinguish between 'M' months and 'M' minutes.
					if dur[di] == unitM && si > indexOfMonths {
						mi = indexOfT
					}
					f, err := strconv.ParseFloat(b.String(), 64)
					if err != nil {
						return m, err
					}
					m[mi] = f
				}
				b.Reset()
				di++
				si++
				if di < len(dur) && dur[di] != delimTime {
					mi++
				}
				// Matches - break to store next number.
				break
			}
			// Not a match.
			if si == len(seps)-1 {
				return m, errors.New("error: letter not a valid unit")
			}
			si++
			if seps[si] == delimTime {
				continue
			}
			mi++
		}
	}
	if di < len(dur)-1 {
		return m, errors.New("error: invalid ISODate")
	}
	return m, nil
}

func floatToDuration(i, nanos float64) time.Duration {
	var f float64 = i * nanos
	var n int64 = int64(f)
	return time.Duration(n)
}

func castDuration(value string) (time.Duration, error) {
	matches, err := parseISODateDuration(value)
	if err != nil {
		return 0, err
	}
	years := floatToDuration(matches[0], yearNanos)
	months := floatToDuration(matches[1], monthNanos)
	weeks := floatToDuration(matches[2], weekNanos)
	days := floatToDuration(matches[3], dayNanos)
	hours := floatToDuration(matches[4], hourNanos)
	minutes := floatToDuration(matches[5], minuteNanos)
	seconds := floatToDuration(matches[6], secondNanos)

	return years + months + days + weeks + hours + minutes + seconds, nil
}

func castDurationRegex(value string) (time.Duration, error) {
	matches := durationRegexp.FindStringSubmatch(value)
	if len(matches) == 0 {
		return 0, fmt.Errorf("Invalid duration:\"%s\"", value)
	}
	years := parseIntDuration(matches[1], hoursInYear)
	months := parseIntDuration(matches[2], hoursInMonth)
	days := parseIntDuration(matches[3], hoursInDay)
	hours := parseIntDuration(matches[4], time.Hour)
	minutes := parseIntDuration(matches[5], time.Minute)
	seconds := parseSeconds(matches[6])
	return years + months + days + hours + minutes + seconds, nil
}

func parseIntDuration(v string, multiplier time.Duration) time.Duration {
	if len(v) == 0 {
		return 0
	}
	// Ignoring error here because only digits could come from the regular expression.
	d, _ := strconv.Atoi(v[0 : len(v)-1])
	return time.Duration(d) * multiplier
}

func parseSeconds(v string) time.Duration {
	if len(v) == 0 {
		return 0
	}
	// Ignoring error here because only valid arbitrary precision floats could come from the regular expression.
	d, _ := strconv.ParseFloat(v[0:len(v)-1], 64)
	return time.Duration(d * 10e8)
}

func uncastDuration(in interface{}) (string, error) {
	v, ok := in.(time.Duration)
	if !ok {
		return "", fmt.Errorf("invalid duration - value:%v type:%v", in, reflect.ValueOf(in).Type())
	}
	y := v / hoursInYear
	r := v % hoursInYear
	m := r / hoursInMonth
	r = r % hoursInMonth
	d := r / hoursInDay
	r = r % hoursInDay
	return strings.ToUpper(fmt.Sprintf("P%dY%dM%dDT%s", y, m, d, r.String())), nil
}
