package dbkvconfig

import (
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// getLastUpdate to get data last update and support config without pubsub
func (redigo *redigoWrapper) getLastUpdate(key string) (lastUpdate, error) {
	dataLastUpdate := lastUpdate{}
	conn := redigo.pool.Get()
	defer conn.Close()

	result, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil || len(result) == 0 {
		return dataLastUpdate, err
	}
	dataLastUpdate.fieldChanges = result[fieldLastChange]
	dataLastUpdate.timeUnixLastChanges, err = strconv.ParseInt(result[fieldLastChangesTime], 10, 64)
	return dataLastUpdate, err
}

// setLastUpdate to set redis and support config without pubsub
func (redigo *redigoWrapper) setLastUpdate(key string, data lastUpdate, ttl int) error {
	dataRedis := map[string]string{
		fieldLastChange:      data.fieldChanges,
		fieldLastChangesTime: strconv.FormatInt(data.timeUnixLastChanges, 10),
	}

	conn := redigo.pool.Get()
	defer conn.Close()

	_, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(dataRedis)...)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXPIRE", key, ttl)
	return err
}

// publishLastUpdate to publish message which field is updating
func (redigo *redigoWrapper) publishLastUpdate(key, fieldChanges string) error {
	conn := redigo.pool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", key, fieldChanges)
	return err
}
