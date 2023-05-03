package sqldata

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Id int

type (
	keys struct {
		chain_name string
		key_name   string
	}

	chaindata struct {
		chain_name        string
		validator_address string
	}
)

func NewChainData() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS chaindata (chain_name VARCHAR, validator_address VARCHAR)")
	if err != nil {
		panic(err)
	}
	return db
}

func NewVotesData() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS votesdata (proposal_ID VARCHAR, validator_address VARCHAR, vote_option VARCHAR)")
	if err != nil {
		panic(err)
	}
	return db
}

func NewKeysData() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS keysdata (chain_name VARCHAR, key_name VARCHAR, key_address VARCHAR)")
	if err != nil {
		panic(err)
	}
	return db
}

func VotesDataInsert(proposal_ID, validator_address, vote_option string) {
	db := NewVotesData()

	stmt, err := db.Prepare("INSERT INTO votesdata(proposal_ID, validator_address, vote_option) values(?,?,?)")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(proposal_ID, validator_address, vote_option)
	checkErr(err)

	Id, err := res.LastInsertId()
	checkErr(err)

	fmt.Printf("Successfully added votes data to the db with ID: %v\n", Id)
	db.Close()
}

func InsertKey(chain_name, key_name, key_address string) {
	db := NewKeysData()

	stmt, err := db.Prepare("INSERT INTO keysdata(chain_name, key_name, key_address) values(?,?,?)")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(chain_name, key_name, key_address)
	checkErr(err)

	Id, err := res.LastInsertId()
	checkErr(err)

	fmt.Printf("Successfully added key to the db with ID: %v\n", Id)
	db.Close()
}

func GetKeyAddress(key string) (string, error) {
	db := NewKeysData()

	var addr string
	stmt, err := db.Prepare("SELECT key_address FROM keysdata WHERE key_name=?")
	checkErr(err)
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&addr)
	checkErr(err)

	return addr, nil
}

func ChainDataInsert(chain_name string, validator_address string) {
	db := NewChainData()

	stmt, err := db.Prepare("INSERT INTO chaindata(chain_name, validator_address) values(?,?)")
	checkErr(err)
	defer stmt.Close()

	res, err := stmt.Exec(chain_name, validator_address)
	checkErr(err)

	Id, err := res.LastInsertId()
	checkErr(err)

	fmt.Printf("Successfully added validator address and chain name to the db with ID: %v\n", Id)
	db.Close()
}

func ListKeys() ([]keys, error) {
	log.Printf("Fetching keys...")
	db := NewKeysData()

	rows, err := db.Query("SELECT chain_name, key_name FROM keysdata")
	checkErr(err)
	defer rows.Close()

	var k []keys
	for rows.Next() {
		var data keys
		if err := rows.Scan(&data.chain_name, &data.key_name); err != nil {
			return k, err
		}
		k = append(k, data)
	}
	if err = rows.Err(); err != nil {
		return k, err
	}

	return k, nil
}

func ChainDataList() ([]chaindata, error) {
	log.Printf("Getting list...")
	db := NewChainData()

	rows, err := db.Query("SELECT chain_name, validator_address FROM chaindata")
	checkErr(err)
	defer rows.Close()

	var data []chaindata
	for rows.Next() {
		var chain chaindata
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
		var chain chaindata
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
