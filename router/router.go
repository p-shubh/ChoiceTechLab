package router

import (
	"excel-file-upload/handler"

	"github.com/gin-gonic/gin"
)

func Router() {
	router := gin.Default()

	router.POST("/upload", handler.UploadExcel)
	router.GET("/records", handler.GetRecords)
	router.PUT("/records/:id", handler.UpdateRecord)
	router.DELETE("/records/:id", handler.DeleteRecord)

	router.Run(":8080")
}
