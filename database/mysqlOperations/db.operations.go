package connections

import (
	"context"
	"database/sql"
	"encoding/json"
	"excel-file-upload/config"
	redis_operation "excel-file-upload/database/redis"
	model "excel-file-upload/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db = DbConnection()

var rdb = redis_operation.Rdb
var ctx = context.Background()

func ConnectMySQL() {
	var err error
	// var db = DbConnection()
	dsn := config.LoadConfig().MySQLDSN
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v\n", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping MySQL: %v\n", err)
	}
	log.Println("Connected to MySQL!")
	CreateTable()
}
func DbConnection() *sql.DB {
	var err error
	dsn := config.LoadConfig().MySQLDSN
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v\n", err)
	}

	return db

}

func CreateTable() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS records (
		id INT AUTO_INCREMENT,
		first_name VARCHAR(255) NOT NULL,
		last_name VARCHAR(255) NOT NULL,
		company_name VARCHAR(255) NOT NULL,
		address VARCHAR(255) NOT NULL,
		city VARCHAR(255) NOT NULL,
		county VARCHAR(255) NOT NULL,
		postal VARCHAR(255) NOT NULL,
		phone VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		web VARCHAR(255) NOT NULL,
		PRIMARY KEY (id)
	)`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	} else {
		log.Println("Table created successfully!")
	}
}

func StoreToMySQL(records []model.Record) {
	for _, record := range records {
		_, err := db.Exec("INSERT INTO records (first_name, last_name, company_name, address, city, county, postal, phone, email, web) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			record.FirstName, record.LastName, record.CompanyName, record.Address, record.City, record.County, record.Postal, record.Phone, record.Email, record.Web)
		if err != nil {
			log.Printf("Failed to insert record into MySQL: %v", err)
		}
	}
}

func CacheRecords(records []model.Record) {
	data, err := json.Marshal(records)
	if err != nil {
		log.Fatalf("Failed to marshal records: %v", err)
	}
	err = rdb.Set(ctx, "records", data, 5*time.Minute).Err()
	if err != nil {
		log.Fatalf("Failed to cache records in Redis: %v", err)
	}
}

func UpdateRecord(c *gin.Context) {
	id := c.Param("id")
	var record model.Record
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("UPDATE records SET first_name = ?, last_name = ?, company_name = ?, address = ?, city = ?, county = ?, postal = ?, phone = ?, email = ?, web = ? WHERE id = ?",
		record.FirstName, record.LastName, record.CompanyName, record.Address, record.City, record.County, record.Postal, record.Phone, record.Email, record.Web, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record in MySQL"})
		return
	}

	// Update cache
	data, err := rdb.Get(ctx, "records").Result()
	if err == nil {
		var records []model.Record
		err := json.Unmarshal([]byte(data), &records)
		if err == nil {
			for i, r := range records {
				if strconv.Itoa(r.Id) == id {
					records[i] = record
					break
				}
			}
			CacheRecords(records)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Record updated successfully"})
}

func DeleteRecord(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM records WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record from MySQL"})
		return
	}

	// Update cache
	data, err := rdb.Get(ctx, "records").Result()
	if err == nil {
		var records []model.Record
		err := json.Unmarshal([]byte(data), &records)
		if err == nil {
			for i, r := range records {
				if strconv.Itoa(r.Id) == id {
					records = append(records[:i], records[i+1:]...)
					break
				}
			}
			CacheRecords(records)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Record deleted successfully"})
}
