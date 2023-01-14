package consumer

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"time"
	//"strconv"
	//"math"
	//"sort"
	//"sync"
	//"net/http"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

/*const (
	earthRadius = 6378137.0
	pricePerMater = 0.000001
	kmPerBattery = 0.05 // battery(%) * kmPerBattery = x km
	layout = "2006-01-02T15:04:05+09:00"
)*/

func BidResult(contract *client.Contract, data Data) (Data, error) {
	energies, err := queryByUserAndBidTime(contract, data.UserName, data.FirstBidTime, data.LastBidTime)
	if err != nil {
		return data, err
	}
	data.GetAmount = 0
	data.GetSolar = 0
	data.GetWind = 0
	data.GetThermal = 0

	for _, energy := range energies {
		data.GetAmount += energy.Amount

		switch energy.SmallCategory {
		case "solar":
			data.GetSolar += energy.Amount
		case "wind":
			data.GetWind += energy.Amount
		case "thermal":
			data.GetThermal += energy.Amount
		}
	}

	return data, nil
}

func queryByUserAndBidTime(contract *client.Contract, owner string, startTime int64, endTime int64) ([]Energy, error) {
	sStartTime := fmt.Sprintf("%d", startTime)
	sEndTime := fmt.Sprintf("%d", endTime)
	fmt.Printf("Async Submit Transaction: QueryByUserAndBidTime: %s, startTime:%v, endTime:%v\n", owner, time.Unix(startTime, 0), time.Unix(endTime, 0))
	result := []Energy{}

	queryLoop:
	for {
		evaluateResult, err := contract.EvaluateTransaction("QueryByUserAndBidTime", owner, "sold", sStartTime, sEndTime)
		if err != nil {
			fmt.Printf("BID RESULT ERROR: %v\n", err.Error())
			// panic(fmt.Errorf("failed to evaluate transaction: %w", err))
		} else {
			err = json.Unmarshal(evaluateResult, &result)
			if(err != nil && len(evaluateResult) > 0) {
				fmt.Printf("unmarshal error in queryByUserAndBidTime\n")
			} else {
				fmt.Printf("%s break queryLoop\n", owner)
				fmt.Printf("%s: queryBuyResult: %v\n", owner, len(result))
				break queryLoop
			}
		} 
	}
	return result, nil

}