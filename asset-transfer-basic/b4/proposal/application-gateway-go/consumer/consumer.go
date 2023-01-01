package consumer

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	//"log"
	"path"
	"time"
	//"net/http"
	//"encoding/json"
	//"bytes"
	"sync"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID         = "Org2MSP"
	cryptoPath    = "../../../../test-network/organizations/peerOrganizations/org2.example.com"
	certPath      = cryptoPath + "/users/User1@org2.example.com/msp/signcerts/cert.pem"
	keyPath       = cryptoPath + "/users/User1@org2.example.com/msp/keystore/"
	tlsCertPath   = cryptoPath + "/peers/peer0.org2.example.com/tls/ca.crt"
	peerEndpoint  = "localhost:9051"
	gatewayPeer   = "peer0.org2.example.com"
	channelName   = "mychannel"
	chaincodeName = "basic"
)

type Energy struct {
	DocType          string    `json:"DocType"`
	Amount float64 `json:"Amount"`
	UnitPrice        float64   `json:"Unit Price"`
	BidPrice         float64   `json:"Bid Price"`
	GeneratedTime    int64 `json:"Generated Time"`
	AuctionStartTime int64 `json:"Auction Start Time"`
	BidTime          int64 `json:"Bid Time"`
	ID               string    `json:"ID"`
	LargeCategory    string    `json:"LargeCategory"`
	Latitude         float64   `json:"Latitude"`
	Longitude        float64   `json:"Longitude"`
	Owner            string    `json:"Owner"`
	Producer         string    `json:"Producer"`
	Priority float64 `json:"Priority"`
	SmallCategory    string    `json:"SmallCategory"`
	Status           string    `json:"Status"`
	Error string `json:"Error"`
}

var startTime time.Time // シミュレーション開始時間
var auctionInterval time.Duration // オークション時間
var speed int // 何倍速か

var StartTime int64
var EndTime int64
var Diff int64
var Speed int64
var Interval int64
var TokenLife int64
var StartHour int

// ゴールーチンで各ユーザ起動
// input: シミュレーション開始時間
func AllConsumers(start int64, end int64, diff int64, auctionSpeed int64, auctionInterval int64, life int64, startHour int) {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	StartTime = start
	EndTime = end
	Diff = diff
	StartHour = startHour
	fmt.Println(StartTime)
	Speed = auctionSpeed
	Interval = auctionInterval
	TokenLife = life

	userData := [][]Data{}

	//40.23299059978385, 140.0137246267796
	//40.182882947704655, 140.0626533051892

	// 充電開始時間(差分)、バッテリー容量(Wh)、チャージ済み(Wh)、充電時間(hour)、最終的なバッテリー残量(0から1), seed
	var wg sync.WaitGroup

	for i := 0; i < 1; i++ {
		wg.Add(1)
		userData = append(userData, []Data{})
		go func(n int) {
			defer wg.Done()
			userData[n] = Consume(contract, fmt.Sprintf("consumer%d", n), 40.182882947704655, 40.23299059978385, 140.0137246267796, 140.0626533051892, 0, 66000, 1000, 12, 1, int64(n))
		}(i)
	}

	/*go func() {
		defer wg.Done()
		Produce(contract, "real-wind-producer1", 140.010266575019, 140.014538870921, "wind", 12000000, WindOutput, 2)
	}()*/
	wg.Wait()
	fmt.Println("all consumer end")
	for i := 0; i < len(userData); i++ {
		fmt.Printf("%s result:\n", userData[i][0].UserName)
		for _, user := range userData[i] {
			fmt.Printf("UserName:%s, Latitude:%v, Longitude:%v, TotalAmountWanted:%v, LatestAuctionStartTime:%v, FirstBidTime:%v, LastBidTime:%v, BatteryLife:%v, Requested:%v, BidAmount:%v, BidSolar:%v, BidWind:%v, BidThermal:%v, GetAmount:%v, GetSolar:%v, GetWind:%v, GetThermal:%v\n", user.UserName, user.Latitude, user.Longitude, user.TotalAmountWanted, 
		user.LatestAuctionStartTime, user.FirstBidTime, user.LastBidTime, user.BatteryLife, user.Requested, user.BidAmount, user.BidSolar, user.BidWind, user.BidThermal, user.GetAmount, 
	user.GetSolar, user.GetWind, user.GetThermal)
		}
		fmt.Println("")
	}
	
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, gatewayPeer)

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to create gRPC connection: %w", err))
	}

	return connection
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	files, err := ioutil.ReadDir(keyPath)
	if err != nil {
		panic(fmt.Errorf("failed to read private key directory: %w", err))
	}
	privateKeyPEM, err := ioutil.ReadFile(path.Join(keyPath, files[0].Name()))

	if err != nil {
		panic(fmt.Errorf("failed to read private key file: %w", err))
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
