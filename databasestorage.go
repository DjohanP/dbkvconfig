package dbkvconfig

// getValueConfig to get config from DB
func (db *database) getValueConfig(key string) (string, error) {
	var value string
	err := db.queryGetByKey.Get(&value, key)
	return value, err
}

// insertValueConfig for insert new config to DB
func (db *database) insertValueConfig(key, value string) error {
	_, err := db.queryInsertConfig.Exec(key, value)
	return err
}

// updateConfigQuery for update config to DB
func (db *database) updateConfigQuery(key, value string) error {
	_, err := db.queryUpdateConfig.Exec(key, value)
	return err
}
