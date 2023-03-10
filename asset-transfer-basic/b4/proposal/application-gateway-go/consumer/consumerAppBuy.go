package consumer

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"time"
	"strconv"
	"math"
	"sort"
	"sync"
	"math/rand"
	//"net/http"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type Output struct {
	Message string `json:"Message"`
	Amount float64 `json:"Amount"`
}

const (
	earthRadius = 6378137.0
	pricePerMater = 0.000001
	kmPerBattery = 0.1 // battery(%) * kmPerBattery = x km
	layout = "2006-01-02T15:04:05+09:00"
)

func Bid(contract *client.Contract, data Data) (Data, error) {
	endTimer := time.NewTimer(time.Duration((EndTime - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime)) / Speed) * time.Second)

	search := 100 - data.BatteryLife
	searchRange := search * kmPerBattery * 1000 // 1000m->500mに変更
	// fmt.Printf("searchRange:%g\n", searchRange)

	// var successList []Energy
	var err error
	// var errEnergies []Energy

	lowerLat, upperLat, lowerLng, upperLng := determineRange(searchRange, data.Latitude, data.Longitude)
	energies, err := queryByLocationRange(contract, data.UserName, lowerLat, upperLat, lowerLng, upperLng)
	if err != nil {
		fmt.Println("query error")
		return data, err
	}
	if(len(energies) == 0){
		return data, nil
	}
	
	// fmt.Println(energies)
	// fmt.Printf("length of energies: %d\n", len(energies))

	timestamp := (time.Now().Unix() - Diff - StartTime) * Speed + StartTime
	auctionStartTimeCompare := timestamp - 60 * Interval

	validEnergies := []Energy{}

	for _, energy := range energies {
		distance := distance(data.Latitude, data.Longitude, energy.Latitude, energy.Longitude)
		if distance <= searchRange && auctionStartTimeCompare <= energy.AuctionStartTime {
			myBidPrice = energy.UnitPrice + distance * pricePerMater
			if (myBidPrice > energy.BidPrice || (myBidPrice == energy.BidPrice && (1 - data.BatteryLife) > energy.Priority)) {
				validEnergies = append(validEnergies, energy)
			}
			// fmt.Println("it's valid")
			// fmt.Printf("id:%s, latitude:%g, longitude:%g, unitPrice:%g, distance:%g, bidPrice:%g\n", 
			// energy.ID, energy.Latitude, energy.Longitude, energy.UnitPrice, distance, energy.BidPrice)
		}else {
			// fmt.Println("it's invalid")
			// fmt.Printf("id:%s, latitude: %g, longitude:%g, unitPrice:%g, distance:%g, auctionStartTime:%d\n",
		// energy.ID, energy.Latitude, energy.Longitude, energy.UnitPrice, distance, energy.AuctionStartTime)
		}
		
	}

	sort.SliceStable(validEnergies, func(i, j int) bool {
		return validEnergies[i].GeneratedTime < validEnergies[j].GeneratedTime
	})

	sort.SliceStable(validEnergies, func(i, j int) bool {
        return validEnergies[i].BidPrice < validEnergies[j].BidPrice
    })
	//fmt.Println(validEnergies)

	/*fmt.Println("sort validEnergies")
	for i := 0; i < len(validEnergies) ; i++ {
		if (i < 7) {
			fmt.Printf("id: %s, bidPrice:%v, generatedTime:%v\n", validEnergies[i].ID, validEnergies[i].BidPrice, validEnergies[i].GeneratedTime)
		} else {break}
	}*/

	leftAmount := data.Requested
	success := []Energy{}
	
	loop:
		for {
			select {
			case <- endTimer.C:
				break loop
			default:
				// default
				
			}
			if(leftAmount == 0 || len(validEnergies) == 0) {
				break loop
			}
			// fmt.Printf("requested Amount:%g\n", leftAmount)
			// fmt.Printf("valid energy token:%d\n", len(validEnergies))
			/*if(tokenNum > len(validEnergies)){
				bidNum = len(validEnergies)
			}else {
				bidNum = tokenNum
			}
			fmt.Printf("max:%d\n", bidNum)*/
			want := leftAmount
			tokenCount := 0
			for i := 0; i < len(validEnergies); i++ {
				if (want == 0) {
					break
				}
				if (validEnergies[i].Amount > want) {
					validEnergies[i].Amount = want
					want = 0
					tokenCount++
					break
				} else {
					want -= validEnergies[i].Amount
					tokenCount++
				}
			}

			// tempSuccess := bid(contract, validEnergies, tokenCount, input)

			tempSuccess, tempAmount := bid(contract, validEnergies, tokenCount, data)

			success = append(success, tempSuccess...)
			validEnergies = validEnergies[tokenCount:]

			// tokenNum -= len(tempSuccess)
			leftAmount -= tempAmount
		}
	
	data.BidAmount = 0
	data.BidSolar = 0
	data.BidWind = 0 
	data.BidThermal = 0
	data.LatestAuctionStartTime = 0
	data.LastBidTime = 0
	data.FirstBidTime = time.Now().Unix()

	for i := 0; i < len(success); i++ {
		data.BidAmount += success[i].Amount
		switch success[i].SmallCategory {
		case "solar":
			data.BidSolar += success[i].Amount
		case "wind":
			data.BidWind += success[i].Amount
		case "thermal":
			data.BidThermal += success[i].Amount
		}
		if (data.LatestAuctionStartTime < success[i].AuctionStartTime) {
			data.LatestAuctionStartTime = success[i].AuctionStartTime
		}
		if (data.LastBidTime < success[i].BidTime) {
			data.LastBidTime = success[i].BidTime
		} else if (data.FirstBidTime > success[i].BidTime) {
			data.FirstBidTime = success[i].BidTime
		}
	}
	fmt.Printf("%s bid return : %fWh\n", data.UserName, data.BidAmount)
	return data, nil
	

	// return successList, autcionStartMin, err
}

func bid(contract *client.Contract, energies []Energy, tokenCount int, data Data) ([]Energy, float64) {
	successEnergy := []Energy{}
	//leftEnergy := energies
	
	//c := make(chan Energy)
	var wg sync.WaitGroup
	for i := 0; i < tokenCount; i++ {
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			// fmt.Printf("id:%s, auctionStartTime:%v\n", energies[i].ID, time.Unix(energies[i].AuctionStartTime, 0))
			// id string, newOwner string, newBidPrice float64, priority float64, amount float64, timestamp int64, newID string) (*Output, error) {
			// var output Output
			output, timestamp, err := bidOnToken(contract, energies[i].ID, energies[i].BidPrice, data.UserName, data.BatteryLife, energies[i].Amount)
			if err != nil {
				fmt.Println("function bid error")
				energies[i].Error = "bidOnTokenError: " + err.Error()
				energies[i].Status = "F"
				//c <- energies[i]
			}else if (output.Message == "your bid was successful" || output.Message == "amount changed" || output.Message == "your bid was successful. New token was created") {
				// fmt.Printf("%s bid output: %s\n", data.UserName, output.Message)
				energies[i].BidTime = timestamp
				energies[i].Amount = output.Amount
				energies[i].Status = "S"
				//<- energies[i]
				/*bidResult, err := readToken(contract, energies[i].ID)
				if err != nil {
					energies[i].Error = "readTokenError: " + err.Error()
					// energies[i].MyBidStatus = err.Error()
					c <- energies[i]
				} else{
					bidResult.Error = "OK"
				}
				c <- bidResult*/

				// successEnergy = append(successEnergy, bidResult)
			// auctionstart + 5min 経ったら見に行く
			} else {
				// fmt.Printf("else: %s\n", output.Message)
				energies[i].Error = "OK"
				energies[i].Status = "F"
				// c <- energies[i]
			}
		}(i)
	}
	wg.Wait()

	var successAmount float64
	for i := 0; i < tokenCount; i++ {
		// energy := <-c
		if (energies[i].Status == "S") {
			successEnergy = append(successEnergy, energies[i])
			successAmount += energies[i].Amount
			/*if (energies[i].AuctionStartTime > latestAuction) {
				latestAuction = energies[i].AuctionStartTime
			} 
			if (energies[i].BidTime > lastBidTime) {
				lastBidTime = energies[i].BidTime
			} else if (energies[i].BidTime < firstBidTime){
				firstBidTime = energies[i].BidTime
			}*/
		}
	}

	return successEnergy, successAmount
}

// id string, newOwner string, newBidPrice float64, priority float64, amount float64, timestamp int64, newID string
func bidOnToken(contract *client.Contract, energyId string, bidPrice float64, username string, batteryLife float64, amount float64) (Output, int64, error) {
	//fmt.Printf("Evaluate Transaction: BidOnToken, function returns asset attributes\n")
	endTimer := time.NewTimer(time.Duration((EndTime - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime)) / Speed) * time.Second)

	var output Output
	timestamp := (time.Now().Unix() - Diff - StartTime) * Speed + StartTime
	sTimestamp := fmt.Sprintf("%v", timestamp)
	sBidPrice := fmt.Sprintf("%v", bidPrice)
	sPriority := fmt.Sprintf("%v", 1 - batteryLife)
	sAmount := fmt.Sprintf("%v", amount)

	rand.Seed(time.Now().UnixNano())
	// create id
	newid := fmt.Sprintf("%s%s-%d", energyId, sTimestamp, rand.Intn(10000))

	// fmt.Printf("bid id:%s, timestamp:%s, price:%s\n", energyId, sTimestamp, sBidPrice)
	count := 0
	bidLoop:
	for {
		select {
		case <- endTimer.C:
			return output, timestamp, fmt.Errorf("time up")
		default:
			// sonomama
			evaluateResult, err := contract.SubmitTransaction("BidOnToken", energyId, username, sBidPrice, sPriority, sAmount, sTimestamp, newid)
			if err != nil {
				fmt.Printf("bid error: %s, %v\n", energyId, err.Error())
				rand.Seed(time.Now().UnixNano())
				timer := time.NewTimer(time.Duration(rand.Intn(1000000000)) * time.Nanosecond / time.Duration(Speed))
				count++
				<- timer.C
			} else {
				err = json.Unmarshal(evaluateResult, &output)
				if err != nil {
					fmt.Println("unmarshal error")
				} else {
					break bidLoop
				}
			}
		}
	}
	// fmt.Printf("bid count %s, %d\n", energyId, count)
	/* "your bid was successful" */
	return output, timestamp, nil
}

func determineRange(length float64, myLatitude float64, myLongitude float64) (lowerLat float64, upperLat float64, lowerLng float64, upperLng float64) {
	// 緯度固定で経度求める
	rlat := myLatitude * math.Pi / 180
	r := length / earthRadius
	angle := math.Cos(r)

	lngTmp := (angle - math.Sin(rlat) * math.Sin(rlat)) / (math.Cos(rlat) * math.Cos(rlat))
	rlngDifference := math.Acos(lngTmp)
	lngDifference := rlngDifference * 180 / math.Pi
	returnLowerLng := myLongitude - lngDifference
	returnUpperLng := myLongitude + lngDifference

	// 経度固定で緯度求める
	// rlng := myLongitude * math.Pi / 180
	//latTmp := angle / (math.Sin(rlat) + math.Cos(rlat))
	rSinLat := math.Sin(rlat)
	rCosLat := math.Cos(rlat)
	square := math.Sqrt(math.Pow(rSinLat, 2) + math.Pow(rCosLat, 2))
	latTmp := math.Asin(angle / square)
	solutionRLat := latTmp - math.Acos(rSinLat / square)
	// 緯度はプラスなため、solutionLatは常にmylatitudeより小さい
	returnLowerLat := solutionRLat * 180 / math.Pi
	returnUpperLat := 2 * myLatitude - math.Abs(returnLowerLat) //緯度が0のとき、lowerLatがマイナスなため。日本は関係ないが。


	// fmt.Printf("lowerLat:%g\n", returnLowerLat)
	// fmt.Printf("upperLat:%g\n", returnUpperLat)
	// fmt.Printf("lowerLng:%g\n", returnLowerLng)
	// fmt.Printf("upperLng:%g\n", returnUpperLng)

	return returnLowerLat, returnUpperLat, returnLowerLng, returnUpperLng

}

func queryByLocationRange(contract *client.Contract, owner string, lowerLat float64, upperLat float64, lowerLng float64, upperLng float64) ([]Energy, error) {
	strLowerLat := strconv.FormatFloat(lowerLat, 'f', -1, 64)
	strUpperLat := strconv.FormatFloat(upperLat, 'f', -1, 64)
	strLowerLng := strconv.FormatFloat(lowerLng, 'f', -1, 64)
	strUpperLng := strconv.FormatFloat(upperLng, 'f', -1, 64)

	// fmt.Printf("Async Submit Transaction: QueryByLocationRange'\n")

	result := []Energy{}
	evaluateResult, err := contract.EvaluateTransaction("QueryByLocationRange", "generated", owner, strLowerLat, strUpperLat, strLowerLng, strUpperLng)
	if err != nil {
		return result, err
		// panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	// fmt.Println(len(evaluateResult))

	err = json.Unmarshal(evaluateResult, &result)
	if(err != nil && len(evaluateResult) > 0) {
		return result, err
	}

	return result, nil

}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	// 緯度経度をラジアンに変換
	rlat1 := lat1 * math.Pi / 180
	rlng1 := lng1 * math.Pi / 180
	rlat2 := lat2 * math.Pi / 180
	rlng2 := lng2 * math.Pi / 180

	// 2点の中心角を求める。
	/*cos(c)=cos(a)cos(b) + sin(a)sin(b)cos(c)
	= cos(pi/2 - lat1)cos(pi/2 - lat2) + sin(lat1)sin(lat2)cos(lng1 - lng2)
	= cos(sin(lat1)sin(lat2) + sin(lat1)sin(lat2)cos(lng1 - lng2))
	*/
	angle := 
		math.Sin(rlat1) * math.Sin(rlat2) +
		math.Cos(rlat1) * math.Cos(rlat2) *
		math.Cos(rlng1 - rlng2)

	r := math.Acos(angle)
	distance := earthRadius * r
	
	return distance
}