package common

import (
	"fmt"
	"time"
)

func MeasureTime(context string, t time.Time) {
	d := time.Since(t)
	fmt.Printf("Execution time of %s: %v\n", context, d)
}

func CurrentTime() time.Time {
	return time.Now()
}