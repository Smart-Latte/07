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

	for {
		// fmt.Printf("producer: %s, time: %s", energy.Producer, sGeneratedTime)
		evaluateResult, err := contract.SubmitTransaction("QueryByUserAndGeneratedTime", energy.Producer, sGeneratedTime)
		if err != nil {
			fmt.Printf("auction end query error: %s\n", err.Error())
		} else {
			err = json.Unmarshal(evaluateResult, &energies)
			if(err != nil) {
				fmt.Println("unmarshal error")
			} else {
				break
			}
		}
	}

	/*for _, e := range energies {
		fmt.Println(e)
	}*/

	var wg sync.WaitGroup
	// fmt.Printf("energies length: %d\n", len(energies))
	for i:= 0; i < len(energies); i++ {
		wg.Add(1)

		go func(n int) {
			defer wg.Done()

			if (energies[n].Status == "sold" || energies[n].Status == "old") {
				return
			}
			count := 0
			for {
				if (count > 0) {
					fmt.Printf("%s count is%d\n", energies[n].ID, count)
				}
				count++
				message, err := auctionEndTransaction(contract, energies[n].ID, timestamp)
				if err != nil {
					fmt.Println(err)
				} else {
					stopmassage1 := "the energy " + energies[n].ID + " was generated more than 30min ago. This was not sold."
					stopmassage2 := "the energy " + energies[n].ID + " was sold"
					if message == stopmassage1 || message == stopmassage2 {
						// fmt.Printf("%s: %s\n", energies[n].ID, message)
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
	// fmt.Printf("Evaluate Transaction: auctionEnd\n")
	
	evaluateResult, err := contract.SubmitTransaction("AuctionEnd", energyId, timestamp)
	if err != nil {
		return "", err
	}
	massage := string(evaluateResult)

	// fmt.Printf("*** %s Result:%s\n", energyId, massage)
	return massage, nil
}
