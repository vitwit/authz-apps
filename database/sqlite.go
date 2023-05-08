package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Id int

type (
	Validator struct {
		ChainName string
		Address   string
	}

	keys struct {
		ChainName string
		KeyName   string
		Address   string
	}

	Sqlitedb struct {
		db *sql.DB
	}
)

// Opens connection to SQLite database
func NewDatabase() (*Sqlitedb, error) {
	db, err := sql.Open("sqlite3", "./slackbot.db")
	return &Sqlitedb{
		db: db,
	}, err
}

// Closes the connection
func (a *Sqlitedb) Close() error {
	return a.db.Close()
}

// Creates all the required tables in database
func (a *Sqlitedb) InitializeTables() error {
	_, err := a.db.Exec("CREATE TABLE IF NOT EXISTS validators (chainName VARCHAR, address VARCHAR PRIMARY KEY)")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS keys (chainName VARCHAR, keyName VARCHAR, keyAddress VARCHAR)")
	return err
}

// Stores validator information
func (s *Sqlitedb) AddValidator(name, address string) error {
	stmt, err := s.db.Prepare("INSERT INTO validators(chainName, address) values(?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(name, address)
	return err
}

// Removes validator information
func (s *Sqlitedb) RemoveValidator(address string) error {
	stmt, err := s.db.Prepare("DELETE FROM validators WHERE address=?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(address)
	return err
}

// Stores Keys information
func (a *Sqlitedb) AddKey(chainName, keyName, keyAddress string) error {
	stmt, err := a.db.Prepare("INSERT INTO keys(chainName, keyName, keyAddress) values(?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(chainName, keyName, keyAddress)
	return err
}

// Checks whether the validator already exists in the database
func (s *Sqlitedb) HasValidator(validatorAddress string) bool {
	stmt, err := s.db.Prepare("SELECT EXISTS(SELECT 1 FROM validators WHERE address = ?)")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()

	var exists bool
	err = stmt.QueryRow(validatorAddress).Scan(&exists)
	if err != nil {
		log.Println(err)
	}
	return exists
}

// Gets all the stored validators data
func (s *Sqlitedb) GetValidators() ([]Validator, error) {
	rows, err := s.db.Query("SELECT chainName, address FROM validators")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var validators []Validator
	for rows.Next() {
		var validator Validator
		if err := rows.Scan(&validator.ChainName, &validator.Address); err != nil {
			return nil, err
		}
		validators = append(validators, validator)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return validators, nil
}

// Gets all the validator address stored in validators
func (s *Sqlitedb) GetValidatorAddress() (ValidatorAddress []string, err error) {
	rows, err := s.db.Query("SELECT address FROM validators")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, err
		}
		ValidatorAddress = append(ValidatorAddress, address)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ValidatorAddress, nil
}

// Gets Key address of a specific key
func (a *Sqlitedb) GetKeyAddress(key string) (string, error) {
	var addr string
	stmt, err := a.db.Prepare("SELECT keyAddress FROM keys WHERE keyName=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(key).Scan(&addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// Gets required data regarding keys
func (a *Sqlitedb) GetKeys() ([]keys, error) {
	log.Printf("Fetching keys...")

	rows, err := a.db.Query("SELECT chainName, keyName,keyAddress FROM keys")
	if err != nil {
		return []keys{}, err
	}
	defer rows.Close()

	var k []keys
	for rows.Next() {
		var data keys
		if err := rows.Scan(&data.ChainName, &data.KeyName, &data.Address); err != nil {
			return k, err
		}
		k = append(k, data)
	}
	if err = rows.Err(); err != nil {
		return k, err
	}

	return k, nil
}
