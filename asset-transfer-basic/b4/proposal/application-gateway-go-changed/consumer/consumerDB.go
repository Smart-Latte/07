package consumer

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func DbResister(dataList [][]Data) {
	db, err :=  sql.Open("sqlite3", "db/test1.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "ConsumerData" ("ID" INTEGER PRIMARY KEY, "UserName" TEXT, "Latitude" REAL, "Longitude" REAL, "TotalAmountWanted" REAL, 
	"FirstBidTime" INTEGER, "LastBidTime" INTEGER, "BatteryLife" REAL, "Requested" REAL, "BidAmount" REAL, "BidSolar" REAL, "BidWind" REAL, "BidThermal" REAL, 
	"GetAmount" REAL, "GetSolar" REAL, "GetWind" REAL, "GetThermal" REAL)`)
	if err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	stmt, err := tx.Prepare(`INSERT INTO ConsumerData (ID, UserName, Latitude, Longitude, TotalAmountWanted, FirstBidTime, LastBidTime, BatteryLife, Requested, BidAmount, BidSolar, 
		BidWind, BidThermal, GetAmount, GetSolar, GetWind, GetThermal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()


	for i := 0; i < len(dataList); i++ {
		for j := 0; j < len(dataList[i]); j++ {
			id := fmt.Sprintf("%d%s", time.Now().Unix(), dataList[i][j].UserName)
			_, err := stmt.Exec(id, dataList[i][j].UserName, dataList[i][j].Latitude, dataList[i][j].Longitude, dataList[i][j].TotalAmountWanted, dataList[i][j].FirstBidTime,
			dataList[i][j].LastBidTime, dataList[i][j].BidWind, dataList[i][j].BidThermal, dataList[i][j].GetAmount, dataList[i][j].GetSolar, dataList[i][j].GetWind, dataList[i][j].GetThermal)
			if err != nil {
				panic(err)
			}
		}

	}
	tx.Commit()
	//tx.Rollback()

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
