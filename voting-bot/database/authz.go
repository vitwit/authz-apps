package database

type AuthzKeys struct {
	ChainName      string
	KeyName        string
	GranteeAddress string
	AuthzStatus    string
	Type           string
}

// Stores Keys information
func (a *Sqlitedb) AddAuthzKey(chainName, keyName, keyAddress, keyType string) error {
	stmt, err := a.db.Prepare("INSERT INTO keys(chainName, keyName, granteeAddress, type) values(?,?,?,?)")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(chainName, keyName, keyAddress, keyType)
	return err
}

// Updates authorization status
func (a *Sqlitedb) UpdateAuthzStatus(status, keyAddress, keyType string) error {
	stmt, err := a.db.Prepare("UPDATE keys SET authzStatus = ? WHERE granteeAddress = ? AND type = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(status, keyAddress, keyType)
	return err
}

// Gets Key address of a specific key
func (a *Sqlitedb) GetAuthzKeyAddress(name, keyType string) (string, error) {
	var addr string
	stmt, err := a.db.Prepare("SELECT granteeAddress FROM keys WHERE keyName=? AND type=?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(name, keyType).Scan(&addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

func (a *Sqlitedb) GetChainKey(chainName, keyType string) (string, error) {
	var addr string
	stmt, err := a.db.Prepare("SELECT keyName FROM keys WHERE chainName=? AND type = ?")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	err = stmt.QueryRow(chainName, keyType).Scan(&addr)
	if err != nil {
		return "", err
	}

	return addr, nil
}

// Gets required data regarding keys
func (a *Sqlitedb) GetKeys() ([]AuthzKeys, error) {
	rows, err := a.db.Query("SELECT chainName, keyName, granteeAddress, authzStatus, type FROM keys")
	if err != nil {
		return []AuthzKeys{}, err
	}
	defer rows.Close()

	var k []AuthzKeys
	for rows.Next() {
		var data AuthzKeys
		if err := rows.Scan(&data.ChainName, &data.KeyName, &data.GranteeAddress, &data.AuthzStatus, &data.Type); err != nil {
			return k, err
		}
		k = append(k, data)
	}
	if err = rows.Err(); err != nil {
		return k, err
	}

	return k, nil
}
