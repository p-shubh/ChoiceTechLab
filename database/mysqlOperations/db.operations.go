package connections

import (
	"database/sql"
	"excel-file-upload/config"
	model "excel-file-upload/models"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectMySQL() {
	var err error
	var db = DbConnection()
	defer db.Close()
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
	var db = DbConnection()
	defer db.Close()
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

func StoreToMySQL(records []model.Record) []model.Record {
	var db = DbConnection()
	defer db.Close()
	for _, record := range records {
		_, err := db.Exec("INSERT INTO records (first_name, last_name, company_name, address, city, county, postal, phone, email, web) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			record.FirstName, record.LastName, record.CompanyName, record.Address, record.City, record.County, record.Postal, record.Phone, record.Email, record.Web)
		if err != nil {
			log.Printf("Failed to insert record into MySQL: %v", err)
		}
	}
	rows, err := db.Query("SELECT id, first_name, last_name, company_name, address, city, county, postal, phone, email, web FROM records")
	if err != nil {
		fmt.Println("error", "Failed to fetch records")
		return nil
	}
	defer rows.Close()
	var recordss []model.Record
	for rows.Next() {
		var record model.Record
		err := rows.Scan(&record.Id, &record.FirstName, &record.LastName, &record.CompanyName, &record.Address, &record.City, &record.County, &record.Postal, &record.Phone, &record.Email, &record.Web)
		if err != nil {
			fmt.Println("error", "Failed to scan record")
			return nil
		}
		recordss = append(recordss, record)
	}
	return recordss
}
