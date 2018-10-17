package main

import (
	"fmt"
	"strings"
	"time"
)

//time format const (
//    ANSIC       = "Mon Jan _2 15:04:05 2006"
//    UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
//    RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
//    RFC822      = "02 Jan 06 15:04 MST"
//    RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
//    RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
//    RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
//    RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
//    RFC3339     = "2006-01-02T15:04:05Z07:00"
//    RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
//    Kitchen     = "3:04PM"
//    // Handy time stamps.
//    Stamp      = "Jan _2 15:04:05"
//    StampMilli = "Jan _2 15:04:05.000"
//    StampMicro = "Jan _2 15:04:05.000000"
//    StampNano  = "Jan _2 15:04:05.000000000"
//)
func GoStdTime() string {
	return "2006-01-02 15:04:05"
}

func GoStdUnixDate() string {
	return "Mon Jan _2 15:04:05 MST 2006"
}

func GoStdRubyDate() string {
	return "Mon Jan 02 15:04:05 -0700 2006"
}

func GetTmStr(tm time.Time, format string) string {
	patterns := []string{
		"y", "2006",
		"m", "01",
		"d", "02",

		"Y", "2006",
		"M", "01",
		"D", "02",

		"h", "03", //12小时制
		"H", "15", //24小时制

		"i", "04",
		"s", "05",

		"t", "pm",
		"T", "PM",
	}
	return convStr(tm, format, patterns)
}

func GetTmShortStr(tm time.Time, format string) string {
	patterns := []string{
		"y", "06",
		"m", "1",
		"d", "2",

		"Y", "06",
		"M", "1",
		"D", "2",

		"h", "3", //12小时制
		"H", "15", //24小时制

		"i", "4",
		"s", "5",

		"t", "pm",
		"T", "PM",
	}

	return convStr(tm, format, patterns)
}

func convStr(tm time.Time, format string, patterns []string) string {
	replacer := strings.NewReplacer(patterns...)
	strfmt := replacer.Replace(format)
	return tm.Format(strfmt)
}

func GetLocaltimeStr() string {
	now := time.Now().Local()
	year, mon, day := now.Date()
	hour, min, sec := now.Clock()
	zone, _ := now.Zone()
	return fmt.Sprintf("%d-%d-%d %02d:%02d:%02d %s", year, mon, day, hour, min, sec, zone)
}

func GetGmtimeStr() string {
	now := time.Now()
	year, mon, day := now.UTC().Date()
	hour, min, sec := now.UTC().Clock()
	zone, _ := now.UTC().Zone()
	return fmt.Sprintf("%d-%d-%d %02d:%02d:%02d %s", year, mon, day, hour, min, sec, zone)
}

func GetRFC3339TimeStr(tm time.Time) string {
	return tm.Format(time.RFC3339)
}

func ParseRFC3339TimeStr(tmStr string) (time.Time, error) {
	t1, e := time.Parse(time.RFC3339, tmStr)
	return t1, e
}

// MakeTimestamp 获取当前时间戳，毫秒
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetUnixTimeStr(ut int64, format string) string {
	t := time.Unix(ut, 0)
	return GetTmStr(t, format)
}

func GetUnixTimeShortStr(ut int64, format string) string {
	t := time.Unix(ut, 0)
	return GetTmShortStr(t, format)
}

func Greatest(arr []time.Time) time.Time {
	var temp time.Time
	for _, at := range arr {
		if temp.Before(at) {
			temp = at
		}
	}
	return temp
}

type TimeSlice []time.Time

func (s TimeSlice) Len() int {
	return len(s)
}

func (s TimeSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s TimeSlice) Less(i, j int) bool {
	if s[i].IsZero() {
		return false
	}
	if s[j].IsZero() {
		return true
	}
	return s[i].Before(s[j])
}
