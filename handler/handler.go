package handler

import (
	"context"
	"encoding/json"
	connections "excel-file-upload/database/mysqlOperations"
	redis_operation "excel-file-upload/database/redis"
	model "excel-file-upload/models"
	"log"
	"net/http"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var db = connections.DbConnection()

var rdb = redis_operation.Rdb
var ctx = context.Background()

func UploadExcel(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}
	defer file.Close()

	xlFile, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
		return
	}

	records := parseExcel(xlFile)
	connections.StoreToMySQL(records)
	connections.CacheRecords(records)

	c.JSON(http.StatusOK, gin.H{"status": "File processed successfully"})
}

func parseExcel(xlFile *excelize.File) []model.Record {
	rows := xlFile.GetRows("Sheet1")
	// if err != nil {
	// 	log.Fatalf("Failed to get rows from Excel: %v", err)
	// }

	var records []model.Record
	for _, row := range rows[1:] {
		if len(row) < 10 {
			continue
		}
		record := model.Record{
			FirstName:   row[0],
			LastName:    row[1],
			CompanyName: row[2],
			Address:     row[3],
			City:        row[4],
			County:      row[5],
			Postal:      row[6],
			Phone:       row[7],
			Email:       row[8],
			Web:         row[9],
		}
		records = append(records, record)
	}
	return records
}
func GetRecords(c *gin.Context) {
	data, err := rdb.Get(ctx, "records").Result()
	if err == redis.Nil {
		log.Println("Data not found in Redis, fetching from MySQL")
		rows, err := db.Query("SELECT id, first_name, last_name, company_name, address, city, county, postal, phone, email, web FROM records")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
			return
		}
		defer rows.Close()
		var records []model.Record
		for rows.Next() {
			var record model.Record
			err := rows.Scan(&record.Id, &record.FirstName, &record.LastName, &record.CompanyName, &record.Address, &record.City, &record.County, &record.Postal, &record.Phone, &record.Email, &record.Web)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan record"})
				return
			}
			records = append(records, record)
		}
		connections.CacheRecords(records)
		c.JSON(http.StatusOK, records)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records from Redis"})
	} else {
		var records []model.Record
		err = json.Unmarshal([]byte(data), &records)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal records"})
		} else {
			c.JSON(http.StatusOK, records)
		}
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
			connections.CacheRecords(records)
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
			connections.CacheRecords(records)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Record deleted successfully"})
}
