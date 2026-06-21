package runtime

import (
	"time"
)

func TimeNow() time.Time {
	return time.Now()
}

func TimeSleep(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}
