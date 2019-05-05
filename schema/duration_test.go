package schema

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

var (
	castDurationSuccessData = []struct {
		desc  string
		value string
		want  time.Duration
	}{
		{"OnlyHour", "P2H", floatToDuration(2, hourNanos)},
		{"2Years", "P2Y", floatToDuration(2, yearNanos)},
		{"SecondsWithDecimal", "P22.519S", floatToDuration(22.519, secondNanos)},
		{"Month Duration", "P1M", floatToDuration(1, monthNanos)},
		{"Week Duration", "P1W", floatToDuration(1, weekNanos)},
		{"Day Duration", "P1D", floatToDuration(1, dayNanos)},
		{"Week Duration with fraction", "P1.5W", floatToDuration(1.5, weekNanos)},
		{"OnlyPeriod", "P3Y6M4D",
			floatToDuration(3, yearNanos) +
				floatToDuration(6, monthNanos) +
				floatToDuration(4, dayNanos)},
		{"OnlyTime", "PT12H30M5S",
			floatToDuration(12, hourNanos) +
				floatToDuration(30, minuteNanos) +
				floatToDuration(5, secondNanos)},
		{"Complex", "P3Y6M4DT12H30M5S",
			floatToDuration(3, yearNanos) +
				floatToDuration(6, monthNanos) +
				floatToDuration(4, dayNanos) +
				floatToDuration(12, hourNanos) +
				floatToDuration(30, minuteNanos) +
				floatToDuration(5, secondNanos)},
		{"Duration", "P1Y2M10DT2H30M",
			floatToDuration(1, yearNanos) +
				floatToDuration(2, monthNanos) +
				floatToDuration(10, dayNanos) +
				floatToDuration(2, hourNanos) +
				floatToDuration(30, minuteNanos)},
		{"Duration with fraction", "P5.1Y3.5M3.4W2.5D",
			floatToDuration(5.1, yearNanos) +
				floatToDuration(3.5, monthNanos) +
				floatToDuration(3.4, weekNanos) +
				floatToDuration(2.5, dayNanos)},
		{"Duration with time with fractions", "P1.5Y2.5M10.5DT2.5H30.5M",
			floatToDuration(1.5, yearNanos) +
				floatToDuration(2.5, monthNanos) +
				floatToDuration(10.5, dayNanos) +
				floatToDuration(2.5, hourNanos) +
				floatToDuration(30.5, minuteNanos)},
	}

	castDurationErrorData = []struct {
		desc  string
		value string
	}{
		{"WrongStartChar", "C2H"},
		{"HourDefaultZero", "PH"},
		{"StringFieldsAreIgnored", "PfooHdddS"},
		{"String NULL", "NULL"},
		{"String NULL_TYPE", "NULL_TYPE"},
		{"String NOT_ALLOWED", "NOT_ALLOWED"},
		{"String NOT_AVAILABLE", "NOT_AVAILABLE"},
		{"String TEMPLATE", "TEMPLATE"},
		{"String TYPE", "TYPE"},
		{"String NO_MATCHING_TYPE", "NO_MATCHING_TYPE"},
		{"Empty P", "P"},
		{"Double P", "PP"},
		{"P with no units", "P1"},
		{"P with unit P", "P1P"},
		{"P with numeric prefix", "1P"},
	}
)

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func BenchmarkCastDuration_Success(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, d := range castDurationSuccessData {
			res, _ := castDuration(d.value)
			_ = res
		}
	}
}

func BenchmarkCastDuration_Error(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, d := range castDurationErrorData {
			res, _ := castDuration(d.value)
			_ = res
		}
	}
}

func TestCastDuration_Success(t *testing.T) {
	for _, d := range castDurationSuccessData {
		t.Run(d.desc, func(t *testing.T) {
			is := is.New(t)
			got, err := castDuration(d.value)
			is.NoErr(err)
			is.Equal(got, d.want)
		})
	}
}

func TestCastDuration_Error(t *testing.T) {
	for _, d := range castDurationErrorData {
		t.Run(d.desc, func(t *testing.T) {
			is := is.New(t)
			_, err := castDuration(d.value)
			is.True(err != nil)
		})
	}
}

func TestUncastDuration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		data := []struct {
			desc  string
			value time.Duration
			want  string
		}{
			{"1Year", 1*hoursInYear + 1*hoursInMonth + 1*hoursInDay + 1*time.Hour + 1*time.Minute + 500*time.Millisecond, "P1Y1M1DT1H1M0.5S"},
		}
		for _, d := range data {
			t.Run(d.desc, func(t *testing.T) {
				is := is.New(t)
				got, err := uncastDuration(d.value)
				is.NoErr(err)
				is.Equal(d.want, got)
			})
		}
	})
	t.Run("Error", func(t *testing.T) {
		data := []struct {
			desc  string
			value interface{}
		}{
			{"InvalidType", 10},
		}
		for _, d := range data {
			t.Run(d.desc, func(t *testing.T) {
				is := is.New(t)
				_, err := uncastDuration(d.value)
				is.True(err != nil)
			})
		}
	})
}
