package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func uploadHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fh, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided (field 'file')"})
			return
		}

		tickets, skipped, err := ParseFile(fh)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Ticket{}).Error; err != nil {
				return err
			}

			if len(tickets) > 0 {
				if err := tx.CreateInBatches(&tickets, 200).Error; err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"inserted": len(tickets),
			"skipped":  skipped,
			"file":     fh.Filename,
		})
	}
}

func monthsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var months []string
		err := db.Raw(`
			SELECT DISTINCT strftime('%Y-%m', created_date, '+7 hours') AS month
			FROM tickets
			ORDER BY month DESC
		`).Scan(&months).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, months)
	}
}

func ticketsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		month := c.Query("month")
		var tickets []Ticket
		var err error

		if month != "" {
			// เจาะจงเดือนที่ขอ
			err = db.Where("strftime('%Y-%m', created_date, '+7 hours') = ?", month).
				Order("created_date DESC").
				Find(&tickets).Error
		} else {
			// default: 3 เดือนล่าสุด นับจากวันนี้
			cutoff := time.Now().AddDate(0, -3, 0)
			err = db.Where("created_date >= ?", cutoff).
				Order("created_date DESC").
				Find(&tickets).Error
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tickets)
	}
}

func incidentsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var incidents []Incident
		err := db.Order("date DESC").Find(&incidents).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, incidents)
	}
}

func saveIncidentHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Date        string `json:"date" binding:"required"`
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			RootCause   string `json:"root_cause"`
			TicketCount int    `json:"ticket_count"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var incident Incident
		err := db.Where("date = ?", input.Date).First(&incident).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		incident.Date = input.Date
		incident.Title = input.Title
		incident.Description = input.Description
		incident.RootCause = input.RootCause
		incident.TicketCount = input.TicketCount

		if err := db.Save(&incident).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, incident)
	}
}

func detectedIncidentsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type Result struct {
			Date        string `json:"date"`
			TicketCount int    `json:"ticket_count"`
		}
		var results []Result
		err := db.Raw(`
			SELECT strftime('%Y-%m-%d', created_date, '+7 hours') AS date, COUNT(*) AS ticket_count
			FROM tickets
			GROUP BY date
			HAVING ticket_count >= 50
			ORDER BY date DESC
		`).Scan(&results).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, results)
	}
}
