package producer

import (
	"fmt"
	"math/rand"
	"time"
	"encoding/json"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func Create(contract *client.Contract, input Input) () {
	
	errCount := 0
	var energy Energy
	var err error

	createLoop:
	for {
		energy, err = createToken(contract, input)
		if err != nil {
			fmt.Printf("create Token Error: %s\n", err.Error())
			errCount++
			rand.Seed(time.Now().UnixNano())
			timer := time.NewTimer(time.Duration(rand.Intn(1000000000)) * time.Nanosecond / time.Duration(Speed))
			<- timer.C
			if errCount > 3 {
				fmt.Println("many create error")
				break createLoop
			}
		} else {break createLoop}
	}
	if err != nil {
		return 
	} else {
		Auction(contract, energy)
	}

}

func createToken(contract *client.Contract, input Input) (Energy, error) {
	// fmt.Printf("Submit Transaction: CreateToken, creates new token with ID, Latitude, Longitude, Owner, Large Category, Small Category and timestamp \n")
	var largeCategory string
	if (input.Category == "solar" || input.Category == "wind") {
		largeCategory = "green"
	} else {
		largeCategory = "depletable"
	}

	rand.Seed(time.Now().UnixNano())
	// create id
	id := fmt.Sprintf("%d%s-%d", input.Timestamp, input.User, rand.Intn(10000))

	sTimestamp := fmt.Sprintf("%v", input.Timestamp)
	sLat := fmt.Sprintf("%v", input.Latitude)
	sLon := fmt.Sprintf("%v", input.Longitude)
	sAmo := fmt.Sprintf("%v", input.Amount)

	var energy Energy

	// fmt.Println("create")

	evaluateResult, err := contract.SubmitTransaction("CreateEnergyToken", id, sLat, sLon, input.User, sAmo, largeCategory, input.Category, sTimestamp)
	if err != nil {
		return energy, err
	}

	err = json.Unmarshal(evaluateResult, &energy)
	if err != nil {
		return energy, err
	}

	// fmt.Printf("*** %s: Transaction committed successfully\n", id)

	return energy, nil
}