package producer

import (
	"fmt"
	"time"
	"encoding/json"
	"sync"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func Auction(contract *client.Contract, energy Energy) {
	timer := time.NewTimer(time.Duration(60 - ((time.Now().Unix() - Diff - StartTime) * Speed + StartTime - energy.AuctionStartTime)) * time.Second / time.Duration(Speed))
	<- timer.C

	ticker := time.NewTicker(time.Duration(Interval * 60) * time.Second / time.Duration(Speed))

	timestamp := fmt.Sprintf("%d", (time.Now().Unix() -Diff - StartTime) * Speed + StartTime)

	fmt.Printf("create time: %v\n", time.Unix(energy.AuctionStartTime, 0))
	fmt.Printf("auctionEnd call: %v\n", time.Unix((time.Now().Unix() - Diff - StartTime) * Speed + StartTime, 0))
	isSold := auctionEnd(contract, energy, timestamp)
	if isSold == true {
		return
	}

	for i := 0; i < 30; i++ {
		<- ticker.C
		timestamp := fmt.Sprintf("%d", (time.Now().Unix() -Diff - StartTime) * Speed + StartTime)
		fmt.Printf("create time: %v\n", time.Unix(energy.AuctionStartTime, 0))
		fmt.Printf("auctionEnd call: %v\n", time.Unix((time.Now().Unix() - Diff - StartTime) * Speed + StartTime, 0))
		isSold := auctionEnd(contract, energy, timestamp)
		if isSold == true {
			ticker.Stop()
			break
		}
	}
}

func auctionEnd(contract *client.Contract, energy Energy, timestamp string) bool {
	var energies []Energy
	// var err error
	isSold := false
	var amountSold float64
	sGeneratedTime := fmt.Sprintf("%d", energy.GeneratedTime)

	for {
		// fmt.Printf("producer: %s, time: %s", energy.Producer, sGeneratedTime)
		evaluateResult, err := contract.SubmitTransaction("QueryByUserAndGeneratedTime", energy.Producer, sGeneratedTime)
		if err != nil {
			fmt.Println("auction end query error")
		} else {
			err = json.Unmarshal(evaluateResult, &energies)
			if(err != nil) {
				fmt.Println("unmarshal error")
			} else {
				break
			}
		}
	}

	for _, e := range energies {
		fmt.Println(e)
	}

	var wg sync.WaitGroup
	fmt.Printf("energies length: %d\n", len(energies))
	for i:= 0; i < len(energies); i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			if (energies[n].Status == "sold" || energies[n].Status == "old") {
				return
			}
			count := 0
			for {
				fmt.Printf("%s count is%d\n", energies[n].ID, count)
				count++
				message, err := auctionEndTransaction(contract, energies[n].ID, timestamp)
				if err != nil {
					fmt.Println(err)
				} else {
					stopmassage1 := "the energy " + energies[n].ID + " was generated more than 30min ago. This was not sold."
					stopmassage2 := "the energy " + energies[n].ID + " was sold"
					if message == stopmassage1 || message == stopmassage2 {
						fmt.Println("sold")
					} else {
						energies[n].Amount = 0
					}
					break
				}
			}

		}(i)
	}
	wg.Wait()

	for i:= 0; i < len(energies); i++ {
		amountSold += energies[i].Amount
	}

	if (amountSold == energy.Amount) {
		isSold = true
	}
	return isSold
}
func auctionEndTransaction(contract *client.Contract, energyId string, timestamp string) (string, error) {
	fmt.Printf("Evaluate Transaction: auctionEnd\n")
	
	evaluateResult, err := contract.SubmitTransaction("AuctionEnd", energyId, timestamp)
	if err != nil {
		return "", err
	}
	massage := string(evaluateResult)

	fmt.Printf("*** Result:%s\n", massage)
	return massage, nil
}
