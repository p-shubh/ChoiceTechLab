package main

import (
	"excel-file-upload/config"
	connections "excel-file-upload/database/mysqlOperations"
	redis_operation "excel-file-upload/database/redis"
	"excel-file-upload/router"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if len(os.Getenv("MYSQL_USER")) == 0 {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file", err.Error())
		}
	}
}
func init() {
	fmt.Printf("%+v\n", *config.LoadConfig())
	func() {
		connections.ConnectMySQL()
		redis_operation.ConnectRedis(config.LoadConfig().RedisAddr, config.LoadConfig().RedisPassword, config.LoadConfig().RedisDB)
	}()
}
func main() {
	router.Router()
}
