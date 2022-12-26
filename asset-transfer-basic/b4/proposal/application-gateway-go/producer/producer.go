package producer

import (
	"fmt"
	"time"
)

// var startTime time.Time
var nowTime time.Time
var startTime time.Time

func AllProducers(start time.Time) {
	startTime = start
	nowTime = time.Now()
	timestamp := startTime.Add(time.Since(nowTime))
	fmt.Println(timestamp)
}