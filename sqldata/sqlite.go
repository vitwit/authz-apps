package sqldata

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var Id int

func ChainDataCreate() (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS chaindata (chain_name VARCHAR, validator_address VARCHAR)")
	if err != nil {
		panic(err)
	}
	return db

}
func ChainDataInsert(chain_name string, validator_address string) {

	db := ChainDataCreate()

	stmt, err := db.Prepare("INSERT INTO chaindata(chain_name, validator_address) values(?,?)")
	checkErr(err)

	res, err := stmt.Exec(chain_name, validator_address)
	checkErr(err)

	Id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(Id)
	db.Close()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
