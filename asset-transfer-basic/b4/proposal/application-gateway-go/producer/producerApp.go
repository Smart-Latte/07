package producer

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func Produce(contract *client.Contract, username string, lat float64, lon float64, category string, output float64, outputList [dayNum][hourNum]float64, seed int64) {
	// output per min during un hour
	var myOutput[dayNum][hourNum] float64
	var timing[dayNum][hourNum] float64
	var wg sync.WaitGroup
	counter := 0

	for i := 0; i < dayNum; i++ {
		for j := 0; j < hourNum; j++ {
			outputPerHour := output * outputList[i][j]
			if (outputPerHour >= 60 * 1000) {
				myOutput[i][j] = 1000
				timing[i][j] = 60 * 60 / (outputPerHour / 1000) // second
			} else {
				myOutput[i][j] = outputPerHour / 60
				timing[i][j] = 60
			}
		}
	}

	rand.Seed(seed)
	wait := rand.Intn(60)
	timer := time.NewTimer(time.Duration(5 + int64(wait) + time.Now().Unix() - Diff - StartTime) * time.Second / time.Duration(Speed))

	<- timer.C
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
			<-ticker.C
			// Create
			// create counter + startHour
		// id string, latitude float64, longitude float64, producer string, amount float64, largeCategory string, smallCategory string, timestamp int64)
			var input Input = Input{User: username, Latitude: lat, Longitude: lon, Amount: thisOut, Category: category, Timestamp: ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime)}
			wg.Add(1)
			go func(i Input) {
				defer wg.Done()
				Create(contract, i)
			}(input)
			// wg.Wait()
			thisTimeCounter++
		}
	}
	wg.Wait()
 }