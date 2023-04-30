package sqldata

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Id int

func NewChainData() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS chaindata (chain_name VARCHAR, validator_address VARCHAR)")
	if err != nil {
		panic(err)
	}
	return db
}

func ChainDataInsert(chain_name string, validator_address string) {
	db := NewChainData()

	stmt, err := db.Prepare("INSERT INTO chaindata(chain_name, validator_address) values(?,?)")
	checkErr(err)

	res, err := stmt.Exec(chain_name, validator_address)
	checkErr(err)

	Id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(Id)
	db.Close()
}

type alldata struct {
	chain_name        string
	validator_address string
}

func ChainDataList() ([]alldata, error) {
	log.Printf("Getting list")
	db := NewChainData()

	rows, err := db.Query("SELECT chain_name, validator_address FROM chaindata")
	checkErr(err)
	defer rows.Close()

	var data []alldata
	for rows.Next() {
		var chain alldata
		if err := rows.Scan(&chain.chain_name, &chain.validator_address); err != nil {
			return data, err
		}
		data = append(data, chain)
	}
	if err = rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

func GetAllValAddrs() (valAddrs []string, err error) {
	log.Printf("Getting Validator addresses")
	db := NewChainData()

	rows, err := db.Query("SELECT validator_address FROM chaindata")
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var chain alldata
		if err := rows.Scan(&chain.validator_address); err != nil {
			return []string{}, err
		}
		valAddrs = append(valAddrs, chain.validator_address)
	}
	if err = rows.Err(); err != nil {
		return []string{}, err
	}
	return valAddrs, nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
