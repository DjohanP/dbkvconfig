package dbkvconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// NewDBKVConfig Init Library support slave master DB
func NewDBKVConfig(listKey []KeyValueInit, master, slave *sqlx.DB, redisPool *redis.Pool, opts ...OptionsCfg) (*DBKVCfg, error) {
	var err error
	newDBKVConfig := new(DBKVCfg)

	newDBKVConfig.watchInterval = 10000 // Default 10000 ms
	newDBKVConfig.ttlRedis = 300        // Default 300 s
	newDBKVConfig.nameApplication = DefaultApplication
	newDBKVConfig.keyRedisListener = KeyRedisDefault
	newDBKVConfig.valueStore = make(map[string]keyValueStore)

	redis := new(redigoWrapper)
	redis.pool = redisPool
	newDBKVConfig.redisWrapper = redis

	newDBKVConfig.keyNameColumn = "key"
	newDBKVConfig.valueNameColumn = "value"
	newDBKVConfig.tableName = "config_kv"

	if len(opts) > 0 {
		if opts[0].WatchInterval > 0 {
			newDBKVConfig.watchInterval = opts[0].WatchInterval
		}
		if opts[0].TTLRedis > 0 {
			newDBKVConfig.ttlRedis = opts[0].TTLRedis
		}
		newDBKVConfig.useRedisPubSub = opts[0].UseRedisPubSub
		if opts[0].NameApplication != "" {
			newDBKVConfig.nameApplication = opts[0].NameApplication
		}
		if opts[0].KeyRedisListener != "" {
			newDBKVConfig.keyRedisListener = opts[0].KeyRedisListener
		}
		if opts[0].KeyNameColumn != "" {
			newDBKVConfig.keyNameColumn = opts[0].KeyNameColumn
		}
		if opts[0].ValueNameColumn != "" {
			newDBKVConfig.valueNameColumn = opts[0].ValueNameColumn
		}
		if opts[0].TableName != "" {
			newDBKVConfig.tableName = opts[0].TableName
		}
	}

	// Prepared Queries
	database := new(database)
	database.dbConnMaster = master
	database.dbConnSlave = slave
	queriesSelectByKey := fmt.Sprintf(getConfigByKey, newDBKVConfig.valueNameColumn, newDBKVConfig.tableName, newDBKVConfig.keyNameColumn)
	database.queryGetByKey, err = database.dbConnSlave.Preparex(queriesSelectByKey)
	if err != nil {
		return nil, err
	}

	queriesInsertConfig := fmt.Sprintf(insertConfigQuery, newDBKVConfig.tableName, newDBKVConfig.keyNameColumn, newDBKVConfig.valueNameColumn)
	database.queryInsertConfig, err = database.dbConnMaster.Preparex(queriesInsertConfig)
	if err != nil {
		return nil, err
	}

	queriesUpdateConfig := fmt.Sprintf(updateConfigQuery, newDBKVConfig.tableName, newDBKVConfig.valueNameColumn, newDBKVConfig.keyNameColumn)
	database.queryUpdateConfig, err = database.dbConnMaster.Preparex(queriesUpdateConfig)
	if err != nil {
		return nil, err
	}

	newDBKVConfig.db = database

	// Build Config from DB
	for _, key := range listKey {
		newDBKVConfig.valueStore[key.KeyConfig] = keyValueStore{
			value: key.DefaultValue,
		}
		err := newDBKVConfig.reloadConfig(key.KeyConfig)
		if err != nil {
			return nil, fmt.Errorf("Failed init config %s ErrorMessage:%s", key.KeyConfig, err.Error())
		}
	}
	if newDBKVConfig.useRedisPubSub {
		newDBKVConfig.subscribeRedisConfig()
	} else {
		newDBKVConfig.listenRedisConfig()
	}

	return newDBKVConfig, nil
}

// reloadConfig to get data from db and set config
func (dbkv *DBKVCfg) reloadConfig(key string) error {
	valueDB, err := dbkv.db.getValueConfig(key)
	if err != nil {
		return err
	}
	defer func() {
		dbkv.lastUpdateUnix = time.Now().Unix()
	}()

	if val, ok := dbkv.valueStore[key]; ok {
		if val.value != nil {
			switch val.value.(type) {
			case string:
				val.value = valueDB
			case int64:
				val.value, err = strconv.ParseInt(valueDB, 10, 64)
			case int:
				val.value, err = strconv.Atoi(valueDB)
			case bool:
				val.value, err = strconv.ParseBool(valueDB)
			default: // custom type from user in this library must use struct json only
				typeVal := reflect.TypeOf(val.value).Elem()
				v := reflect.New(typeVal)
				newVal := v.Interface()
				err = json.Unmarshal([]byte(valueDB), newVal)
				if err != nil {
					return err
				}
				val.value = newVal
			}
			if err != nil {
				return err
			}
			dbkv.valueStore[key] = val
		}

	}
	return nil
}

// subscribeRedisConfig To listen from redis an reload config
func (dbkv *DBKVCfg) subscribeRedisConfig() error {
	rc := dbkv.redisWrapper.pool.Get()
	pubSub := redis.PubSubConn{Conn: rc}
	if err := pubSub.PSubscribe(dbkv.keyRedisListener); err != nil {
		return err
	}
	go func() {
		for {
			switch v := pubSub.Receive().(type) {
			case redis.Message:
				keyUpdate := string(v.Data)
				if keyUpdate == "" {
					continue
				}
				// Get DB Config and construct config
				err := dbkv.reloadConfig(keyUpdate)
				if err != nil {
					log.Printf("DBKVConfig Fail Reload Config Name Application:%s KeyUpdate:%s Error:%s\n", dbkv.nameApplication, keyUpdate, err.Error())
					continue
				}
				log.Printf("DBKVConfig Success Reload Config Name Application:%s KeyUpdate:%s\n", dbkv.nameApplication, keyUpdate)
			}
		}
	}()
	return nil
}

// listenRedisConfig To listen from redis an reload config
func (dbkv *DBKVCfg) listenRedisConfig() {
	var (
		ticker = time.NewTicker(time.Duration(dbkv.watchInterval) * time.Millisecond)
	)
	go func() {
		log.Println("InitGan")
		defer ticker.Stop()
		for {
			// Waiting interval
			select {
			case <-ticker.C:
			}
			dataLastUpdate, err := dbkv.redisWrapper.getLastUpdate(dbkv.keyRedisListener)
			if err != nil { // Don't change anything
				if err != redis.ErrNil {
					log.Printf("DBKVConfig Fail Get Data from Redis Listener Name Application:%s Errror:%s\n", dbkv.nameApplication, err.Error())
				}
				continue
			}
			if dataLastUpdate.fieldChanges == "" || dataLastUpdate.timeUnixLastChanges < dbkv.lastUpdateUnix {
				continue
			}
			// Get DB Config and construct config
			err = dbkv.reloadConfig(dataLastUpdate.fieldChanges)
			if err != nil {
				log.Printf("DBKVConfig Fail Reload Config Name Application:%s KeyUpdate:%s Error:%s\n", dbkv.nameApplication, dataLastUpdate.fieldChanges, err.Error())
				continue
			}
			log.Printf("DBKVConfig Success Reload Config Name Application:%s KeyUpdate:%s\n", dbkv.nameApplication, dataLastUpdate.fieldChanges)
		}
	}()
}

// GetConfig func to get config return interface
func (dbkv *DBKVCfg) GetConfig(key string) (interface{}, error) {
	store, ok := dbkv.valueStore[key]
	if !ok {
		return nil, fmt.Errorf(ConfigNotFoundMsg)
	}
	return store.value, nil
}

// InsertConfig insert new config value without publish message
func (dbkv *DBKVCfg) InsertConfig(key, value string) error {
	return dbkv.db.insertValueConfig(key, value)
}

// UpdateConfig validating first after that update on DB and set to redis
func (dbkv *DBKVCfg) UpdateConfig(key, value string) error {
	valOld, err := dbkv.GetConfig(key)
	if err != nil {
		return err
	}

	// Prevent race condition and don't rapdily changes
	if dbkv.useRedisPubSub {
		if time.Now().Unix() < dbkv.lastUpdateUnix+5 {
			return fmt.Errorf(UpdateTooFast)
		}
	} else {
		dataLastUpdate, err := dbkv.redisWrapper.getLastUpdate(dbkv.keyRedisListener)
		if err != nil {
			return err
		}
		if dataLastUpdate.timeUnixLastChanges != 0 && time.Now().Unix() < dataLastUpdate.timeUnixLastChanges+int64(dbkv.watchInterval)+5 {
			return fmt.Errorf(UpdateTooFast)
		}
	}

	// Validating val
	switch valOld.(type) {
	case string:
	case int64:
		_, err = strconv.ParseInt(value, 10, 64)
	case int:
		_, err = strconv.Atoi(value)
	case bool:
		_, err = strconv.ParseBool(value)
	default: // custom type from user in this library must use struct json only
		typeVal := reflect.TypeOf(valOld).Elem()
		v := reflect.New(typeVal)
		newVal := v.Interface()
		err = json.Unmarshal([]byte(value), newVal)
	}
	if err != nil {
		return err
	}
	err = dbkv.db.updateConfigQuery(key, value)
	if err != nil {
		return err
	}

	if dbkv.useRedisPubSub {
		err = dbkv.redisWrapper.publishLastUpdate(dbkv.keyRedisListener, key)
	} else {
		err = dbkv.redisWrapper.setLastUpdate(dbkv.keyRedisListener, lastUpdate{
			fieldChanges:        key,
			timeUnixLastChanges: time.Now().Unix(),
		}, dbkv.ttlRedis)
	}

	return err
}
