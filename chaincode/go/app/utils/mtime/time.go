package mtime

import (
	"strconv"
	"time"
)

var (
	formatIntDate    = "20060102"
	formatIntFull    = "20060102150405"
	formatStringFull = "2006-01-02 15:04:05"
)

// example 20180102
func TodayDateInt() int {
	str := time.Now().Format(formatIntDate)
	tm, _ := strconv.Atoi(str)
	return tm
}

func Now() string {
	return time.Now().Format(formatStringFull)
}
