package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	db, err := gorm.Open(sqlite.Open("tickets.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&Ticket{}, &Incident{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	r.POST("/upload", uploadHandler(db))
	r.GET("/months", monthsHandler(db))
	r.GET("/tickets", ticketsHandler(db))
	r.GET("/incidents", incidentsHandler(db))
	r.POST("/incidents", saveIncidentHandler(db))
	r.GET("/detected-incidents", detectedIncidentsHandler(db))

	log.Println("listening on :8081")
	r.Run(":8081")
}