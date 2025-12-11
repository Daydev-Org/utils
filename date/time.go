package date

import "time"

func NowUnix() int64 { return time.Now().Unix() }

func AddTime(delta time.Duration) time.Time {
	return time.Now().Add(delta)
}
