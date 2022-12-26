package main

import (
	// "fmt"
	"time"
	"sync"
	prdc "github.com/Smart-Latte/fabric-samples/blockchain-application/b4/proposal/producer"
	cnsm "github.com/Smart-Latte/fabric-samples/blockchain-application/b4/proposal/consumer"
)

func main() {
	startTime := time.Date(2022, time.April, 1, 4, 0, 0, 0, time.Local)
	interval := 1
	speed := 6
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		cnsm.AllConsumers(startTime, speed, interval)
	}()
	go func() {
		defer wg.Done()
		prdc.AllProducers(startTime)
	}()
	wg.Wait()
}