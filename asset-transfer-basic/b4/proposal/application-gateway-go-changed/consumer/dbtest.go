package consumer
import (
	"database/sql"
	"fmt"
	"log"
	"time"
	//"os"

	_ "github.com/mattn/go-sqlite3"
)


func main() {
	/*_, err := os.Stat("db")
	if err != nil {
		panic(err)
	}
	fmt.Println("../db")
	_, err = os.Stat("../db/test.db")
	if err != nil {
		panic(err)
	}
	fmt.Println("../db/test.db")*/


	db, err :=  sql.Open("sqlite3", "db/test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "Data" ("ID" INTEGER PRIMARY KEY, "UserName" TEXT, "Latitude" REAL, "Longitude" REAL, "TotalAmountWanted" REAL, 
	"FirstBidTime" INTEGER, "LastBidTime" INTEGER, "BatteryLife" REAL, "Requested" REAL, "BidAmount" REAL, "BidSolar" REAL, "BidWind" REAL, "BidThermal" REAL, 
	"GetAmount" REAL, "GetSolar" REAL, "GetWind" REAL, "GetThermal" REAL)`)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare(`INSERT INTO Data (ID, UserName, Latitude, Longitude, TotalAmountWanted, FirstBidTime, LastBidTime, BatteryLife, Requested, BidAmount, BidSolar, 
		BidWind, BidThermal, GetAmount, GetSolar, GetWind, GetThermal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d%d", time.Now().Unix(), i)
		username := fmt.Sprintf("user%d", i)
		lat := float64(i) * 1.1
		lon := float64(i) * 1.1
		wanted := float64(i) * 10000.111
		first := i * 20000
		last := i * 20000
		life := float64(i) * 0.9
		req := float64(i) * 1000.1
		amount := float64(i) * 500.8
		solar := float64(i) * 200.8
		wind := float64(i) * 300

		_, err := stmt.Exec(id, username, lat, lon, wanted, first, last, life, req, amount, solar, wind, 0, amount, solar, wind, 0)
		if err != nil {
			panic(err)
		}
	}
	isOk := true
	if isOk {
		tx.Commit()
	} else {
		tx.Rollback()
	}


	rows, err := db.Query(
		`SELECT * FROM Data`,
	)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var d Data
		err := rows.Scan(&d.ID, &d.UserName, &d.Latitude, &d.Longitude, &d.TotalAmountWanted, &d.FirstBidTime, &d.LastBidTime, &d.BatteryLife, 
			&d.Requested, &d.BidAmount, &d.BidSolar, &d.BidWind, &d.BidThermal, &d.GetAmount, &d.GetSolar, &d.GetWind, &d.GetThermal)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(d)
	}
}
