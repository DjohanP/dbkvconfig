package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/DjohanP/dbkvconfig"
)

// UserConfig testing config with json
type UserConfig struct {
	MaxAge          int         `json:"max_age"`
	MinAge          int         `json:"min_age"`
	InformationUser Information `json:"information"`
}

// Information testing config with nested json
type Information struct {
	HasCar bool `json:"has_car"`
}

func main() {
	master, err := sqlx.Open("postgres", "")
	if err != nil {
		log.Fatalln("Failed Connect DB Master", err)
	}
	if err := master.Ping(); err != nil {
		log.Fatalln("Failed Connect DB Master", err)
	}

	slave, err := sqlx.Open("postgres", "")
	if err != nil {
		log.Fatalln("Failed Connect DB Slave", err)
	}
	if err := slave.Ping(); err != nil {
		log.Fatalln("Failed Connect DB Slave", err)
	}
	master.SetMaxOpenConns(10)
	slave.SetMaxOpenConns(10)

	redisPool := &redis.Pool{
		MaxIdle:   50,
		MaxActive: 10000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")
			// Connection error handling
			if err != nil {
				log.Printf("ERROR: fail initializing the redis pool: %s", err.Error())
				os.Exit(1)
			}
			return conn, err
		},
	}
	dbKV, err := dbkvconfig.NewDBKVConfig([]dbkvconfig.KeyValueInit{
		dbkvconfig.KeyValueInit{
			KeyConfig:    "key1",
			DefaultValue: &UserConfig{},
		},
		dbkvconfig.KeyValueInit{
			KeyConfig:    "key2",
			DefaultValue: 0,
		},
		dbkvconfig.KeyValueInit{
			KeyConfig:    "key3",
			DefaultValue: int64(0),
		}, dbkvconfig.KeyValueInit{
			KeyConfig:    "key4",
			DefaultValue: false,
		}}, master, slave, redisPool, dbkvconfig.OptionsCfg{
		TableName:       "config_kv",
		NameApplication: "AppTestHTTP",
		WatchInterval:   100,
		TTLRedis:        100,
		UseRedisPubSub:  true,
	})
	if err != nil {
		log.Fatalln("Failed Init DBKVLib", err)
	}
	handler := HTTPHandler{
		dbkvlib: dbKV,
	}

	http.HandleFunc("/config/get", handler.GetConfigHandler)
	http.HandleFunc("/config/update", handler.UpdateConfigHandler)
	http.HandleFunc("/config/insert", handler.InsertConfigHandler)
	http.HandleFunc("/check/eligible/person", handler.CheckEligibleUserHandler)
	http.HandleFunc("/check/method", handler.CheckMethodHandler)

	port := ":9090"
	log.Println("Start Serve", port)
	log.Fatalln(http.ListenAndServe(port, nil))
}
