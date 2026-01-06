package utils

import "time"

func ConvertIntToDuration(value int64) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func ConvertDurationToInt(duration time.Duration) int64 {
	return int64(duration / time.Millisecond)
}
