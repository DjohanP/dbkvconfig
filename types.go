package dbkvconfig

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// DBKVCfg struct library DBKV Config
type DBKVCfg struct {
	// Storage Library
	db           *database
	redisWrapper *redigoWrapper

	// Saving config store
	valueStore map[string]keyValueStore

	// Redis Config
	watchInterval    int // In ms, to listen redis interval checking
	ttlRedis         int
	useRedisPubSub   bool
	keyRedisListener string // Key redis to known had changes on DB

	// To get info when reloading config changes
	nameApplication string

	// Database Config
	keyNameColumn   string
	valueNameColumn string
	tableName       string

	// Config to Get Last Update Application
	lastUpdateUnix int64
}

// OptionsCfg Config for user preferences
type OptionsCfg struct {
	WatchInterval    int
	TTLRedis         int
	UseRedisPubSub   bool
	NameApplication  string
	KeyRedisListener string
	KeyNameColumn    string
	ValueNameColumn  string
	TableName        string
}

type database struct {
	dbConnSlave       *sqlx.DB
	dbConnMaster      *sqlx.DB
	queryGetByKey     *sqlx.Stmt
	queryInsertConfig *sqlx.Stmt
	queryUpdateConfig *sqlx.Stmt
}

type redigoWrapper struct {
	pool *redis.Pool
}

// KeyValueInit from User
// User must set default value ex. for int using 0 or string using ""
// If config with struct must using struct json with pointer
type KeyValueInit struct {
	KeyConfig    string
	DefaultValue interface{}
}

type keyValueStore struct {
	typeConfig string
	value      interface{}
}

type dataDBConfig struct {
	Key   string `db:"key"`
	Value string `db:"value"`
}

type lastUpdate struct {
	fieldChanges        string
	timeUnixLastChanges int64
}
