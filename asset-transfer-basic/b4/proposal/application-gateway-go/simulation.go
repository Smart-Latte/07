package main

import (
	// "fmt"
	"time"
	"sync"
	oprt "github.com/Smart-Latte/fabric-samples/blockchain-application/b4/proposal/operator"
	prdc "github.com/Smart-Latte/fabric-samples/blockchain-application/b4/proposal/producer"
	cnsm "github.com/Smart-Latte/fabric-samples/blockchain-application/b4/proposal/consumer"
)

const dayNum = 2
const hourNum = 24

var tempData[dayNum][hourNum] float64 = [dayNum][hourNum]float64 {
	{87, 84, 85, 84, 83, 84, 90, 97, 109, 117, 121, 126, 130, 127, 125, 118, 116, 106, 100, 94, 92, 89, 88, 87}, 
	{85, 82, 82, 80, 76, 69, 81, 101, 112, 121, 130, 145, 140, 141, 143, 140, 128, 114, 106, 99, 94, 90, 88, 86}}
var insolation[dayNum][hourNum]  float64 =[dayNum][hourNum]float64 {
	{0, 0, 0, 0, 0, 3, 15, 52, 221, 293, 343, 366, 360, 320, 250, 114, 75, 9, 0, 0, 0, 0, 0, 0}, 
	{0, 0, 0, 0, 0, 3, 23, 99, 130, 214, 193, 319, 343, 309, 260, 156, 48, 8, 0, 0, 0, 0, 0, 0}}
var windOutput[dayNum][hourNum]  float64 =[dayNum][hourNum]float64 {
	{0.91, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1}, 
	{0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1}}
var solarOutput[dayNum][hourNum] float64

func main() {
	for i := 0; i < dayNum; i++ {
		for j := 0; j < hourNum; j++ {
			solarOutput[i][j] = 0.97 * 0.95 * 0.94 * 0.97 * 0.9 * (1 - 0.45 * (tempData[i][j] * 0.1 - 25) / 100) * (insolation[i][j] * 10 / 3.6 / 1000)
		}
	}
	startHour := 5
	startTime := time.Date(2015, time.March, 27, startHour, 0, 0, 0, time.Local).Unix()
	nowTime := time.Now().Unix()
	diff := nowTime - startTime
	endTime := time.Date(2015, time.March, 27, startHour + 2, 0, 0, 0, time.Local).Unix()
	var interval int64 = 1
	var tokenLife int64 = 30
	var speed int64 = 2
	var wg sync.WaitGroup

	oprt.InitOperator()
	wg.Add(3)
	go func() {
		defer wg.Done()
		oprt.Operator(startTime, endTime, diff, speed, solarOutput, windOutput, startHour)
	}()
	go func() {
		defer wg.Done()
		prdc.AllProducers(startTime, endTime, diff, speed, interval, tokenLife, solarOutput, windOutput, startHour)
	}()
	go func() {
		defer wg.Done()
		// start int64, end int64, diff int64, auctionSpeed int64, auctionInterval int64, life int64, startHour int
		cnsm.AllConsumers(startTime, endTime, diff, speed, interval, tokenLife, startHour)
	}()
	wg.Wait()
}