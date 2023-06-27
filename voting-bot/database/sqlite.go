package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var Id int

type (
	Validator struct {
		ChainName string
		Address   string
	}

	keys struct {
		ChainName  string
		KeyName    string
		KeyAddress string
		Status     string
	}

	voteLogs struct {
		Date          int64
		ChainName     string
		ProposalTitle string
		ProposalID    string
		VoteOption    string
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
	_, err := a.db.Exec("CREATE TABLE IF NOT EXISTS validators (chainName VARCHAR PRIMARY KEY, address VARCHAR )")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS logs (date INTEGER ,chainName VARCHAR, proposalId VARCHAR, voteOption VARCHAR)")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("ALTER TABLE logs ADD COLUMN proposalTitle VARCHAR")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS keys (chainName VARCHAR PRIMARY KEY, keyName VARCHAR, keyAddress VARCHAR)")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("ALTER TABLE keys ADD COLUMN authzStatus VARCHAR DEFAULT 'false'")
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

// Updates authorization status
func (a *Sqlitedb) UpdateAuthzStatus(status, keyAddress string) error {
	stmt, err := a.db.Prepare("UPDATE keys SET authzStatus = ? WHERE keyAddress = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(status, keyAddress)
	return err
}

func (a *Sqlitedb) AddLog(chainName, proposalTitle, proposalID, voteOption string) error {
	stmt, err := a.db.Prepare("INSERT INTO logs(date, chainName, proposalTitle, proposalID, voteOption) values(?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(time.Now().UTC().Unix(), chainName, proposalTitle, proposalID, voteOption)
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

func (a *Sqlitedb) GetChainKey(ChainName string) (string, error) {
	var addr string
	stmt, err := a.db.Prepare("SELECT keyName FROM keys WHERE chainName=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(ChainName).Scan(&addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

func (a *Sqlitedb) GetChainValidator(ChainName string) (string, error) {
	var addr string
	stmt, err := a.db.Prepare("SELECT address FROM validators WHERE chainName=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(ChainName).Scan(&addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// Gets required data regarding keys
func (a *Sqlitedb) GetKeys() ([]keys, error) {
	log.Printf("Fetching keys...")

	rows, err := a.db.Query("SELECT chainName, keyName, keyAddress, authzStatus FROM keys")
	if err != nil {
		return []keys{}, err
	}
	defer rows.Close()

	var k []keys
	for rows.Next() {
		var data keys
		if err := rows.Scan(&data.ChainName, &data.KeyName, &data.KeyAddress, &data.Status); err != nil {
			return k, err
		}
		k = append(k, data)
	}
	if err = rows.Err(); err != nil {
		return k, err
	}

	return k, nil
}

// Gets required data regarding votes
func (a *Sqlitedb) GetVoteLogs(chainName, startDate, endDate string) ([]voteLogs, error) {
	log.Printf("Fetching votes...")
	layout := "2006-01-02"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return nil, err
	}
	var end int64
	if len(endDate) < 1 {
		end = time.Now().UTC().Unix()
	} else {
		end1, err := time.Parse(layout, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		end = end1.Unix()
	}
	if start.Unix() >= end {
		return nil, fmt.Errorf("start date is not valid as it is greater than end date")
	}
	query := "SELECT date, chainName, proposalTitle, proposalId, voteOption FROM logs WHERE chainName = ? AND date BETWEEN ? AND ? "
	rows, err := a.db.Query(query, chainName, start.Unix(), end)
	if err != nil {
		return []voteLogs{}, err
	}
	defer rows.Close()

	var k []voteLogs
	for rows.Next() {
		var data voteLogs
		if err := rows.Scan(&data.Date, &data.ChainName, &data.ProposalTitle, &data.ProposalID, &data.VoteOption); err != nil {
			return k, err
		}
		k = append(k, data)
	}
	if err = rows.Err(); err != nil {
		return k, err
	}

	return k, nil
}
