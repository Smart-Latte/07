package producer

import (
	"fmt"
	"time"
	"math/rand"
	"math"
	"sync"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func DummyWindProducer(contract *client.Contract, username string, lLat float64, uLat float64, lLon float64, uLon float64, category string, ratingOutput float64, ratingSpeed float64, 
	cutIn float64, outputList [dayNum][hourNum]float64, seed int64){

		rand.Seed(seed)
		lat := rand.Float64() * (uLat - lLat) + lLat
		lon := rand.Float64() * (uLon - lLon) + lLon
		fmt.Printf("lat: %g, lon: %g\n", lat, lon)

		SeaWindProducer(contract, username, lat, lon, category, ratingOutput, ratingSpeed, cutIn, outputList, seed)

	}


func SeaWindProducer(contract *client.Contract, username string, lat float64, lon float64, category string, ratingOutput float64, ratingSpeed float64, cutIn float64, 
	speedList [dayNum][hourNum]float64, seed int64) {
	var outputList[dayNum][hourNum] float64

	for i := 0; i < dayNum; i++ {
		for j := 0; j < hourNum; j++ {
			if (speedList[i][j] >= cutIn) {
				outputList[i][j] = math.Pow((speedList[i][j] / ratingSpeed), 3) * ratingOutput
			}else {
				outputList[i][j] = 0
			}
		}
		fmt.Println(outputList[i])
	}
	Produce(contract, username, lat, lon, category, 1, outputList, seed)
	
}

func DummySolarProducer(contract *client.Contract, username string, lLat float64, uLat float64, lLon float64, uLon float64, category string, 
	output float64, outputList [dayNum][hourNum]float64, seed int64) {
	rand.Seed(seed)
	lat := rand.Float64() * (uLat - lLat) + lLat
	lon := rand.Float64() * (uLon - lLon) + lLon
	fmt.Printf("lat: %g, lon: %g\n", lat, lon)

	Produce(contract, username, lat, lon, category, output, outputList, seed)
	
}

func Produce(contract *client.Contract, username string, lat float64, lon float64, category string, output float64, outputList [dayNum][hourNum]float64, seed int64) {

	endTimer := time.NewTimer(time.Duration((EndTime - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime)) / Speed) * time.Second)

	// output per min during un hour
	var myOutput[dayNum][hourNum] float64
	var timing[dayNum][hourNum] float64
	var wg sync.WaitGroup
	counter := 0

	var maxCreateInterval float64 = 2.5 // min
	var maxCreateAmount float64 = 2500 // Wh

	/*for i := 0; i <  dayNum; i++ {
		for j := 0; j < hourNum; j++ {
			fmt.Printf("%v ", output * outputList[i][j])
		}
		fmt.Println("")
	}*/

	for i := 0; i < dayNum; i++ {
		for j := 0; j < hourNum; j++ {
			outputPerHour := output * outputList[i][j]

			if (outputPerHour >= maxCreateAmount * 60 / maxCreateInterval) {
				myOutput[i][j] = maxCreateAmount
				timing[i][j] = 60 * 60 / (outputPerHour / maxCreateAmount)
			} else {
				myOutput[i][j] = outputPerHour / 60 * maxCreateInterval
				timing[i][j] = 60 * maxCreateInterval
			}
			// fmt.Printf("output:%v, timing:%v ", myOutput[i][j], timing[i][j])
		}
	}
	

	rand.Seed(seed)
	wait := rand.Intn(60)
	waitNano := rand.Intn(1000000000)
	fmt.Printf("%s wait : %d, waitNano:%d\n", username, wait + 5, waitNano)
	timer := time.NewTimer((time.Duration(waitNano) * time.Nanosecond + time.Duration(5 + int64(wait) + time.Now().Unix() - Diff - StartTime) * time.Second) / time.Duration(Speed))

	<- timer.C
	loop:
		for counter < 24 {
			thisTime := counter + StartHour
			var thisTiming float64
			var thisOut float64
			if thisTime < 24 {
				thisOut = myOutput[0][thisTime]
				thisTiming = timing[0][thisTime]
			} else if (thisTime < 48) {
				thisOut = myOutput[1][thisTime - 24]
				thisTiming = timing[1][thisTime - 24]
			} else {
				fmt.Println("simulation is too long")
			}
			ticker := time.NewTicker(time.Second * time.Duration(thisTiming) / time.Duration(Speed))

			thisTimeCounter := 0
			for {
				if (float64(thisTimeCounter) >= 60 * 60 / thisTiming) {
					ticker.Stop()
					fmt.Printf("producer counter:%d, this time counter:%d\n", counter, thisTimeCounter)
					break
				}
				// ログ
				select {
				case <-ticker.C:
					if (thisOut > 0) {
						timestamp := (time.Now().Unix() - Diff - StartTime) * Speed + StartTime
						var input Input = Input{User: username, Latitude: lat, Longitude: lon, Amount: thisOut, Category: category, Timestamp: timestamp}
						wg.Add(1)
						go func(i Input) {
							defer wg.Done()
							Create(contract, i)
						}(input)
					}
					// wg.Wait()
					thisTimeCounter++
				case <- endTimer.C:
					ticker.Stop()
					timestamp := (time.Now().Unix() - Diff - StartTime) * Speed + StartTime
					fmt.Printf("PRODUCER END TIMER: %v\n", time.Unix(timestamp, 0))
					break loop
				}
				// Create
				// create counter + startHour
			// id string, latitude float64, longitude float64, producer string, amount float64, largeCategory string, smallCategory string, timestamp int64)
			}
		}
	wg.Wait()
 }