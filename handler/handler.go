package handler

import (
	"context"
	"encoding/json"
	connections "excel-file-upload/database/mysqlOperations"
	redis_operation "excel-file-upload/database/redis"
	model "excel-file-upload/models"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/xuri/excelize/v2"
)

var ctx = context.Background()

func UploadExcel(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}

	// Save the file locally
	if err := c.SaveUploadedFile(file, "save.xlsx"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
		return
	}

	records := parseExcel()
	data := connections.StoreToMySQL(records)
	redis_operation.CacheRecords(data)

	c.JSON(http.StatusOK, gin.H{"status": "File processed successfully"})
}

func parseExcel() []model.Record {

	f, err := excelize.OpenFile("save.xlsx")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Get value from cell by given worksheet name and cell reference.
	list := f.GetSheetList()
	cell, err := f.GetCellValue(list[0], "B2")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(cell)
	// Get all the rows in the Sheet1.
	rows, err := f.GetRows(list[0])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var employees []model.Record
	for i, row := range rows {
		if i == 0 {
			// Skip the header row
			continue
		}
		if len(row) < 10 {
			continue // Skip incomplete rows
		}
		employees = append(employees, model.Record{
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
		})
	}
	// Delete the file after processing.
	if err := os.Remove("save.xlsx"); err != nil {
		fmt.Println("Error deleting file:", err)
	}
	return employees
}
func GetRecords(c *gin.Context) {
	var rdb = redis_operation.ConnectRediss()
	var db = connections.DbConnection()
	defer db.Close()
	defer rdb.Close()
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
		redis_operation.CacheRecords(records)
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
	var rdb = redis_operation.ConnectRediss()
	var db = connections.DbConnection()
	defer db.Close()
	defer rdb.Close()
	id := c.Param("id")
	var record model.Record
	record.Id, _ = strconv.Atoi(id)
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("UPDATE records SET id = ?, first_name = ?, last_name = ?, company_name = ?, address = ?, city = ?, county = ?, postal = ?, phone = ?, email = ?, web = ? WHERE id = ?", id,
		record.FirstName, record.LastName, record.CompanyName, record.Address, record.City, record.County, record.Postal, record.Phone, record.Email, record.Web, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record in MySQL", "status": err.Error()})
		return
	}

	// Update cache
	// data, err := rdb.Get(ctx, "records").Result()
	// if err == nil {
	// 	var records []model.Record
	// 	err := json.Unmarshal([]byte(data), &records)
	// 	if err == nil {
	// 		for i, r := range records {
	// 			if strconv.Itoa(r.Id) == id {
	// 				records[i] = record
	// 				break
	// 			}
	// 		}
	// 		connections.CacheRecords(records)
	// 	}
	// }

	// Update cache
	data, err := rdb.Get(ctx, "records").Result()
	if err == nil {
		var records []model.Record
		err := json.Unmarshal([]byte(data), &records)
		fmt.Printf("%+v\n", record)
		if err == nil {
			for i, r := range records {
				if strconv.Itoa(r.Id) == id {
					records[i] = record
					break
				}
			}
			redis_operation.CacheRecords(records)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Record updated successfully"})
}

func DeleteRecord(c *gin.Context) {
	var rdb = redis_operation.ConnectRediss()
	var db = connections.DbConnection()
	defer db.Close()
	defer rdb.Close()
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
			redis_operation.CacheRecords(records)
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Record deleted successfully"})
}
