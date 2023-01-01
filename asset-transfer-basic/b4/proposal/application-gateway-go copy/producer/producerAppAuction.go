package producer

import (
	"fmt"
	"time"
	"encoding/json"
	"sync"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func Auction(contract *client.Contract, energy Energy) {
	timer := time.NewTimer(time.Duration(Interval * 60 - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime - energy.AuctionStartTime)) * time.Second / time.Duration(Speed))
	endTimer := time.NewTimer(time.Duration((EndTime - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime)) / Speed) * time.Second)

	auctionEndCount := 1
	select {
	case <- timer.C:
		ticker := time.NewTicker(time.Duration(Interval * 60) * time.Second / time.Duration(Speed))

		timestamp := fmt.Sprintf("%d", (time.Now().Unix() -Diff - StartTime) * Speed + StartTime)

		// fmt.Printf("auctionEndCall: id: %s, count: %d\n", energy.ID, auctionEndCount)
		isSold := auctionEnd(contract, energy, timestamp)
		if isSold == true {
			return
		}

		for i := 1; i < int(TokenLife / Interval); i++ {
			auctionEndCount++
			select {
			case <- ticker.C:
				timestamp := fmt.Sprintf("%d", (time.Now().Unix() -Diff - StartTime) * Speed + StartTime)
				if (auctionEndCount == 31) {
					fmt.Printf("auctionEndCall: id: %s, count: %d\n", energy.ID, auctionEndCount)
				}
				isSold := auctionEnd(contract, energy, timestamp)
				if isSold == true {
					ticker.Stop()
					return
				}
			case <- endTimer.C:
				ticker.Stop()
				return
			}
		}
	case <-  endTimer.C:
		timer.Stop()
		return
	}

}

func auctionEnd(contract *client.Contract, energy Energy, timestamp string) bool {
	var energies []Energy
	// var err error
	isSold := false
	var amountSold float64
	sGeneratedTime := fmt.Sprintf("%d", energy.GeneratedTime)

	auctionEndLoop:
	for {
		// fmt.Printf("producer: %s, time: %s", energy.Producer, sGeneratedTime)
		evaluateResult, err := contract.SubmitTransaction("AuctionEnd", energy.ID, sGeneratedTime)
		if err != nil {
			fmt.Printf("auction end error: %s\n", err.Error())
		} else {
			message = string(evaluateResult)
			if (message == "the energy was generated more than 30min ago. This was not sold." || message == "auction ended") {
				isSold = true
			}
			break auctionEndLoop
		}
	}
	return isSold
}
func auctionEndTransaction(contract *client.Contract, energyId string, timestamp string) (string, error) {
	// fmt.Printf("Evaluate Transaction: auctionEnd\n")
	
	evaluateResult, err := contract.SubmitTransaction("AuctionEnd", energyId, timestamp)
	if err != nil {
		return "", err
	}
	massage := string(evaluateResult)

	// fmt.Printf("*** %s Result:%s\n", energyId, massage)
	return massage, nil
}
