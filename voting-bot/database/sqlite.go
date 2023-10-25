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

	voteLogs struct {
		Date          int64
		ChainName     string
		ProposalTitle string
		ProposalID    string
		VoteOption    string
	}

	RewardsCommission struct {
		ChainID    string `json:"chainID"`
		Denom      string `json:"denom"`
		ValAddr    string `json:"val_addr"`
		Rewards    string `json:"rewards"`
		Commission string `json:"commission"`
		Date       string `json:"date"`
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

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS logs (date INTEGER, chainName VARCHAR, proposalId VARCHAR, voteOption VARCHAR, proposalTitle VARCHAR)")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS keys (chainName VARCHAR, keyName VARCHAR, granteeAddress VARCHAR, type VARCHAR, authzStatus VARCHAR DEFAULT 'false', PRIMARY KEY (chainName, type))")
	if err != nil {
		return err
	}

	_, err = a.db.Exec("CREATE TABLE IF NOT EXISTS income (chainId VARCHAR, denom VARCHAR, valAddress VARCHAR, rewards VARCHAR, commission VARCHAR, date VARCHAR)")
	if err != nil {
		return err
	}

	return nil
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

// Update vote logs information
func (a *Sqlitedb) UpdateVoteLog(chainName, proposalID, voteOption string) error {
	stmt, err := a.db.Prepare("UPDATE logs SET voteOption = ? WHERE chainName = ? AND proposalID = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(voteOption, chainName, proposalID)
	return err
}

// Adds vote logs information
func (s *Sqlitedb) AddLog(chainName, proposalTitle, proposalID, voteOption string) error {
	stmt, err := s.db.Prepare("SELECT EXISTS(SELECT 1 FROM logs WHERE chainName = ? AND proposalID = ?)")
	if err != nil {
		log.Println(err)
		return err
	}

	var exists bool
	err = stmt.QueryRow(chainName, proposalID).Scan(&exists)
	if err != nil {
		log.Println(err)
		return err
	}

	stmt.Close()

	if exists && voteOption != "" {
		stmt, err = s.db.Prepare("UPDATE logs SET date=?, proposalTitle=?, voteOption=? WHERE chainName=? AND proposalID=?")
		if err != nil {
			return err
		}

		defer stmt.Close()

		_, err = stmt.Exec(time.Now().UTC().Unix(), proposalTitle, voteOption, chainName, proposalID)
		return err
	} else {
		stmt, err = s.db.Prepare("INSERT INTO logs(date, chainName, proposalTitle, proposalID, voteOption) values(?,?,?,?,?)")
		if err != nil {
			return err
		}

		defer stmt.Close()

		_, err = stmt.Exec(time.Now().UTC().Unix(), chainName, proposalTitle, proposalID, voteOption)
		return err
	}

}

func (a *Sqlitedb) AddRewards(chainId, denom, valAddr, rewards, commission string) error {
	stmt, err := a.db.Prepare("INSERT INTO income(chainId, denom, valAddress, rewards, commission, date) values(?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(chainId, denom, valAddr, rewards, commission, time.Now().Format("2006-01-02"))
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

// Gets required data regarding votes
func (a *Sqlitedb) GetVoteLogs(chainName, startDate, endDate string) ([]voteLogs, error) {
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

// Gets required data regarding rewards
func (a *Sqlitedb) GetRewards(chainId, date string) ([]RewardsCommission, error) {
	log.Printf("Fetching rewards...")
	var k []RewardsCommission

	if date != "" {
		query := "SELECT chainId, denom, valAddress, rewards, commission, date FROM income WHERE chainId = ? AND date = ? ORDER BY date DESC"
		rows, err := a.db.Query(query, chainId, date)
		if err != nil {
			return []RewardsCommission{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var data RewardsCommission
			if err := rows.Scan(&data.ChainID, &data.Denom, &data.ValAddr, &data.Rewards, &data.Commission, &data.Date); err != nil {
				return k, err
			}
			k = append(k, data)
		}
		if err := rows.Err(); err != nil {
			return k, err
		}
	} else {
		query := "SELECT chainId, denom, valAddress, rewards, commission, date FROM income WHERE chainId = ? ORDER BY date DESC"
		rows, err := a.db.Query(query, chainId)
		if err != nil {
			return []RewardsCommission{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var data RewardsCommission
			if err := rows.Scan(&data.ChainID, &data.Denom, &data.ValAddr, &data.Rewards, &data.Commission, &data.Date); err != nil {
				return k, err
			}
			k = append(k, data)
		}
		if err := rows.Err(); err != nil {
			return k, err
		}
	}

	return k, nil
}
