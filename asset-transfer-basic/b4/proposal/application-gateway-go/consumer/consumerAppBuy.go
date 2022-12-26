package consumer

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"time"
	//"strconv"
	//"math"
	//"sort"
	//"sync"
	//"net/http"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

const (
	earthRadius = 6378137.0
	pricePerMater = 0.000001
	kmPerBattery = 0.05 // battery(%) * kmPerBattery = x km
	layout = "2006-01-02T15:04:05+09:00"
)

func Bid(contract *client.Contract, userName string, latitude float64, longitude float64, energyAmount int, batteryLife float64) (
	[]Energy, time.Time, error) {
	searchRange := (100 - float64(batteryLife)) * kmPerBattery * 1000 // 1000m->500mに変更
	fmt.Printf("searchRange:%g\n", searchRange)

	var successList []Energy
	var autcionStartMin time.Time
	var err error

	var tokenNum int = input.Token
	// var errEnergies []Energy

	lowerLat, upperLat, lowerLng, upperLng := determineRange(searchRange, input.Latitude, input.Longitude)
	energies, err := queryByLocationRange(contract, lowerLat, upperLat, lowerLng, upperLng)
	if err != nil {
		fmt.Println("query error")
		return energies, err
	}
	if(len(energies) == 0){
		return energies, nil
	}
	
	// fmt.Println(energies)
	fmt.Printf("length of energies: %d\n", len(energies))

	timestamp := time.Now()
	auctionStartTimeCompare := timestamp.Add(time.Minute * -5)

	validEnergies := []Energy{}

	for _, energy := range energies {
		distance := distance(input.Latitude, input.Longitude, energy.Latitude, energy.Longitude)
		if energy.Owner != input.User && distance <= searchRange && auctionStartTimeCompare.After(energy.AuctionStartTime) == false {
			energy.BidPrice = energy.UnitPrice + distance * pricePerMater
			validEnergies = append(validEnergies, energy)
			fmt.Println("it's valid")
			fmt.Printf("id:%s, latitude:%g, longitude:%g, unitPrice:%g, distance:%g, bidPrice:%g\n", 
			energy.ID, energy.Latitude, energy.Longitude, energy.UnitPrice, distance, energy.BidPrice)
		}else {
			fmt.Println("it's invalid")
			fmt.Printf("id:%s, latitude: %g, longitude:%g, unitPrice:%g, distance:%g, auctionStartTime:%s\n",
		energy.ID, energy.Latitude, energy.Longitude, energy.UnitPrice, distance, energy.AuctionStartTime.Format(layout))
		}
		
	}

	sort.Slice(validEnergies, func(i, j int) bool {
        return validEnergies[i].BidPrice > validEnergies[j].BidPrice
    })
	//fmt.Println(validEnergies)

	// validEnergiesのうち、上からtokenNum個分Bid

	var bidNum int
	success := []Energy{}
	
	for {
		if(tokenNum == 0 || len(validEnergies) == 0) {
			break
		}
		fmt.Printf("requested token:%d\n", tokenNum)
		fmt.Printf("valid energy token:%d\n", len(validEnergies))
		if(tokenNum > len(validEnergies)){
			bidNum = len(validEnergies)
		}else {
			bidNum = tokenNum
		}
		fmt.Printf("max:%d\n", bidNum)

		tempSuccess := bid(contract, validEnergies, bidNum, input)

		success = append(success, tempSuccess...)
		validEnergies = validEnergies[bidNum:]
		tokenNum -= len(tempSuccess)
	}

	return success, nil
	

	return successList, autcionStartMin, err
}

func bid(contract *client.Contract, energies []Energy, bidNum int, input Input) []Energy {
	successEnergy := []Energy{}
	//leftEnergy := energies
	
	c := make(chan Energy)

	for i := 0; i < bidNum; i++ {

		go func(i int, c chan Energy){
			fmt.Printf("id:%s, auctionStartTime:%s\n", energies[i].ID, energies[i].AuctionStartTime.Format(layout))
			message, err := bidOnToken(contract, energies[i].ID, energies[i].BidPrice, input.User)
			if err != nil {
				energies[i].Error = "bidOnTokenError: " + err.Error()
				c <- energies[i]
			}
			fmt.Println(message)
			if (message == "your bid was successful") {
				go httpPost(energies[i], input)
				bidResult, err := readToken(contract, energies[i].ID)
				if err != nil {
					energies[i].Error = "readTokenError: " + err.Error()
					// energies[i].MyBidStatus = err.Error()
					c <- energies[i]
				} else{
					bidResult.Error = "OK"
				}
				c <- bidResult
				// successEnergy = append(successEnergy, bidResult)
			// auctionstart + 5min 経ったら見に行く
			} else {
				energies[i].Error = "OK"
				c <- energies[i]
			}
		}(i, c)

	}

	for i := 0; i < bidNum; i++ {
		energy := <-c
		if (energy.Owner == input.User && energy.Error == "OK") {
			successEnergy = append(successEnergy, energy)
		}
	}

	return successEnergy
}

func bidOnToken(contract *client.Contract, energyId string, bidPrice float64, username string) (string, error) {
	//fmt.Printf("Evaluate Transaction: BidOnToken, function returns asset attributes\n")
	var timestamp = time.Now()
	var stringTimestamp = timestamp.Format(layout)
	var stringBidPrice = strconv.FormatFloat(bidPrice, 'f', -1, 64)
	//fmt.Printf("id:%s, timestamp:%s, price:%s\n", energyId, stringTimestamp, stringBidPrice)
	evaluateResult, err := contract.SubmitTransaction("BidOnToken", energyId, username, stringBidPrice, stringTimestamp)
	if err != nil {
		return "", err
		// panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	//result := formatJSON(evaluateResult)
	message := string(evaluateResult)
	/* "your bid was successful" */
	return message, nil
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
	returnUpperLat := 2 * myLatitude - math.Abs(lowerLat) //緯度が0のとき、lowerLatがマイナスなため。日本は関係ないが。


	fmt.Printf("lowerLng:%g\n", returnLowerLat)
	fmt.Printf("uperLng:%g\n", returnUpperLat)
	fmt.Printf("lowerLng:%g\n", returnLowerLng)
	fmt.Printf("uperLng:%g\n", returnUpperLng)

	return returnLowerLat, returnUpperLat, returnLowerLng, returnUpperLng

}

func queryByLocationRange(contract *client.Contract, lowerLat float64, upperLat float64, lowerLng float64, upperLng float64) ([]Energy, error) {
	strLowerLat := strconv.FormatFloat(lowerLat, 'f', -1, 64)
	strUpperLat := strconv.FormatFloat(upperLat, 'f', -1, 64)
	strLowerLng := strconv.FormatFloat(lowerLng, 'f', -1, 64)
	strUpperLng := strconv.FormatFloat(upperLng, 'f', -1, 64)

	fmt.Printf("Async Submit Transaction: QueryByLocationRange'\n")

	result := []Energy{}
	evaluateResult, err := contract.EvaluateTransaction("QueryByLocationRange", "generated", strLowerLat, strUpperLat, strLowerLng, strUpperLng)
	if err != nil {
		return result, err
		// panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Println(len(evaluateResult))

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