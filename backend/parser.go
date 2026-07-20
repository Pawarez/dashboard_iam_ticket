package main

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ParseFile — entry point สำหรับ HTTP upload (multipart)
func ParseFile(fh *multipart.FileHeader) ([]Ticket, int, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return parse(f, fh.Filename)
}

// ParseFileFromPath — entry point สำหรับ test/debug จาก local disk
func ParseFileFromPath(path string) ([]Ticket, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return parse(f, filepath.Base(path))
}

// parse — core logic กลาง ไม่รู้จัก multipart หรือ os.File เลย รู้แค่ io.Reader
func parse(r io.Reader, filename string) ([]Ticket, int, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".xlsx":
		return parseExcel(r, filename)
	default:
		return nil, 0, fmt.Errorf("invalid file type: %s (only .xlsx supported)", ext)
	}
}

func parseExcel(r io.Reader, filename string) ([]Ticket, int, error) {
	xf, err := excelize.OpenReader(r)
	if err != nil {
		return nil, 0, err
	}
	defer xf.Close()

	sheetName := xf.GetSheetName(0)
	rows, err := xf.GetRows(sheetName)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) < 2 {
		return nil, 0, fmt.Errorf("file has no data rows")
	}

	idx := normalizeHeader(rows[0])
	required := []string{"ticket id", "ticket type", "priority", "created date"}
	for _, r := range required {
		if _, ok := idx[r]; !ok {
			return nil, 0, fmt.Errorf("missing required column: %s", r)
		}
	}

	var tickets []Ticket
	skipped := 0

	for _, row := range rows[1:] {
		if allEmpty(row) {
			continue
		}

		createdDate, err := parseDate(get(row, idx, "created date"))
		if err != nil || createdDate.IsZero() {
			skipped++
			continue
		}

		ticket := Ticket{
			TicketID:          get(row, idx, "ticket id"),
			TicketType:        get(row, idx, "ticket type"),
			Subject:           get(row, idx, "subject"),
			Site:              stringPtrIfValid(get(row, idx, "site")),
			SiteGroup:         stringPtrIfValid(get(row, idx, "site group")),
			RegionSite:        stringPtrIfValid(get(row, idx, "region site")),
			DomainGroup:       get(row, idx, "domain group"),
			Company:           get(row, idx, "company"),
			Country:           get(row, idx, "country"),
			Priority:          get(row, idx, "priority"),
			TicketStatus:      get(row, idx, "ticket status"),
			CustomerName:      get(row, idx, "customer name"),
			GroupAssignee:     stringPtrIfValid(get(row, idx, "group assignee")),
			Assignee:          stringPtrIfValid(get(row, idx, "assignee")),
			ProductT1:         stringPtrIfValid(get(row, idx, "product categorization tier 1")),
			ProductT2:         stringPtrIfValid(get(row, idx, "product categorization tier 2")),
			ProductT3:         stringPtrIfValid(get(row, idx, "product categorization tier 3")),
			CreatedDate:       createdDate,
			DetailDescription: get(row, idx, "detailed description"),
			SourceFile:        filename,
		}

		if resolvedAt, err := parseDate(get(row, idx, "resolved date")); err == nil && !resolvedAt.IsZero() {
			ticket.ResolvedAt = &resolvedAt
		}
		if closedDate, err := parseDate(get(row, idx, "closed date")); err == nil && !closedDate.IsZero() {
			ticket.ClosedDate = &closedDate
		}
		if completeTime, err := parseDate(get(row, idx, "completed date")); err == nil && !completeTime.IsZero() {
			ticket.CompleteTime = &completeTime
		}

		tickets = append(tickets, ticket)
	}

	return tickets, skipped, nil
}

func stringPtrIfValid(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return nil
	}
	return &value
}

func parseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return time.Time{}, nil
	}
	loc := time.FixedZone("Asia/Bangkok", 7*60*60)
	return time.ParseInLocation("02/01/2006 3:04 PM", value, loc)
}

func normalizeHeader(row []string) map[string]int {
	idx := map[string]int{}
	for i, v := range row {
		idx[strings.ToLower(strings.TrimSpace(v))] = i
	}
	return idx
}

func get(row []string, idx map[string]int, col string) string {
	i, ok := idx[col]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func allEmpty(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
