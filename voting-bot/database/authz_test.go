package database

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS keys (chainName VARCHAR, keyName VARCHAR, granteeAddress VARCHAR, type VARCHAR,authzStatus VARCHAR DEFAULT 'false', PRIMARY KEY (chainName,type))")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	sqlitedb := &Sqlitedb{db: db}

	err = sqlitedb.AddAuthzKey("chain1", "keyname", "keyaddress", "voting")
	assert.NoError(t, err)
	log, err := sqlitedb.GetAuthzKeyAddress("keyname", "voting")
	assert.NoError(t, err)

	expectedAddress := "keyaddress"
	assert.Equal(t, expectedAddress, log)
	key, err := sqlitedb.GetChainKey("chain1", "voting")
	assert.NoError(t, err)

	expectedKey := "keyname"
	assert.Equal(t, expectedKey, key)

	logs, err := sqlitedb.GetKeys()
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	expectedLog := AuthzKeys{
		ChainName:      "chain1",
		KeyName:        "keyname",
		GranteeAddress: "keyaddress",
		AuthzStatus:    "false",
		Type:           "voting",
	}
	assert.Equal(t, expectedLog, logs[0])
}
