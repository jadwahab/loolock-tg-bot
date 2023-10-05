package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DBParams struct {
	DB *sql.DB
}

func (dbp *DBParams) Insert(btcinscnumber int, btctxid, bsvtxid, btcinscid string) error {
	// Prepare the SQL statement
	stmt, err := dbp.DB.Prepare("INSERT INTO reinks (created, modified, btcinscnumber, btctxid, bsvtxid, btcinscid) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer stmt.Close()

	// Set the values for the SQL statement
	created := time.Now()
	modified := time.Now()

	// Execute the SQL statement with the values
	_, err = stmt.Exec(created, modified, btcinscnumber, btctxid, bsvtxid, btcinscid)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Println("Inserted row into 'reinks' table successfully at index: " + fmt.Sprint(btcinscnumber))
	return nil
}

type Reink struct {
	Created       time.Time
	Modified      time.Time
	BTCInsCNumber int
	BTCTxID       string
	BSVTxID       string
	BTCInsCID     string
}

func (dbp *DBParams) GetReink(btcinscnumber int) (*Reink, error) {
	// Prepare the SQL statement
	stmt, err := dbp.DB.Prepare("SELECT created, modified, btcinscnumber, btctxid, bsvtxid, btcinscid FROM reinks WHERE btcinscnumber = $1")
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(btcinscnumber)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows of the result set and return the first result
	for rows.Next() {
		var reink Reink
		err := rows.Scan(&reink.Created, &reink.Modified, &reink.BTCInsCNumber, &reink.BTCTxID, &reink.BSVTxID, &reink.BTCInsCID)
		if err != nil {
			fmt.Println("Error:", err)
			return nil, err
		}
		// fmt.Printf("%+v\n", reink)
		return &reink, nil
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// If no results were found, return nil and an error
	return nil, fmt.Errorf("no results found for btcinscnumber %d", btcinscnumber)

}

func (dbp *DBParams) GetLatestRink() (int, error) {
	// Prepare the SQL statement
	stmt, err := dbp.DB.Prepare("select * from reinks order by btcinscnumber desc limit 1")
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}
	defer rows.Close()

	// Iterate over the rows of the result set and return the first result
	for rows.Next() {
		var reink Reink
		err := rows.Scan(&reink.Created, &reink.Modified, &reink.BTCInsCNumber, &reink.BTCTxID, &reink.BSVTxID, &reink.BTCInsCID)
		if err != nil {
			fmt.Println("Error:", err)
			return 0, err
		}
		// fmt.Printf("%+v\n", reink)
		return reink.BTCInsCNumber, nil
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	// If no results were found, return nil and an error
	return 0, fmt.Errorf("no results found")

}
