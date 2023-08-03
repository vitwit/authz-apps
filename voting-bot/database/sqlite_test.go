package database

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidators(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS validators (chainName VARCHAR PRIMARY KEY, address VARCHAR )")
	if err != nil {
		fmt.Println(err)
		t.Fatalf("Failed to create test table: %v", err)
	}

	sqlitedb := &Sqlitedb{db: db}

	err = sqlitedb.AddValidator("chain1", "cosmos1...")
	assert.NoError(t, err)

	exists := sqlitedb.HasValidator("cosmos1...")
	assert.NoError(t, err)

	expectedRes := true
	assert.Equal(t, expectedRes, exists)

	log, err := sqlitedb.GetValidatorAddress()
	assert.NoError(t, err)
	assert.Len(t, log, 1)

	expectedAddress := []string{"cosmos1..."}
	assert.Equal(t, expectedAddress, log)
	val, err := sqlitedb.GetChainValidator("chain1")
	assert.NoError(t, err)

	expectedVal := "cosmos1..."
	assert.Equal(t, expectedVal, val)

	logs, err := sqlitedb.GetValidators()
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	expectedLog := Validator{
		ChainName: "chain1",
		Address:   "cosmos1...",
	}
	assert.Equal(t, expectedLog, logs[0])
	sqlitedb.RemoveValidator("cosmos1...")
	logs, err = sqlitedb.GetValidators()
	assert.NoError(t, err)

	expectedLen := 0
	assert.Equal(t, expectedLen, len(logs))
}

func TestKeys(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS keys (chainName VARCHAR PRIMARY KEY, keyName VARCHAR, keyAddress VARCHAR)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}
	db.Exec("ALTER TABLE keys ADD COLUMN authzStatus VARCHAR DEFAULT 'false'")

	sqlitedb := &Sqlitedb{db: db}

	err = sqlitedb.AddKey("chain1", "keyname", "keyaddress")
	assert.NoError(t, err)
	log, err := sqlitedb.GetKeyAddress("keyname")
	assert.NoError(t, err)

	expectedAddress := "keyaddress"
	assert.Equal(t, expectedAddress, log)
	key, err := sqlitedb.GetChainKey("chain1")
	assert.NoError(t, err)

	expectedKey := "keyname"
	assert.Equal(t, expectedKey, key)

	logs, err := sqlitedb.GetKeys()
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	expectedLog := keys{
		ChainName:  "chain1",
		KeyName:    "keyname",
		KeyAddress: "keyaddress",
		Status:     "false",
	}
	assert.Equal(t, expectedLog, logs[0])
}

func TestVoteLogs(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS logs (date INTEGER, chainName VARCHAR, proposalId VARCHAR, voteOption VARCHAR)")
	if err != nil {
		fmt.Println(err)
		t.Fatalf("Failed to create test table: %v", err)
	}
	_, err = db.Exec("ALTER TABLE logs ADD COLUMN proposalTitle VARCHAR")
	if err != nil {
		t.Fatalf("Failed to alter test table: %v", err)
	}
	sqlitedb := &Sqlitedb{db: db}

	err = sqlitedb.AddLog("chain1", "proposaltitle", "proposal1", "yes")
	assert.NoError(t, err)

	logs, err := sqlitedb.GetVoteLogs("chain1", "2007-04-01", "")
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	expectedLog := voteLogs{
		Date:          time.Now().UTC().Unix(),
		ChainName:     "chain1",
		ProposalTitle: "proposaltitle",
		ProposalID:    "proposal1",
		VoteOption:    "yes",
	}
	assert.Equal(t, expectedLog, logs[0])

	err = sqlitedb.AddLog("chain1", "proposaltitle2", "proposal2", "")
	assert.NoError(t, err)

	err = sqlitedb.UpdateVoteLog("chain1", "proposal2", "no")
	assert.NoError(t, err)

	// vote on non-existing proposal, this will not be stored in logs
	err = sqlitedb.UpdateVoteLog("chain1", "proposal3", "yes")
	assert.NoError(t, err)

	logs, err = sqlitedb.GetVoteLogs("chain1", "2007-04-01", "")
	assert.NoError(t, err)
	assert.Len(t, logs, 2)

	expectedLogs := []voteLogs{
		{
			Date:          time.Now().UTC().Unix(),
			ChainName:     "chain1",
			ProposalTitle: "proposaltitle",
			ProposalID:    "proposal1",
			VoteOption:    "yes",
		},
		{
			Date:          time.Now().UTC().Unix(),
			ChainName:     "chain1",
			ProposalTitle: "proposaltitle2",
			ProposalID:    "proposal2",
			VoteOption:    "no",
		},
	}
	assert.Equal(t, expectedLogs, logs)
}
