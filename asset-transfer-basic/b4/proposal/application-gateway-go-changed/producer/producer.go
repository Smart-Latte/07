package producer

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path"
	"time"
	"sync"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID         = "Org1MSP"
	cryptoPath    = "../../../../test-network/organizations/peerOrganizations/org1.example.com"
	certPath      = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/cert.pem"
	keyPath       = cryptoPath + "/users/User1@org1.example.com/msp/keystore/"
	tlsCertPath   = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint  = "localhost:7051"
	gatewayPeer   = "peer0.org1.example.com"
	channelName   = "mychannel"
	chaincodeName = "basic"
)

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

type Input struct {
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	User            string    `json:"user"`
	Amount float64 `json:"amount"`
	Category string `json:"category"`
	Timestamp int64 `json:"timestamp"`
}

type EndInput struct {
	ID string `json:"ID"`
	Amount float64 `json:"Amount"`
	Time int64 `json:"Time"`
}

const (
	dayNum = 2
	hourNum = 24
)

var StartTime int64
var EndTime int64
var Diff int64
var Speed int64
var Interval int64
var TokenLife int64
var StartHour int
var SolarOutput [dayNum][hourNum]float64
var WindOutput [dayNum][hourNum] float64
var SeaWindOutput [dayNum][hourNum] float64

func AllProducers(start int64, end int64, difference int64, mySpeed int64, auctionInterval int64, life int64, sOutput [dayNum][hourNum]float64, wOutput [dayNum][hourNum]float64, 
	swOutput [dayNum][hourNum]float64, hour int) {
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
	Diff = difference
	Speed = mySpeed
	Interval = auctionInterval
	TokenLife = life
	StartHour = hour
	SolarOutput = sOutput
	WindOutput = wOutput
	SeaWindOutput = swOutput

	var thermalOutput[dayNum][hourNum]  float64 =[dayNum][hourNum]float64 {
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}

	var wg sync.WaitGroup

	wg.Add(5)
	go func() {
		defer wg.Done()
		Produce(contract, "real-solar-producer0", 40.2297629645958, 140.010266575019, "solar", 1000000, SolarOutput, 0)
	}()
	go func() {
		defer wg.Done()
		SeaWindProducer(contract, "real-wind-producer0", 40.2160279724715, 140.002846271612, "wind", 1990000, 12.5, 2.5, SeaWindOutput, 1)
	}()
	go func() {
		defer wg.Done()
		SeaWindProducer(contract, "real-wind-producer1", 40.2095028757269, 139.997337258476, "wind", 1990000, 12.5, 2.5, SeaWindOutput, 2)
	}()
	go func() {
		defer wg.Done()
		SeaWindProducer(contract, "real-wind-producer2", 40.2021377588529, 140.068615482843, "wind", 1990000, 12.5, 2.5, SeaWindOutput, 3)
	}()
	go func() {
		defer wg.Done()
		Produce(contract, "real-thermal-producer0", 40.2021377588529, 140.068615482843, "thermal", 1000000, thermalOutput, 4)
	}()

	/*for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(n int) {
			// DummyWindProducer(contract, fmt.Sprintf("windProducerGroup%d", n), 40, 41, 140, 141, "wind", 1100, 12.5, 2.5, WindOutput, int64(n + 1000))
			DummyWindProducer(contract, fmt.Sprintf("windProducerGroup%d", n), 40.17463042136363, 40.1932732666231, 139.992165531859, 140.068615482843, "wind", 11000, 12.5, 2.5, WindOutput, int64(n + 1000))
			// DummySolarProducer(contract, fmt.Sprintf("solarProducerGroup%d", n), 40, 41, 140, 141, "solar", 4000, SolarOutput, int64(n + 10000))
			DummySolarProducer(contract, fmt.Sprintf("solarProducerGroup%d", n), 40, 41, 140, 141, "solar", 40000, SolarOutput, int64(n + 10000))

		}(i)
	}*/

	wg.Wait()

	fmt.Printf("all producer end\n")

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
