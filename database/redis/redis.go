package redis_operation

import (
	"context"
	"encoding/json"
	"excel-file-upload/config"
	model "excel-file-upload/models"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func ConnectRediss() *redis.Client {
	addr := config.LoadConfig().RedisAddr
	password := config.LoadConfig().RedisPassword
	db := config.LoadConfig().RedisDB
	ctx := context.Background()
	// Ensure that you have Redis running on your system
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	status, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalln("Redis connection was refused")
	}
	fmt.Println("Redis connection status : ", status)
	return rdb
}
func CacheRecords(records []model.Record) {
	var ctx = context.Background()
	rdb := ConnectRediss()
	defer rdb.Close()
	// fmt.Printf("%+v\n", records)
	data, err := json.Marshal(records)
	if err != nil {
		log.Fatalf("Failed to marshal records: %v", err)
	}
	err = rdb.Set(ctx, "records", data, 0).Err()
	if err != nil {
		log.Fatalf("Failed to cache records in Redis: %v", err)
	}
}
