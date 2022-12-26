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

func BidResult(contract *client.Contract, userName string, auctionStartMin time.Time) {

}