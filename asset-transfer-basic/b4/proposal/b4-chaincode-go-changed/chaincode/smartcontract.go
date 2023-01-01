package chaincode

import (
	"encoding/json"
	"fmt"
	"sort"
	//"errors"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Energy struct {
	DocType          string    `json:"DocType"`
	Amount float64 `json:"Amount"`
	BidAmount float64 `json:"BidAmount"`
	SoldAmount float64 `json:"SoldAmount"`
	UnitPrice        float64   `json:"Unit Price"`
	BidPrice         float64   `json:"Bid Price"`
	GeneratedTime    int64 `json:"Generated Time"`
	BidTime          int64 `json:"Bid Time"`
	ID               string    `json:"ID"`
	EnergyID string `json:"EnergyID"`
	LargeCategory    string    `json:"LargeCategory"`
	Latitude         float64   `json:"Latitude"`
	Longitude        float64   `json:"Longitude"`
	Owner            string    `json:"Owner"`
	Producer         string    `json:"Producer"`
	Priority float64 `json:"Priority"`
	SmallCategory    string    `json:"SmallCategory"`
	Status           string    `json:"Status"`
}

type Output struct {
	Message string `json:"Message"`
	Amount float64 `json:"Amount"`
}

const (
	tokenLife = 30 // minute
	auctionInterval = 1 // minute
)

// InitLedger adds a base set of assets to the ledger
// Owner: Brad, Jin Soo, Max, Adriana, Michel
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	energies := []Energy{
		{DocType: "cost", ID: "solar-power-cost", UnitPrice: 0.02,
			LargeCategory: "green", SmallCategory: "solar"},
		{DocType: "cost", ID: "wind-power-cost", UnitPrice: 0.02,
			LargeCategory: "green", SmallCategory: "wind"},
		{DocType: "cost", ID: "thermal-power-cost", UnitPrice: 0.03,
			LargeCategory: "depletable", SmallCategory: "thermal"},
		
	}

	for _, energy := range energies {
		energyJSON, err := json.Marshal(energy)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(energy.ID, energyJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) UpdateUnitPrice(ctx contractapi.TransactionContextInterface, 
	smallCategory string, newUnitPrice float64, timestamp int64) error {
		var id = smallCategory + "-power-cost"
		cost, err := s.ReadToken(ctx, id)
		if err != nil {
			return err
		}
		cost.UnitPrice = newUnitPrice
		cost.GeneratedTime = timestamp

		costJSON, err := json.Marshal(cost)
			if err != nil {
				return err
			}

		err = ctx.GetStub().PutState(id, costJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
		return nil
}

// CreateAsset issues a new asset to the world state with given details.
// 新しいトークンの発行
// errorは返り値の型
// 引数は、ID、緯度、経度、エネルギーの種類、発電した時間、発電者、価格
// トークンには、オーナー、ステータスも含める
func (s *SmartContract) CreateEnergyToken(ctx contractapi.TransactionContextInterface,
	id string, latitude float64, longitude float64, producer string, amount float64, largeCategory string, smallCategory string, timestamp int64) (*Energy, error) {
	var energy Energy
	var costId = smallCategory + "-power-cost"

	cost, err := s.ReadToken(ctx, costId)
	if err != nil {
		return &energy, err
	}

	exists, err := s.EnergyExists(ctx, id)

	//get unit price

	if err != nil {
		return &energy, err
	}
	if exists {
		return &energy, fmt.Errorf("the energy %s already exists", id)
	}
	
	energy = Energy{
		DocType:          "token",
		ID:               id,
		Latitude:         latitude,
		Longitude:        longitude,
		Owner:            producer,
		Producer:         producer,
		LargeCategory:    largeCategory,
		SmallCategory:    smallCategory,
		Amount: amount, 
		Status:           "generated",
		GeneratedTime:    timestamp,
		UnitPrice:        cost.UnitPrice,
		BidPrice:         cost.UnitPrice,
		Priority: -1, 
	}
	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return &energy, err
	}

	return &energy, ctx.GetStub().PutState(id, energyJSON)
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
// 購入する
func (s *SmartContract) BidOnEnergy(ctx contractapi.TransactionContextInterface, 
	bidId string, energyId string, bidder string, bidPrice float64, priority float64, amount float64, timestamp int64) (string, error) {
	energy, err := s.ReadToken(ctx, energyId)
	if err != nil {
		return "", err
	}
	
	generatedTimeCompare := timestamp - 60 * tokenLife
	if generatedTimeCompare >= energy.GeneratedTime {
		return fmt.Sprintf("the energy %s was generated more than %dmin ago", energyId, tokenLife), nil
	}

	exists, err := s.EnergyExists(ctx, bidId)

	//get unit price

	if err != nil {
		return "", err
	}
	if exists {
		return "", fmt.Errorf("the energy %s already exists", energyId)
	}
	
	bid := Energy{
		DocType:          "bid",
		ID:               bidId,
		EnergyID: energyId,
		Latitude:         energy.Latitude,
		Longitude:        energy.Longitude,
		Owner:         	  bidder,
		Producer:         energy.Producer,
		LargeCategory:    energy.LargeCategory,
		SmallCategory:    energy.SmallCategory,
		BidAmount: amount,
		Status:           "bid",
		UnitPrice:        energy.UnitPrice,
		BidPrice:         bidPrice,
		Priority: 		  priority, 
		GeneratedTime: energy.GeneratedTime,
		BidTime: timestamp,
	}

	bidJSON, err := json.Marshal(bid)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(bidId, bidJSON)
	if err != nil {
		return "", err
	}

	return "your bid was successful", nil
}

func (s *SmartContract) ChangeToken(ctx contractapi.TransactionContextInterface, energy *Energy) (error) {
	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(energy.ID, energyJSON)
	if err != nil {
		return err
	}
	return nil
}

func (s *SmartContract) AuctionEnd(ctx contractapi.TransactionContextInterface, energyId string, timestamp int64) (string, error) {
	energy, err := s.ReadToken(ctx, energyId)
	if err != nil {
		return "", err
	}

	startTime := timestamp - 2 * auctionInterval * 60
	endTime := timestamp - 1

	bidding, err := s.QueryAuctionEnd(ctx, energyId, startTime, endTime)
	if err != nil {
		return "", err
	}

	sort.SliceStable(bidding, func(i, j int) bool {
		return bidding[i].BidTime < bidding[j].BidTime
	})

	sort.SliceStable(bidding, func(i, j int) bool {
		return bidding[i].Priority > bidding[j].Priority
	})

	sort.SliceStable(bidding, func(i, j int) bool {
		return bidding[i].BidPrice > bidding[j].BidPrice
	})

	generatedTimeCompare := timestamp - 60 * tokenLife

	totalAmount := energy.Amount - energy.SoldAmount
	successList := []*Energy{}
	var returnMessage string

	if generatedTimeCompare  >= energy.GeneratedTime && len(bidding) == 0 {
		energy.Status = "old"
		returnMessage = "the energy was generated more than 30min ago. This was not sold."
	} else {
		for _, b := range bidding {
			if (totalAmount > 0 && b.BidAmount >= totalAmount) {
				b.Amount = totalAmount
				totalAmount = 0
				b.Status = "success"
				successList  = append(successList, b)
			} else if (totalAmount > 0 && b.BidAmount < totalAmount) {
				b.Amount = b.BidAmount
				totalAmount -= b.Amount
				b.Status = "success"
				successList  = append(successList, b)
			} 
		}
	}

	energy.SoldAmount = energy.Amount - totalAmount
	if totalAmount == 0 {
		energy.Status = "sold"
		returnMessage = "auction ended"
	} else {
		returnMessage = "auction continues"
	}

	err = s.UpdateToken(ctx, energy)
	if err != nil {
		return "", err
	}

	for _, bid := range successList {
		err = s.UpdateToken(ctx, bid)
		if err != nil {
			return "", err
		}
	}

	for _, bid := range bidding {
		returnMessage += fmt.Sprintf(" %v\n", bid)
	}
	return returnMessage, nil
}

// AssetExists returns true when asset with given ID exists in world state
// スタブの意味はよく分からない。台帳にアクセスするための関数らしい。一般的には「外部プログラムとの細かなインターフェース制御を引き受けるプログラム」を指すらしい
func (s *SmartContract) EnergyExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	energyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return energyJSON != nil, nil
}

// ReadToken returns the asset stored in the world state with given id.
// トークンを返す
func (s *SmartContract) ReadToken(ctx contractapi.TransactionContextInterface, id string) (*Energy, error) {
	energyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if energyJSON == nil {
		return nil, fmt.Errorf("the energy %s does not exist", id)
	}

	var energy Energy
	err = json.Unmarshal(energyJSON, &energy)
	if err != nil {
		return nil, err
	}

	return &energy, nil
}

func (s *SmartContract) UpdateToken(ctx contractapi.TransactionContextInterface, energy *Energy) error {
	energyJSON, err := json.Marshal(energy)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(energy.ID, energyJSON)
}


func (s *SmartContract) QueryAuctionEnd(ctx contractapi.TransactionContextInterface, energyId string, startTime int64, endTime int64) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"bid","Status":"bid","EnergyID":"%s","Bid Time":{"$gte":%d,"$lte":%d}},
	"use_index":["_design/indexAuctionEndDoc","indexAuctionEnd"]}`, energyId, startTime, endTime)
	// queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByStatus(ctx contractapi.TransactionContextInterface, status string, owner string) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Status":"%s","Owner":{"$ne":"%s"}},
	"use_index":["_design/indexStatusDoc","indexStatus"]}`, status, owner)
	// queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByTime(ctx contractapi.TransactionContextInterface, start int64, end int64) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Generated Time":{"$gte":%d,"$lte":%d}},
	"use_index":["_design/indexTimeDoc","indexTime"]}`, start, end)
	// queryString := fmt.Sprintf(`{"selector":{"docType":"asset","owner":"%s"}}`, owner)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByUserAndBidTime(ctx contractapi.TransactionContextInterface, owner string, status string, startTime int64, endTime int64) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Owner":"%s", "Status":"%s","Bid Time":{"$gte":%d,"$lte":%d}},
	"use_index":["_design/indexUserAndBidTimeDoc","indexUserAndBidTime"]}`, owner, status, startTime, endTime)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByUserAndGeneratedTime(ctx contractapi.TransactionContextInterface, producer string, timestamp int64) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Producer":"%s","Generated Time":%d},
	"use_index":["_design/indexUserAndGeneratedTimeDoc","indexUserAndGeneratedTime"]}`, producer, timestamp)
	
	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByUserAndTime(ctx contractapi.TransactionContextInterface, producer string, startTime int64, endTime int64) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Producer":"%s","Generated Time":{"$gte":%d,"$lte":%d}},
	"use_index":["_design/indexUserAndGeneratedTimeDoc","indexUserAndGeneratedTime"]}`, producer, startTime, endTime)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}


func (s *SmartContract) QueryByUserAndStatus(ctx contractapi.TransactionContextInterface, owner string, status string) ([]*Energy, error) {
	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Owner":"%s", "Status":"%s"},
	"use_index":["_design/indexUserAndStatusDoc","indexUserAndStatus"]}`, owner, status)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) QueryByLocationRange(ctx contractapi.TransactionContextInterface,
	status string, owner string, latitudeLowerLimit float64, latitudeUpperLimit float64,
	longitudeLowerLimit float64, longitudeUpperLimit float64) ([]*Energy, error) {

	queryString := fmt.Sprintf(`{"selector":{"DocType":"token","Status":"%s", "Owner":{"$ne":"%s"},
	"Latitude":{"$gte":%f,"$lte":%f},"Longitude":{"$gte":%f,"$lte":%f}}, "use_index":["_design/indexLocationDoc","indexLocation"]}`,
		status, owner, latitudeLowerLimit, latitudeUpperLimit, longitudeLowerLimit, longitudeUpperLimit)

	energies, err := s.Query(ctx, queryString)

	return energies, err
}

func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, queryString string) ([]*Energy, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var energies []*Energy
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var energy Energy
		err = json.Unmarshal(queryResponse.Value, &energy)
		if err != nil {
			return nil, err
		}
		energies = append(energies, &energy)
	}

	return energies, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]*Energy, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var energies []*Energy
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var energy Energy
		err = json.Unmarshal(queryResponse.Value, &energy)
		if err != nil {
			return nil, err
		}
		energies = append(energies, &energy)
	}

	return energies, nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.EnergyExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the energy %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
